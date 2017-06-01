package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"zenhack.net/go/sandstorm-filesystem/filesystem"
	"zenhack.net/go/sandstorm-filesystem/filesystem/local"
	grain_capnp "zenhack.net/go/sandstorm/capnp/grain"
	grain_ctx "zenhack.net/go/sandstorm/grain/context"
	"zenhack.net/go/sandstorm/websession"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/pogs"
)

type LocalFS struct {
	grain_capnp.UiView_Server
}

func (fs *LocalFS) Restore(p grain_capnp.MainView_restore) error {
	node := &local.Node{}
	return node.Restore(p)
}

func (fs *LocalFS) Drop(p grain_capnp.MainView_drop) error {
	return nil
}

// Returns a "local fs" grain, which allows other grains to access
// its files.
func NewLocalFS() grain_capnp.MainView_Server {
	// Make sure our shared directory exists.
	chkfatal(os.MkdirAll("/var/shared-dir", 0700))

	// Put something in the directory.
	chkfatal(ioutil.WriteFile("/var/shared-dir/index.html", []byte(
		"This is a directory shared by a sandstorm grain.",
	), 0644))

	return &LocalFS{websession.FromHandler(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			p, err := grain_ctx.GetRequestSessionParams(ctx)
			if err != nil {
				// Not a request session.
				w.Write([]byte("This grain doesn't provide much of a user " +
					"interface (this is it), but you can request a " +
					"filesystem via other apps, and have this grain " +
					"fulfill them."))
				return
			}
			sessionContext := p.Context()
			requestInfo, err := p.RequestInfo()
			if err != nil || requestInfo.Len() < 1 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Bad Request"))
				return
			}
			descriptor := requestInfo.At(0)
			_, err = sessionContext.FulfillRequest(
				ctx,
				func(p grain_capnp.SessionContext_fulfillRequest_Params) error {
					// TODO: limit to the thing the user actually asked for; if they didn't ask
					// for write, don't give it to them.
					n, err := local.NewNode("/var/shared-dir")
					if err != nil {
						// This should never happen; we create the above dir on first start.
						panic(err)
					}
					capId := p.Struct.Segment().Message().AddCap(n.MakeClient().Client)
					p.SetCapPtr(capnp.NewInterface(p.Struct.Segment(), capId).ToPtr())
					p.SetDescriptor(descriptor)
					p.NewRequiredPermissions(0)
					displayInfo, err := p.NewDisplayInfo()
					if err != nil {
						return err
					}
					title, err := displayInfo.NewTitle()
					if err != nil {
						return err
					}
					title.SetDefaultText("Grain-local filesystem.")
					return nil
				}).Struct()
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}
			_, err = sessionContext.Close(ctx, func(p grain_capnp.SessionContext_close_Params) error {
				return nil
			}).Struct()
			if err != nil {
				println(err.Error())
			}
		})).
		WithViewInfo(func(p grain_capnp.UiView_getViewInfo) error {
			pogs.Insert(grain_capnp.UiView_ViewInfo_TypeID, p.Results.Struct, viewInfo{
				MatchRequests: []PowerboxDescriptor{{Tags: []Tag{
					{Id: filesystem.Node_TypeID},
					{Id: filesystem.Directory_TypeID},
					{Id: filesystem.RwDirectory_TypeID},
				}}},
			})
			return nil
		})}
}
