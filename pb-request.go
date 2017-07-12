package main

import (
	"bytes"
	"encoding/base64"
	"net/http"

	"zenhack.net/go/sandstorm/capnp/powerbox"
	"zombiezen.com/go/capnproto2"
)

func PbRequest(desc powerbox.PowerboxDescriptor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf := &bytes.Buffer{}
		b64enc := base64.NewEncoder(
			base64.URLEncoding,
			buf,
		)
		err := capnp.NewPackedEncoder(b64enc).
			Encode(desc.Segment().Message())
		if err != nil {
			panic(err)
		}
		b64enc.Close()
		tpls.ExecuteTemplate(w, "pb-request.js", buf.String())
	})
}
