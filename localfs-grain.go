package main

// Implements a "local fs" grain, which allows other grains to access
// its files.

func NewLocalFS() *LocalFS {
	// Make sure our shared directory exists.
	chkfatal(os.MkdirAll("/var/shared-dir"))

	return &LocalFS{}
}

type LocalFS struct {
}

func (l *LocalFS) GetViewInfo(p grain_capnp.UiView_getViewInfo) error {
	pogs.Insert(grain_capnp.UiView_ViewInfo_TypeID, p.Results.Struct, viewInfo{
		MatchRequests: []PowerBoxDescriptor{
			{
				Tags: []Tag{
					{
						Id: filesystem.Node_TypeID,
					},
				},
			},
		},
	})
	return nil
}

func (l *LocalFS) NewSession(p grain_capnp.UiView_newSession) error { return nil }

func (l *LocalFS) NewRequestSession(p grain_capnp.UiView_newRequestSession) error {
}

func (l *LocalFS) NewOfferSession(p grain_capnp.UiView_newOfferSession) error { return nil }
