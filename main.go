package main

import (
	"context"
	"io/ioutil"
	"os"

	"zenhack.net/go/sandstorm/exp/websession"
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
		initLocalFS()
		panic(websession.ListenAndServe(ctx, NewLocalFS(), nil))
	case "httpview":
		initHTTPFS()
	case "zip-uploader":
		initZipUploader()
	default:
		panic("Unexpected action type: " + action)
	}
	panic(websession.ListenAndServe(nil, nil, nil))
}
