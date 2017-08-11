package main

// This is a fuse filesystem that speaks the filesystem capnp protocol.
// It is still a work in progress.

import (
	"context"
	"flag"
	"io"
	"log"

	"zenhack.net/go/sandstorm-filesystem/filesystem"
	"zenhack.net/go/sandstorm-filesystem/filesystem/local"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

var (
	// For now we just export the filesystem to ourselves via capnp; will
	// actually do a network connection later.
	src = flag.String("src", "", "directory to serve")

	dst = flag.String("dst", "", "mountpoint")
)

type Node struct {
	ctx context.Context
	nodefs.Node
	capnode filesystem.Node
}

type dirEntStream struct {
	haveDone bool
	onDone   chan error
	ents     []fuse.DirEntry
}

func newDirEntStream() *dirEntStream {
	return &dirEntStream{
		haveDone: false,
		onDone:   make(chan error, 1),
		ents:     []fuse.DirEntry{},
	}
}

func modeFromStatInfo(info filesystem.StatInfo) uint32 {
	mode := uint32(0400)
	if info.Executable() {
		mode |= 0100
	}
	if info.Writable() {
		mode |= 0200
	}
	switch info.Which() {
	case filesystem.StatInfo_Which_dir:
		mode |= fuse.S_IFDIR
	case filesystem.StatInfo_Which_file:
		mode |= fuse.S_IFREG
	}
	return mode
}

func convertDirEntry(ent filesystem.Directory_Entry) (fuse.DirEntry, error) {
	info, err := ent.Info()
	if err != nil {
		return fuse.DirEntry{}, err
	}
	name, err := ent.Name()
	if err != nil {
		return fuse.DirEntry{}, err
	}
	return fuse.DirEntry{
		Mode: modeFromStatInfo(info),
		Name: name,
	}, nil
}

func (s *dirEntStream) Push(p filesystem.Directory_Entry_Stream_push) error {
	if s.haveDone {
		return nil // TODO: error
	}
	ents, err := p.Params.Entries()
	if err != nil {
		return err
	}
	for i := 0; i < ents.Len(); i++ {
		ent := ents.At(i)
		fuseEnt, err := convertDirEntry(ent)
		if err != nil {
			return err
		}
		s.ents = append(s.ents, fuseEnt)
	}
	return nil
}

func (s *dirEntStream) Done(filesystem.Directory_Entry_Stream_done) error {
	if s.haveDone {
		return nil // TODO: error
	}
	s.haveDone = true
	s.onDone <- io.EOF
	return nil
}

func (s *dirEntStream) Close() error {
	if !s.haveDone {
		s.haveDone = true
		s.onDone <- io.ErrUnexpectedEOF
	}
	return nil
}

func (n *Node) OpenDir(ctx *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	stream := newDirEntStream()
	_, err := filesystem.Directory{n.capnode.Client}.
		List(n.ctx, func(p filesystem.Directory_list_Params) error {
			p.SetStream(filesystem.Directory_Entry_Stream_ServerToClient(stream))
			return nil
		}).Struct()
	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	err = <-stream.onDone
	if err == io.EOF {
		err = nil
	}
	return stream.ents, fuse.ToStatus(err)
}

func (n *Node) GetAttr(out *fuse.Attr, file nodefs.File, context *fuse.Context) fuse.Status {
	info, err := n.capnode.Stat(n.ctx, nil).Info().Struct()
	if err != nil {
		return fuse.ToStatus(err)
	}
	out.Mode = modeFromStatInfo(info)
	out.Owner = context.Owner
	switch info.Which() {
	case filesystem.StatInfo_Which_dir:
	case filesystem.StatInfo_Which_file:
		out.Size = uint64(info.File().Size())
	}
	return fuse.OK
}

func main() {
	flag.Parse()
	if *src == "" || *dst == "" {
		log.Fatal("usage")
	}
	node, err := local.NewNode(*src)
	if err != nil {
		log.Fatal(err)
	}
	root := &Node{
		ctx:     context.Background(),
		Node:    nodefs.NewDefaultNode(),
		capnode: node.MakeClient(),
	}
	srv, _, err := nodefs.MountRoot(*dst, root, nil)
	if err != nil {
		log.Fatal(err)
	}
	srv.Serve()
}
