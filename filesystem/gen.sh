#!/usr/bin/env sh

export GOPATH=${GOPATH:-$HOME/go}
cd $(dirname $0)
capnp compile *.capnp -ogo \
	-I $GOPATH/src/zombiezen.com/go/capnproto2/std \
	-I $GOPATH/src/zenhack.net/go/sandstorm/capnp
