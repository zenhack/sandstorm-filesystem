package main

import (
	"encoding/base64"
	"net/http"

	"zenhack.net/go/sandstorm/capnp/powerbox"
)

func PbRequest(desc powerbox.PowerboxDescriptor) http.Handler {
	msg, err := desc.Segment().Message().MarshalPacked()
	if err != nil {
		panic(err)
	}
	b64msg := base64.URLEncoding.EncodeToString(msg)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		tpls.ExecuteTemplate(w, "pb-request.js", b64msg)
	})
}
