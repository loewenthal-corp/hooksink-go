# hooksink-go - agent guide

This repository is a Go library for receiving Slack-compatible incoming webhook
payloads. It is parser-first: `ParseBody` is the primitive, `Parse` adapts
`*http.Request`, and `Handler` is optional net/http sugar.

## Dev Loop

Activate the pinned toolchain first:

```sh
source bin/activate-hermit
```

Use Taskfile commands as the front door:

```sh
task do          # full local quality gate
task test        # unit tests
task test::ci    # race + coverage
task lint        # Go and workflow lint
task security    # govulncheck, gitleaks, zizmor
```

Do not add router dependencies to the root module. Router demos live in isolated
modules under `examples/`.

## Compatibility

The module declares Go 1.22. Hermit pins the current Go toolchain for secure
stdlib and vulnerability scanning, while `task test::compat` runs the root and
example modules with `GOTOOLCHAIN=go1.22.12` to verify the supported floor. Keep
dependencies compatible with that floor.

## Release

Release automation is release-please manifest mode. Conventional commits on
`main` drive changelog and GitHub Release creation.
