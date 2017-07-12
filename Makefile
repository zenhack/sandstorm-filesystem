
GOPATH ?= $(HOME)/go

all: app.stripped app.debug
clean:
	rm -f app app.stripped app.debug
dev: all
	spk dev -p .sandstorm/sandstorm-pkgdef.capnp:pkgdef

app: pb-requests.capnp.go
	go build -v -i -o app
app.stripped: app app.debug
	strip app -o app.stripped
	objcopy --add-gnu-debuglink=app.debug app.stripped
app.debug: app
	objcopy --only-keep-debug app app.debug

pb-requests.capnp.go: pb-requests.capnp
	capnp compile -ogo \
		-I $(GOPATH)/src/zombiezen.com/go/capnproto2/std \
		-I $(GOPATH)/src/zenhack.net/go/sandstorm/capnp \
		$<

# App is declared as PHONY because (1) go build manages most of the dependency
# tracking, and (2) it's a pain to get make to find all of the deps, so
# sometimes it doesn't rebuild when it should.
#
# The downside is we sometimes end up re-building app.debug and app.stripped
# when we don't need to.
.PHONY: all clean app
