package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"zenhack.net/go/sandstorm-filesystem/filesystem"
	"zenhack.net/go/sandstorm-filesystem/filesystem/httpfs"

	grain_capnp "zenhack.net/go/sandstorm/capnp/grain"
	grain_ctx "zenhack.net/go/sandstorm/grain/context"
)

var (
	rootDir *httpfs.FileSystem
)

func initHTTPFS() {
	r := mux.NewRouter()

	badReq := func(w http.ResponseWriter, ctx string, err error) {
		log.Print(ctx, ":", err)
		w.WriteHeader(400)
		w.Write([]byte("Bad Request"))
	}

	r.Methods("POST").Path("/filesystem-cap").
		HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			buf, err := ioutil.ReadAll(req.Body)
			if err != nil {
				badReq(w, "read body", err)
				return
			}

			sessionCtx := grain_ctx.GetSessionContext(req.Context())
			res, release := sessionCtx.ClaimRequest(
				context.TODO(),
				func(p grain_capnp.SessionContext_claimRequest_Params) error {
					p.SetRequestToken(string(buf))
					return nil
				})
			defer release()
			results, err := res.Struct()
			if err != nil {
				badReq(w, "claim request", err)
				return
			}
			capability, err := results.Cap()
			if err != nil {
				log.Print("Error claiming network cap:", err)
				return
			}
			rootDir = &httpfs.FileSystem{Dir: filesystem.Directory{
				Client: capability.Interface().Client(),
			}}
		})

	r.Methods("GET").Path("/").
		HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			tpls.ExecuteTemplate(w, "fs-viewer-index.html", struct{ HaveFS bool }{
				rootDir != nil,
			})
		})

	r.Methods("GET").PathPrefix("/fs/").
		HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if rootDir == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			http.FileServer(rootDir).ServeHTTP(w, req)
		})

	r.Methods("GET").Path("/pb-request.js").
		Handler(PbRequest(DirectoryReq))

	http.Handle("/", withLock(r))
}
