package libmozhi

import (
	"errors"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

// ============================================================================
// OPENROUTER FREE-MODEL DISCOVERY — used by the OpenRouter islands only.
// Each island names a model *family* (e.g. "google/gemma") instead of a fixed
// model. The current free models of that family are read from OpenRouter's
// public catalogue, so a retired or renamed free model no longer breaks the
// island. Only ":free" ids are ever sent, so requests can never be billed.
// ============================================================================

const (
	// The free catalogue gains models every 2-3 days; retirements don't wait
	// for the TTL, they force a refresh as soon as OpenRouter rejects an id.
	openrouterCatalogTTL = 24 * time.Hour
	// Minimum gap between catalogue requests, even after a failure.
	openrouterRefreshGap = time.Minute
	// How long a rejected (retired or region-blocked) id is skipped.
	openrouterBanTTL = 24 * time.Hour
	// Hard limit of OpenRouter's `models` fallback array ("must have 3 items
	// or fewer"); an invalid id anywhere in it fails the whole request.
	openrouterMaxModels = 3
)

type openrouterFreeModel struct {
	id      string
	created int64
}

var openrouterCatalog struct {
	sync.Mutex
	models  []openrouterFreeModel // free models only, newest first
	fetched time.Time             // last successful fetch
	tried   time.Time             // last fetch attempt
	lastErr error
	banned  map[string]time.Time // id -> skipped until
}

var openrouterCatalogHTTP = &http.Client{Timeout: 15 * time.Second}

// openrouterFetchCatalog downloads the public model list and keeps only the
// free, non-expired models, newest first.
func openrouterFetchCatalog() ([]openrouterFreeModel, error) {
	baseURL := os.Getenv("MOZHI_OPENROUTER_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	res, err := openrouterCatalogHTTP.Get(baseURL + "/models")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 || !gjson.ValidBytes(body) {
		return nil, errors.New("openrouter models list error: " + res.Status)
	}

	now := time.Now()
	free := []openrouterFreeModel{}
	gjson.GetBytes(body, "data").ForEach(func(_, m gjson.Result) bool {
		id := m.Get("id").String()
		if !strings.HasSuffix(id, ":free") {
			return true
		}
		if exp := m.Get("expiration_date").String(); exp != "" {
			t, perr := time.Parse("2006-01-02", exp)
			if perr != nil {
				t, perr = time.Parse(time.RFC3339, exp)
			}
			if perr == nil && !now.Before(t) {
				return true // already expired
			}
		}
		free = append(free, openrouterFreeModel{id: id, created: m.Get("created").Int()})
		return true
	})
	if len(free) == 0 {
		return nil, errors.New("openrouter models list: no free models found")
	}
	sort.SliceStable(free, func(i, j int) bool { return free[i].created > free[j].created })
	return free, nil
}

// openrouterFreeModels returns up to openrouterMaxModels current free models
// of family, newest first. preferred (optional, from the island's _MODEL env)
// goes first if it is still free and not banned.
func openrouterFreeModels(family, preferred string) ([]string, error) {
	c := &openrouterCatalog
	c.Lock()
	defer c.Unlock()

	now := time.Now()
	stale := c.models == nil || now.Sub(c.fetched) > openrouterCatalogTTL
	if stale && now.Sub(c.tried) > openrouterRefreshGap {
		c.tried = now
		fresh, err := openrouterFetchCatalog()
		if err == nil {
			c.models, c.fetched, c.lastErr = fresh, now, nil
		} else {
			c.lastErr = err // keep serving the previous catalogue, if any
		}
	}
	if c.models == nil {
		if c.lastErr != nil {
			return nil, c.lastErr
		}
		return nil, errors.New("openrouter models list unavailable")
	}

	out := []string{}
	add := func(id string) {
		if len(out) >= openrouterMaxModels {
			return
		}
		if until, ok := c.banned[id]; ok && now.Before(until) {
			return
		}
		for _, o := range out {
			if o == id {
				return
			}
		}
		out = append(out, id)
	}
	for _, m := range c.models {
		if m.id == preferred {
			add(m.id)
		}
	}
	for _, m := range c.models {
		if strings.HasPrefix(m.id, family) {
			add(m.id)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("openrouter: no free " + family + " model available")
	}
	return out, nil
}

// openrouterModelRejected reports whether OpenRouter refused the request
// because of one of the models themselves (retired/renamed id, or blocked in
// this region). If so it bans that id and, for retired ids, forces a fresh
// catalogue on the next call, so the island's retry uses the current models.
func openrouterModelRejected(models []string, status int, body []byte) bool {
	msg := gjson.GetBytes(body, "error.message").String()
	retired := status == 404 || strings.Contains(msg, "not a valid model ID")
	blocked := strings.Contains(msg, "not available in your region")
	if len(models) == 0 || (!retired && !blocked) {
		return false
	}

	bad := models[0]
	for _, id := range models {
		if strings.Contains(msg, id) {
			bad = id
			break
		}
	}

	c := &openrouterCatalog
	c.Lock()
	defer c.Unlock()
	if c.banned == nil {
		c.banned = map[string]time.Time{}
	}
	c.banned[bad] = time.Now().Add(openrouterBanTTL)
	if retired {
		c.fetched = time.Time{}
	}
	return true
}

// openrouterErrMsg extracts OpenRouter's error message (without the rest of
// the body, which carries account ids) to make API errors diagnosable.
func openrouterErrMsg(body []byte) string {
	if msg := gjson.GetBytes(body, "error.message").String(); msg != "" {
		return ": " + msg
	}
	return ""
}
