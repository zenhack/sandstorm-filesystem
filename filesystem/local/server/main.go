package main

import (
	"context"
	"flag"
	"log"
	"net"

	"zombiezen.com/go/capnproto2/rpc"

	"zenhack.net/go/sandstorm-filesystem/filesystem"
	"zenhack.net/go/sandstorm-filesystem/filesystem/local"
)

var (
	path = flag.String("path", "", "Path to directory to serve")
	addr = flag.String("socket", "", "Path to unix socket to listen on")
)

type BootstrapImpl struct {
	rootDir filesystem.RwDirectory
}

func (bs *BootstrapImpl) Rootfs(ctx context.Context, p Bootstrap_rootfs) error {
	res, err := p.AllocResults()
	if err != nil {
		return err
	}
	res.SetDir(bs.rootDir)
	return nil
}

func main() {
	flag.Parse()
	rootDir, err := local.NewNode(*path)
	if err != nil {
		log.Fatal("NewNode(): ", err)
	}
	l, err := net.Listen("unix", *addr)
	if err != nil {
		log.Fatalf("Error on Listen(): ", err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Print("Error on Accept():", err)
			continue
		}
		go handleConn(rootDir, conn)
	}
}

func handleConn(rootDir *local.Node, conn net.Conn) {
	defer conn.Close()
	log.Print("Got connection")
	rootClient := filesystem.RwDirectory_ServerToClient(rootDir, nil)
	bootstrap := Bootstrap_ServerToClient(&BootstrapImpl{rootDir: rootClient}, nil)
	rpcConn := rpc.NewConn(rpc.NewStreamTransport(conn), &rpc.Options{
		BootstrapClient: bootstrap.Client,
	})
	<-rpcConn.Done()
	log.Println("Disconnected.")
}
