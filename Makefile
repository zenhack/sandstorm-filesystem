
all: app.stripped app.debug
clean:
	rm -f app app.stripped app.debug
dev: all
	spk dev -p .sandstorm/sandstorm-pkgdef.capnp:pkgdef

app: $(wildcard *.go)
	go build -v -i -o app
app.stripped: app app.debug
	strip app -o app.stripped
	objcopy --add-gnu-debuglink=app.debug app.stripped
app.debug: app
	objcopy --only-keep-debug app app.debug

SANDSTORM_HOME ?= $(HOME)/src/foreign/sandstorm

ro-dir-powerbox-request.base64: ro-dir-powerbox-request.capnp
	python2 encode_powerbox_request.py $(SANDSTORM_HOME) > $@

.PHONY: all clean
