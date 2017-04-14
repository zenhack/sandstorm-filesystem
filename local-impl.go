package main

// Implemenatations of the filesystem interfaces on top of the local
// filesystem.

import (
	"errors"
	"zenhack.net/go/sandstorm-filesystem/filesystem"
)

var (
	IllegalFileNameError = errors.New("Illegal file name")
)

type LocalNode struct {
	canWrite bool
	path     string
	typ      filesystem.Node_Type
}

type LocalDir struct {
	LocalNode
}

type LocalFile struct {
	LocalNode
}

func (n *LocalNode) Type(p filesystem.Node_type) error {
	p.Results.SetType(n.typ)
	return nil
}

func (n *LocalNode) CanWrite(p filesystem.Node_canWrite) error {
	p.Results.SetCanWrite(d.canWrite)
}

func (d *LocalDir) List(p filesystem.Directory_list) error {
	file, err := os.Open(d.path)
	if err != nil {
		// FIXME: might possibly contain private info, e.g. where in
		// our filesystem this is rooted.
		return err
	}
	defer file.Close()
	fis, err := file.Readdir()

	if err != nil {
		return err
	}

	list := p.Results.NewList(len(fis))
	for i := range fis {
		fi := fis[i]
		ent := list.At(i)
		ent.SetName(fi.Name())

		node := &LocalNode{
			path:     d.path + "/" + fi.Name(),
			canWrite: d.canWrite,
			typ:      fiNodeType(fi),
		}

		ent.SetFile(node.MakeClient())
	}

	return nil
}

func (d *LocalDir) Open(p filesystem.Directory_open) error {
	name, err := p.Params.Name()
	if err != nil {
		return err
	}

	if !validFileName(name) {
		return IllegalFileNameError
	}

	path := d.path + "/" + name
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	node := &LocalNode{
		path:     path,
		canWrite: d.canWrite,
		typ:      fiNodeType(fi),
	}

	p.Results.SetNode(node.MakeClient())
}

func validFileName(name string) bool {
	return name != "" &&
		name != "." &&
		name != ".." &&
		!strings.Contains(name, "/")
}

func fiNodeType(fi os.FileInfo) filesystem.Node_Type {
	if fi.IsDir() {
		return filesystem.Node_Type_dir
	} else {
		return filesystem.Node_Type_file
	}
}

func (n *LocalNode) MakeClient() filesystem.Node {
	var client capnp.Client
	switch n.typ {
	case filesystem.Node_Type_dir:
		d := (*LocalDir)(n)
		if n.canWrite {
			client = filesystem.RwDirectory_ServerToClient(d).Client
		} else {
			client = filesystem.Directory_ServerToClient(d).Client
		}
	case filesystem.Node_Type_file:
		f := (*LocalFile)(n)
		if n.canWrite {
			client = filesystem.RwFile_ServerToClient(f).Client
		} else {
			client = filesystem.File_ServerToClient(f).Client
		}
	}
	return filesystem.Node{Client: client}
}
