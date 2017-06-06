package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	grain_capnp "zenhack.net/go/sandstorm/capnp/grain"
	//ws_capnp "zenhack.net/go/sandstorm/capnp/websession"
	"zenhack.net/go/sandstorm/grain"
	"zenhack.net/go/sandstorm/websession"
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
		_, err := grain.ConnectAPI(ctx, grain_capnp.UiView{
			Client: grain_capnp.MainView_ServerToClient(NewLocalFS()).Client,
		})
		chkfatal(err)
	case "httpview":
		initHTTPFS()
		_, err := grain.ConnectAPI(ctx, grain_capnp.UiView_ServerToClient(
			websession.FromHandler(http.DefaultServeMux),
		))
		chkfatal(err)
	case "zip-uploader":
		initZipUploader()
		_, err := grain.ConnectAPI(ctx, grain_capnp.UiView_ServerToClient(
			websession.FromHandler(http.DefaultServeMux),
		))
		chkfatal(err)
	default:
		panic("Unexpected action type: " + action)
	}
	<-ctx.Done()
}
