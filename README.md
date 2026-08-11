> **Fork notice:** versión modificada de [aryak/libmozhi](https://codeberg.org/aryak/libmozhi),
> mantenida por [darlopvil](https://github.com/darlopvil). Cambios y roadmap en los
> [issues](https://github.com/darlopvil/libmozhi-fork/issues). Licencia AGPLv3, igual que upstream.

# LibMozhi
[![AGPLv3](https://shields.io/badge/License-AGPL%20v3-blue.svg)](https://gnu.org/licenses/agpl-3.0.en.html)
[![Matrix](https://img.shields.io/badge/matrix-000000?style=for-the-badge&logo=Matrix&logoColor=white)](https://matrix.to/#/#mozhi:projectsegfau.lt)
[![Go Reference](https://pkg.go.dev/badge/codeberg.org/aryak/libmozhi.svg)](https://pkg.go.dev/codeberg.org/aryak/libmozhi)

Library that backs [Mozhi](https://codeberg.org/aryak/mozhi).

## File Matrix
- autodetect.go - Auto detect language from given text.
- engines.go - Code for getting data from engines themselves
- langlist-\*.go - List of languages supported by the specific engine
- tts.go - Get TTS data for given text
- retrieve.go - Used to get new langlist data
- libmozhi.go - Main functions that wrap around the others to make the thing usable
