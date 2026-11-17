package libmozhi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"strings"
	"time"
	"unicode/utf8"
)

// ============================================================================
// TTS STREAMING — long texts in chunks, audio sent as soon as it exists.
// Every TTS engine caps the text per request (measured nov 2026: Google 200
// characters, TexTra 200+ with ~10s for 240 Japanese characters, Groq 600).
// The text is split at sentence boundaries; the first chunk is synthesized
// before answering, and each following chunk is fetched while the previous
// one plays (200 characters are ~12s of speech, generated in 0.3-8s).
// ============================================================================

// ttsChunkLimit is the measured maximum text per request of each engine.
// Engines not listed get the whole text in one request.
var ttsChunkLimit = map[string]int{
	"google": 200,
	"textra": 120, // ~0.04-0.07s per Japanese character: 200 took 14.5s, too close to the 15s timeout
	"groq":   600,
	// Azure Speech F0 allows 20 requests a minute: the longest text Mozhi
	// accepts (5000 characters by default) takes 5.
	"microsoft": 1000,
}

// ttsRetryWaits paces retries of a failed chunk after the first one, while
// earlier audio is still playing (Groq: 1200 characters per minute).
var ttsRetryWaits = []time.Duration{5 * time.Second, 10 * time.Second, 20 * time.Second}

// splitForTTS splits text into chunks of at most max characters, preferring
// sentence ends, then clause marks, then spaces.
func splitForTTS(text string, max int) []string {
	text = strings.TrimSpace(text)
	if max <= 0 || utf8.RuneCountInString(text) <= max {
		return []string{text}
	}
	var chunks []string
	var cur strings.Builder
	flush := func() {
		if s := strings.TrimSpace(cur.String()); s != "" {
			chunks = append(chunks, s)
		}
		cur.Reset()
	}
	for _, piece := range splitAfterAny(text, ".!?。！？\n") {
		if utf8.RuneCountInString(cur.String())+utf8.RuneCountInString(piece) <= max {
			cur.WriteString(piece)
			continue
		}
		flush()
		if utf8.RuneCountInString(piece) <= max {
			cur.WriteString(piece)
			continue
		}
		// A single sentence longer than max: clauses, then words, then a hard cut.
		for _, part := range splitAfterAny(piece, ",;:、，；：") {
			for _, word := range splitAfterAny(part, " ") {
				for utf8.RuneCountInString(word) > max {
					r := []rune(word)
					flush()
					chunks = append(chunks, string(r[:max]))
					word = string(r[max:])
				}
				if utf8.RuneCountInString(cur.String())+utf8.RuneCountInString(word) > max {
					flush()
				}
				cur.WriteString(word)
			}
		}
	}
	flush()
	return chunks
}

// splitAfterAny cuts s after every rune found in marks, keeping the marks.
func splitAfterAny(s, marks string) []string {
	var out []string
	start := 0
	for i, r := range s {
		if strings.ContainsRune(marks, r) {
			end := i + utf8.RuneLen(r)
			out = append(out, s[start:end])
			start = end
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

// parseWAV returns the fmt chunk and the PCM samples of a RIFF/WAVE file.
// Sizes of 0xFFFFFFFF (streamed WAV, as Groq sends) mean "until the end".
func parseWAV(b []byte) (fmtChunk, pcm []byte, err error) {
	if len(b) < 12 || string(b[0:4]) != "RIFF" || string(b[8:12]) != "WAVE" {
		return nil, nil, errors.New("tts: not a WAV file")
	}
	for p := 12; p+8 <= len(b); {
		id := string(b[p : p+4])
		size := int(binary.LittleEndian.Uint32(b[p+4 : p+8]))
		body := p + 8
		if size < 0 || body+size > len(b) || uint32(size) == 0xFFFFFFFF {
			size = len(b) - body
		}
		switch id {
		case "fmt ":
			fmtChunk = b[p : body+size]
		case "data":
			if fmtChunk == nil {
				return nil, nil, errors.New("tts: WAV data before fmt")
			}
			return fmtChunk, b[body : body+size], nil
		}
		p = body + size + size%2
	}
	return nil, nil, errors.New("tts: WAV without data")
}

// wavFile builds a RIFF/WAVE file from a fmt chunk and PCM samples, with
// the given RIFF and data sizes (0xFFFFFFFF: unknown, for streaming).
func wavFile(fmtChunk, pcm []byte, riffSize, dataSize uint32) []byte {
	var h bytes.Buffer
	h.WriteString("RIFF")
	_ = binary.Write(&h, binary.LittleEndian, riffSize)
	h.WriteString("WAVE")
	h.Write(fmtChunk)
	h.WriteString("data")
	_ = binary.Write(&h, binary.LittleEndian, dataSize)
	h.Write(pcm)
	return h.Bytes()
}

// TTSStream synthesizes text with engine, chunk by chunk. The first chunk is
// synthesized before it returns, so a failure there is an ordinary error and
// the audio type is known; write then sends that chunk and every following
// one as soon as each is ready, flushing w if it can be flushed.
func TTSStream(engine, lang, text string) (contentType string, write func(w io.Writer) error, err error) {
	chunks := splitForTTS(text, ttsChunkLimit[engine])

	// TexTra: reserve every slot the text needs up front, or let Google read
	// all of it. Formats can't be mixed in one stream (WAV vs MP3).
	if engine == "textra" {
		if _, ok := textraTTSLocales[lang]; !ok {
			return "", nil, errors.New("textra tts: unsupported language")
		}
		if textraTake(len(chunks), 0) != nil {
			engine = "google"
			if lang == "zh" {
				lang = "zh-CN"
			}
			chunks = splitForTTS(text, ttsChunkLimit[engine])
		}
	}
	synth := func(chunk string) ([]byte, error) {
		if engine == "textra" {
			start := time.Now()
			audio, err := ttsTexTraRaw(lang, chunk) // slots already reserved
			recordHealth(engine, "tts", start, err)
			return audio, err
		}
		return TTS(engine, lang, chunk)
	}

	first, err := synth(chunks[0])
	if err != nil {
		return "", nil, err
	}
	isWAV := len(first) >= 4 && string(first[0:4]) == "RIFF"
	var wavFmt []byte
	if isWAV {
		contentType = "audio/wave"
	} else {
		contentType = "audio/mpeg"
	}

	write = func(w io.Writer) error {
		emit := func(audio []byte, index int) error {
			out := audio
			if isWAV {
				f, pcm, perr := parseWAV(audio)
				if perr != nil {
					return perr
				}
				if len(chunks) == 1 {
					// One chunk: its real sizes, so players show the real
					// duration (Groq itself sends 0xFFFFFFFF).
					out = wavFile(f, pcm, uint32(4+len(f)+8+len(pcm)), uint32(len(pcm)))
				} else if index == 0 {
					// Streamed WAV header: unknown total size (0xFFFFFFFF).
					wavFmt = f
					out = wavFile(f, pcm, 0xFFFFFFFF, 0xFFFFFFFF)
				} else if !bytes.Equal(f, wavFmt) {
					return errors.New("tts: audio format changed between chunks")
				} else {
					out = pcm
				}
			}
			if _, werr := w.Write(out); werr != nil {
				return werr
			}
			if f, ok := w.(interface{ Flush() error }); ok {
				return f.Flush()
			}
			return nil
		}

		if err := emit(first, 0); err != nil {
			return err
		}
		for i := 1; i < len(chunks); i++ {
			audio, err := synth(chunks[i])
			for _, wait := range ttsRetryWaits {
				if err == nil || engine == "textra" {
					break
				}
				time.Sleep(wait)
				audio, err = synth(chunks[i])
			}
			if err != nil {
				log.Printf("libmozhi: tts stream (%s) stopped at chunk %d/%d: %v", engine, i+1, len(chunks), err)
				return err
			}
			if err := emit(audio, i); err != nil {
				return err
			}
		}
		return nil
	}
	return contentType, write, nil
}
