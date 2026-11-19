# libmozhi — personal fork

Fork of [aryak/libmozhi](https://codeberg.org/aryak/libmozhi), the engine library behind
[Mozhi](https://codeberg.org/aryak/mozhi). It backs my own Mozhi fork,
[darlopvil/mozhi-fork](https://github.com/darlopvil/mozhi-fork): its README lists what changed,
and issues for both repos are welcome there.

License: AGPL-3.0, same as upstream ([LICENSE](LICENSE)).

## Files

Each engine is an island: its own file and its own language list, so adding or removing one
doesn't touch the others.

- `libmozhi.go`: the entry points (`Translate`, `TTS`, `LangList`, `TTSSupports`) and their dispatchers.
- `engines.go`: the upstream engines (Google, DeepL, Yandex, MyMemory, plus Reverso and LibreTranslate, which Mozhi no longer offers) and `TranslateAll` / `TranslateSome`.
- `duckduckgo.go`, `gemini.go`, `textra.go`, `groq.go`, `qwen.go`, `cohere.go`, `microsoft.go`, `langbly.go`, `papago.go`, `volcengine.go`: one engine each, with its `<engine>_langs.go`.
- `langlist-*.go`, `langlists.go`: language lists of the upstream engines. `langs_resolve.go` matches a language across the codes each engine uses (`es` / `es-ES`, `zh-CN` / `zh-Hans`).
- `dict.go`: Google's dictionary. `yandex_translit.go`: Yandex transliteration.
- `jadict.go` and `jadict_<source>.go`: the Japanese dictionary (Jisho, jpdb, MOJi, Weblio).
- `tts.go`, `tts_stream.go`: text-to-speech, and long texts read in streamed chunks. `microsoft_tts.go`: Azure Speech. `edge_tts.go` and `edge_drm.go`: Edge's read-aloud voices, Azure's backup, and their request token.
- `health.go`: the health of every engine, dictionary and voice, from real requests, with short re-checks.
- `isolation.go`: per-engine timeouts and panic recovery, so one engine can't stall All mode.
- `images.go`: image translation (Google).
- `requests.go`: HTTP helpers.
- `cmds/retrieve.go`: regenerates the upstream language lists.

Tests: `go test` (`TestAllEngines` needs network access).
