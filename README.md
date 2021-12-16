# radio-web-api

This project exposes controls to a bluetooth radio (which at the moment are
realized through [some bash
scripts](https://github.com/valentin-krasontovitsch/blue-radio-shell)) via a
web interface.

## Prerequisites

*TODO* describe which executable files are expected to be found, and with what
functionality / behavior.

## Endpoints

All endpoints return 200 on OK, 500 on internal errors and 400 on bad request.
We try to be verbose when something goes wrong. We expect and return JSON, and
put error messages in a field with key `error`.

For more details, start the server and query `/` or look at the function
`explainAPI` in the code.

## Configuration

- `PORT` which port the web api will listen on
- `SPEAKER_ADDRESS` MAC address of speaker
- `LOCAL_AUDIO_PATH` directory containing local audio files

## Implementation

Written in go, using [gin-gonic](https://github.com/gin-gonic/gin) for web
server framework.

## Building

To include the current git commit and possibly tag in the compiled binary as a
version string, build the package as follows:

```
go build -ldflags "-X main.version="(git describe --tags) .
```

When building for instance for a raspberry pi zero w, the correct invocation is

```
env GOOS=linux GOARCH=arm GOARM=6 \
  go build -ldflags "-X main.version="(git describe --tags) .
```
