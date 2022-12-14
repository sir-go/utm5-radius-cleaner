# UTM5 RADIUS sessions cleaner
[![Go](https://github.com/sir-go/utm5-radius-cleaner/actions/workflows/go.yml/badge.svg)](https://github.com/sir-go/utm5-radius-cleaner/actions/workflows/go.yml)

A tool for cleanup stuck RADIUS sessions in UTM5 billing kernel and Cisco ISG Router.

## How it works
1. get active sessions list `rpcf_radius_get_active_sessions`
2. kill sessions with `traf_last_update_date` > `%session_lifetime%` (> 10 min) - `rpcf_radius_drop_session`
3. if failed then restart `utm_radius` daemon and retry
4. if `IP == 0.0.0.0` then kill session on the ISG-Router via telnet `clear sss session user <ip>`

## Flags
`-c <config file path>` - path to `*.toml` config file

example config - `config.toml`

## Docker
```bash
docker build -t r_cleaner .
docker run --rm -it -v ${PWD}/config.toml:/config.toml:ro r_cleaner:latest
```

## Build
```bash
go mod download
gosec -exclude G106 ./... && build -o r_cleaner cmd/cleaner
```
