package main

import (
	"context"
	"net/http"
	"os"

	"zenhack.net/go/sandstorm-filesystem/filesystem"
	"zenhack.net/go/sandstorm-filesystem/filesystem/local"

	"zenhack.net/go/sandstorm/exp/websession"

	grain_capnp "zenhack.net/go/sandstorm/capnp/grain"

	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/pogs"
)

type LocalFS struct {
	*websession.HandlerUiView
}

func (fs *LocalFS) GetViewInfo(ctx context.Context, p grain_capnp.UiView_getViewInfo) error {
	res, err := p.AllocResults()
	if err != nil {
		return err
	}
	pogs.Insert(grain_capnp.UiView_ViewInfo_TypeID, res.Struct, viewInfo{
		MatchRequests: []PowerboxDescriptor{{Tags: []Tag{
			{Id: filesystem.Node_TypeID},
			{Id: filesystem.Directory_TypeID},
			{Id: filesystem.RwDirectory_TypeID},
		}}},
	})
	return nil
}

func (fs *LocalFS) Restore(ctx context.Context, p grain_capnp.MainView_restore) error {
	node := &local.Node{}
	return node.Restore(p)
}

func (fs *LocalFS) Drop(ctx context.Context, p grain_capnp.MainView_drop) error {
	return nil
}

func handleLocalFS(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	p := websession.GetSessionData(req)
	if p.Which() != websession.SessionData_Which_request {
		// Not a request session.
		w.Write([]byte("This grain doesn't provide much of a user " +
			"interface (this is it), but you can request a " +
			"filesystem via other grains, and have this grain " +
			"fulfill them. Try out the zip uploader and filesystem " +
			"viewer grain types from this app."))
		return
	}
	sessionContext := p.Context()
	requestInfo, err := p.Request().RequestInfo()
	if err != nil || requestInfo.Len() < 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
		return
	}
	descriptor := requestInfo.At(0)
	res, release := sessionContext.FulfillRequest(
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
			p.SetCap(capnp.NewInterface(p.Struct.Segment(), capId).ToPtr())
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
		})
	defer release()
	_, err = res.Struct()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	ret, release := sessionContext.Close(ctx, func(p grain_capnp.SessionContext_close_Params) error {
		return nil
	})
	defer release()
	_, err = ret.Struct()
	if err != nil {
		println(err.Error())
	}
}

func initLocalFS() {
	// Make sure our shared directory exists.
	chkfatal(os.MkdirAll("/var/shared-dir", 0700))
	http.HandleFunc("/", handleLocalFS)
}

// Returns a "local fs" grain, which allows other grains to access
// its files.
func NewLocalFS() *LocalFS {
	return &LocalFS{&websession.HandlerUiView{http.DefaultServeMux}}
}
