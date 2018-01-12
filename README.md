# radio-web-api

This project exposes controls to a bluetooth radio (which at the moment are
realized through [some bash
scripts](https://github.com/valentin-krasontovitsch/blue-radio-shell)) via a
web interface.

## Prerequisites

*TODO* describe which executable files are expected to be found, and with what
functionality / behavior.

## Endpoints

- `/connect` (GET) - attempts to connect to the speakers
- `/connected` (GET) - returns whether we are connected, or not
- `/volume` (GET,PUT) - gets / sets volume *TODO*

## Configuration

- `PORT` specifies where to serve

## Implementation

Written in go, using [gin-gonic](https://github.com/gin-gonic/gin) for web
server framework.
