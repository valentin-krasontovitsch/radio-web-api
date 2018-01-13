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

- `/connect` (GET) - attempts to connect to the speakers
- `/connected` (GET) - returns whether we are connected, or not, as
  `{"connected": true}` where the latter is a boolean
- `/volume` (GET,PUT) - gets / sets volume, returns and expects JSON of the
  form `{"volume": 34}` where the number should be between 0 and 100

## Configuration

- `PORT` specifies where to serve

## Implementation

Written in go, using [gin-gonic](https://github.com/gin-gonic/gin) for web
server framework.
