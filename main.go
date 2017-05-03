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
	var uiView grain_capnp.UiView_Server

	action := getAction()

	switch action {
	case "localfs":
		uiView = NewLocalFS()
	case "hello":
		uiView = &Hello{
			websession.FromHandler(
				context.TODO(),
				http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte("Hello!"))
				})),
		}
	case "goodbye":
		fallthrough
	default:
		panic("Unexpected action type: " + action)
	}
	ctx := context.Background()
	_, err := grain.ConnectAPI(ctx, uiView)
	if err != nil {
		panic(err)
	}
	<-ctx.Done()
}

type Hello struct {
	websession.HandlerWebSession
}
