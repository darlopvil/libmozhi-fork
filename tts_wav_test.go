package libmozhi

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestWavFileSizes(t *testing.T) {
	fmtChunk := append([]byte("fmt \x10\x00\x00\x00"), make([]byte, 16)...)
	pcm := []byte{1, 2, 3, 4, 5, 6}
	streamed := wavFile(fmtChunk, pcm, 0xFFFFFFFF, 0xFFFFFFFF)
	f, p, err := parseWAV(streamed)
	if err != nil || !bytes.Equal(f, fmtChunk) || !bytes.Equal(p, pcm) {
		t.Fatalf("streamed WAV doesn't parse back: %v", err)
	}
	exact := wavFile(f, p, uint32(4+len(f)+8+len(p)), uint32(len(p)))
	if got := binary.LittleEndian.Uint32(exact[4:8]); int(got) != len(exact)-8 {
		t.Errorf("RIFF size %d, want %d", got, len(exact)-8)
	}
	if got := binary.LittleEndian.Uint32(exact[len(exact)-len(pcm)-4:]); int(got) != len(pcm) {
		t.Errorf("data size %d, want %d", got, len(pcm))
	}
}

func TestTTSSupports(t *testing.T) {
	for _, c := range []struct {
		engine, lang string
		want         bool
	}{
		{"textra", "ja", true}, {"textra", "es", false}, {"groq", "en", true}, {"groq", "es", false},
		{"google", "es", true}, {"microsoft", "es", true}, {"edge", "ja", true}, {"deepl", "en", false},
	} {
		if got := TTSSupports(c.engine, c.lang); got != c.want {
			t.Errorf("TTSSupports(%s, %s) = %v, want %v", c.engine, c.lang, got, c.want)
		}
	}
}
