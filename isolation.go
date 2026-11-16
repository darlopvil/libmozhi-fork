package libmozhi

import (
	"errors"
	"time"
)

// ============================================================================
// ENGINE ISOLATION — used by TranslateAll and TranslateSome.
// Every engine runs in its own goroutine. A panic or a hang in one engine must
// never take the process down or hold the other engines' results hostage
// (Translate turns panics into errors).
// ============================================================================

// engineDeadline caps how long the multi-engine modes wait for any engine.
// It sits above the worst case of the engines' own HTTP timeouts (3 attempts
// x 15s + backoff) and below nginx's default 60s proxy_read_timeout, so the
// page is always answered before the reverse proxy gives up on it.
const engineDeadline = 50 * time.Second

// errEnginePanic is what Translate returns when an engine panicked.
var errEnginePanic = errors.New("engine panicked")

type engineResult struct {
	index int
	out   LangOut
	err   error
}

// runEngines translates with every engine in parallel and returns the
// successful results in the same order as the engines slice. Engines that
// fail, panic or miss engineDeadline are simply left out.
func runEngines(engines []string, to, from, query string) []LangOut {
	// Buffered so an engine that answers after the deadline never blocks.
	results := make(chan engineResult, len(engines))
	for i := range engines {
		go func(i int) {
			out, err := Translate(engines[i], to, from, query)
			results <- engineResult{index: i, out: out, err: err}
		}(i)
	}

	ok := make([]bool, len(engines))
	outs := make([]LangOut, len(engines))
	timer := time.NewTimer(engineDeadline)
	defer timer.Stop()

collect:
	for pending := len(engines); pending > 0; pending-- {
		select {
		case r := <-results:
			if r.err == nil {
				ok[r.index] = true
				outs[r.index] = r.out
			}
		case <-timer.C:
			break collect
		}
	}

	langout := []LangOut{}
	for i := range engines {
		if ok[i] {
			langout = append(langout, outs[i])
		}
	}
	return langout
}
