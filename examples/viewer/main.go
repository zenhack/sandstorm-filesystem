package main

import (
	"context"
	"net/http"
	ws_capnp "zenhack.net/go/sandstorm/capnp/websession"
	"zenhack.net/go/sandstorm/grain"
	"zenhack.net/go/sandstorm/websession"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	ctx := context.Background()
	ws = websession.FromHandler(ctx, http.DefaultServeMux)
	uivew := grain_capnp.UiView{ws_capnp.WebSession_ServerToClient(ws).Client}
	_, err := grain.ConnectAPI(ctx, uivew)
	if err != nil {
		panic(err)
	}
	<-ctx
}
