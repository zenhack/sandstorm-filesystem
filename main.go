package main

import (
	"context"
	"io/ioutil"
	//"net/http"
	"os"
	grain_capnp "zenhack.net/go/sandstorm/capnp/grain"
	//ws_capnp "zenhack.net/go/sandstorm/capnp/websession"
	"zenhack.net/go/sandstorm/grain"
	//"zenhack.net/go/sandstorm/websession"
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
		file, err := os.Open("/var/action")
		chkfatal(err)
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		chkfatal(err)
		action = string(data)
	} else {
		// Save the action so we can figure out what it was when
		// we're restored.
		file, err := os.Create("/var/action")
		chkfatal(err)
		defer file.Close()
		data := []byte(action)
		n, err := file.Write(data)
		chkfatal(err)
		if n != len(data) {
			panic("Short read")
		}
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
		uiView = &Hello{}
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
}
