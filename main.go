package main

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	bridge_capnp "zenhack.net/go/sandstorm/capnp/sandstormhttpbridge"
	"zenhack.net/go/sandstorm/exp/sandstormhttpbridge"
)

func chkfatal(err error) {
	if err != nil {
		panic(err)
	}
}

// Figure out what "action" from sandstorm-pkgdef.capnp created this
// grain.
func getAction() string {
	if len(os.Args) != 2 {
		panic("len(os.Args) != 2")
	}
	action := os.Args[1]
	if action == "restore" {
		// We previously saved our on-creation action; load it
		// from the file.
		data, err := ioutil.ReadFile("/var/action")
		chkfatal(err)
		action = string(data)
	} else {
		// Save the action so we can figure out what it was when
		// we're restored.
		chkfatal(ioutil.WriteFile("/var/action", []byte(action), 0600))
	}
	return action
}

func main() {
	ctx := context.Background()
	action := getAction()

	switch action {
	case "localfs":
		p, f := NewBridgePromise()
		appHooks := bridge_capnp.AppHooks_ServerToClient(initLocalFS(p), nil)
		bridge, err := sandstormhttpbridge.ConnectWithHooks(ctx, appHooks)
		chkfatal(err)
		f <- bridge
	case "httpview":
		bridge, err := sandstormhttpbridge.Connect(ctx)
		chkfatal(err)
		initHTTPFS(bridge)
	case "zip-uploader":
		bridge, err := sandstormhttpbridge.Connect(ctx)
		chkfatal(err)
		initZipUploader(bridge)
	default:
		panic("Unexpected action type: " + action)
	}
	listenAddr := net.JoinHostPort("", os.Getenv("LISTEN_PORT"))
	panic(http.ListenAndServe(listenAddr, nil))
}
