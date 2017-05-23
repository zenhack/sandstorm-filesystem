package main

import (
	"errors"
	"net/http"
	"os"

	"zenhack.net/go/sandstorm-filesystem/filesystem"
	"zenhack.net/go/sandstorm-filesystem/filesystem/local"
	grain_capnp "zenhack.net/go/sandstorm/capnp/grain"
	websession_capnp "zenhack.net/go/sandstorm/capnp/websession"
	"zenhack.net/go/sandstorm/websession"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/pogs"
)

// Implements a "local fs" grain, which allows other grains to access
// its files.

var (
	NoPowerboxDescriptors = errors.New("request session has no powerbox descriptors")
)

func NewLocalFS() *LocalFS {
	// Make sure our shared directory exists.
	chkfatal(os.MkdirAll("/var/shared-dir", 0700))

	return &LocalFS{}
}

type LocalFS struct {
}

func (l *LocalFS) GetViewInfo(p grain_capnp.UiView_getViewInfo) error {
	pogs.Insert(grain_capnp.UiView_ViewInfo_TypeID, p.Results.Struct, viewInfo{
		MatchRequests: []PowerboxDescriptor{
			{
				Tags: []Tag{
					{
						Id: filesystem.Node_TypeID,
					},
					{
						Id: filesystem.Directory_TypeID,
					},
					{
						Id: filesystem.RwDirectory_TypeID,
					},
				},
			},
		},
	})
	return nil
}

func (l *LocalFS) NewSession(p grain_capnp.UiView_newSession) error {
	ws := websession.FromHandler(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("This grain doesn't provide much of a user " +
				"interface (this is it), but you can request a " +
				"filesystem via other apps, and have this grain " +
				"fulfill them."))
		}))
	p.Results.SetSession(grain_capnp.UiSession{
		websession_capnp.WebSession_ServerToClient(ws).Client,
	})
	return nil
}

func (l *LocalFS) NewRequestSession(p grain_capnp.UiView_newRequestSession) error {
	sessionContext := p.Params.Context()
	sessionContext.FulfillRequest(
		p.Ctx,
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
			return nil
		})
	sessionContext.Close(p.Ctx, func(p grain_capnp.SessionContext_close_Params) error {
		return nil
	})
	return nil
}

func (l *LocalFS) NewOfferSession(p grain_capnp.UiView_newOfferSession) error {
	p.Params.Context().Close(
		p.Ctx,
		func(p grain_capnp.SessionContext_close_Params) error {
			return nil
		})
	return nil
}
