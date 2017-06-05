
all: app.stripped app.debug
clean:
	rm -f app app.stripped app.debug
dev: all
	spk dev -p .sandstorm/sandstorm-pkgdef.capnp:pkgdef

app:
	go build -v -i -o app
app.stripped: app app.debug
	strip app -o app.stripped
	objcopy --add-gnu-debuglink=app.debug app.stripped
app.debug: app
	objcopy --only-keep-debug app app.debug

SANDSTORM_HOME ?= $(HOME)/src/foreign/sandstorm

%-powerbox-request.base64: %-powerbox-request.capnp
	python2 encode_powerbox_request.py $(SANDSTORM_HOME) $< > $@

# App is declared as PHONY because (1) go build manages most of the dependency
# tracking, and (2) it's a pain to get make to find all of the deps, so
# sometimes it doesn't rebuild when it should.
#
# The downside is we sometimes end up re-building app.debug and app.stripped
# when we don't need to.
.PHONY: all clean app
