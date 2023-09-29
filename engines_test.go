package libmozhi_test

import (
	"testing"

	"codeberg.org/aryak/libmozhi"
)

func TestAllEngines(t *testing.T) {
	results := libmozhi.TranslateAll("de", "en", "Hallo Welt!")
	if len(results) <= 1 {
		t.Errorf("Only %d engines returned a result!", len(results))
	}
}
