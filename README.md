WIP CapNProto filesystem schema + example sandstorm apps

## Why

File(system) access is a common enough need that many Sandstorm apps end
up bundling their own file manager UI, and there's a lot of duplication.
It would be better if there was a standard schema that apps could use to
allow one another access to files.

## What

This repo contains a work-in-progress schema that can be requested &
offered via the powerbox, plus some example apps. See [this mailing list
post][1] for more information.

## License

Apache 2.0

## Troubleshooting

The Makefile is a fairly simple wrapper around Go's default build
system. For sandstorm folks less familiar with Go, here are a few common
errors and how to get around them:

  - symptoms: `make` responds with `cannot find package "github.com/gorilla/mux"`
    - treatment: `go get`
  - symptoms: `package zenhack.net/go/sandstorm-filesystem/filesystem: unrecognized import path` ...
    - treatment: make sure your source lives at `$GOPATH/src/zenhack.net/go/sandstorm-filesystem`

[1]: https://groups.google.com/forum/#!searchin/sandstorm-dev/sandstorm-filesystem%7Csort:relevance/sandstorm-dev/sjEldWXrAjc/zjNGGdMpCAAJ
