package main

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"zombiezen.com/go/capnproto2"

	"zenhack.net/go/sandstorm-filesystem/filesystem"
	grain_capnp "zenhack.net/go/sandstorm/capnp/grain"
	util_capnp "zenhack.net/go/sandstorm/capnp/util"
	grain_ctx "zenhack.net/go/sandstorm/grain/context"
	"zenhack.net/go/sandstorm/util"
)

var (
	rootRwDir filesystem.RwDirectory
)

func initZipUploader() {
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
			results, err := sessionCtx.ClaimRequest(
				context.TODO(),
				func(p grain_capnp.SessionContext_claimRequest_Params) error {
					p.SetRequestToken(string(buf))
					return nil
				}).Struct()
			if err != nil {
				badReq(w, "claim request", err)
				return
			}
			capability, err := results.Cap()
			if err != nil {
				log.Print("Error claiming network cap:", err)
				return
			}
			rootRwDir = filesystem.RwDirectory{
				Client: capnp.ToInterface(capability).Client(),
			}
		})

	r.Methods("GET").Path("/").
		HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			tpls.ExecuteTemplate(w, "zip-uploader-index.html", struct{ HaveFS bool }{
				rootRwDir.Client != nil,
			})
		})

	r.Methods("POST").Path("/zipfile").
		HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			badReq := func(s string) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(s))
				return
			}
			if rootRwDir.Client == nil {
				badReq("Don't have filesystem cap")
				return
			}
			formFile, _, err := req.FormFile("zipfile")
			if err != nil {
				badReq(err.Error())
				return
			}
			buf := &bytes.Buffer{}
			_, err = io.Copy(buf, formFile)
			if err != nil {
				badReq(err.Error())
				return
			}
			r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), 0)
			if err != nil {
				badReq(err.Error())
				return
			}
			for _, f := range r.File {
				if f.FileInfo().IsDir() {
					// TODO: handle directories
					continue
				}
				file, err := f.Open()
				if err != nil {
					badReq(err.Error())
					return
				}
				ctx, cancel := context.WithCancel(req.Context())
				out := rootRwDir.Create(
					ctx,
					func(p filesystem.RwDirectory_create_Params) error {
						p.SetName(f.Name)
						p.SetExecutable((f.Mode() & 0111) != 0)
						return nil
					}).File().Write(
					ctx,
					func(p filesystem.RwFile_write_Params) error {
						p.SetStartAt(0)
						p.SetCancel(util_capnp.Handle_ServerToClient(cancel))
						return nil
					}).Sink()
				wc := &util.ByteStreamWriteCloser{
					Obj: out,
					Ctx: ctx,
				}
				_, err = io.Copy(wc, file)
				out.Client.Close()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Print(err)
					return
				}
			}
		})

	r.Methods("GET").PathPrefix("/static/").
		Handler(http.FileServer(http.Dir("")))

	http.Handle("/", withLock(r))
}
