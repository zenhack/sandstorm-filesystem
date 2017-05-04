// Package local implemenst the filesystem interfaces on top of the
// operating system's filesystem.
package local

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"zenhack.net/go/sandstorm-filesystem/filesystem"
	grain_capnp "zenhack.net/go/sandstorm/capnp/grain"
	util_capnp "zenhack.net/go/sandstorm/capnp/util"
	"zenhack.net/go/sandstorm/util"
	"zombiezen.com/go/capnproto2"
)

var (
	IllegalFileName = errors.New("Illegal file name")
	OpenFailed      = errors.New("Open failed")

	illegalWhence = errors.New("Illegal value for 'whence'")

	NotImplemented = errors.New("Not implemented")
)

func NewNode(path string) (*Node, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	ret := &Node{
		path:     path,
		typ:      filesystem.Node_Type_file,
		canWrite: fi.Mode()&0200 != 0,
		isExec:   fi.Mode()&0100 != 0,
	}
	if fi.IsDir() {
		ret.typ = filesystem.Node_Type_dir
	}
	return ret, nil
}

type Node struct {
	canWrite bool
	path     string
	typ      filesystem.Node_Type

	isExec bool // only meaningful for files
}

type Directory struct {
	Node
}

type File struct {
	Node
}

func (n *Node) Save(p grain_capnp.AppPersistent_save) error {
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}
	u8list, err := capnp.NewData(p.Results.Struct.Segment(), data)
	if err != nil {
		return err
	}
	p.Results.SetObjectIdPtr(u8list.List.ToPtr())
	return nil
}

func (n *Node) Restore(p grain_capnp.MainView_restore) error {
	ptr, err := p.Params.ObjectIdPtr()
	if err != nil {
		return err
	}
	err = json.Unmarshal(ptr.Data(), n)
	if err != nil {
		return err
	}
	capId := p.Results.Struct.Segment().Message().AddCap(n.MakeClient().Client)
	p.Results.SetCapPtr(capnp.NewInterface(p.Results.Struct.Segment(), capId).ToPtr())
	return nil
}

func (n *Node) Type(p filesystem.Node_type) error {
	p.Results.SetType(n.typ)
	return nil
}

func (n *Node) CanWrite(p filesystem.Node_canWrite) error {
	p.Results.SetCanWrite(n.canWrite)
	return nil
}

func parseWhence(whence filesystem.Whence) (int, error) {
	switch whence {
	case filesystem.Whence_start:
		return 0, nil
	case filesystem.Whence_end:
		return 2, nil
	default:
		return -1, illegalWhence
	}
}

type cancelHandle context.CancelFunc

func (c cancelHandle) Close() error {
	c()
	return nil
}

func (d *Directory) List(p filesystem.Directory_list) error {
	stream := p.Params.Stream()
	file, err := os.Open(d.path)
	if err != nil {
		// err might contain private info, e.g. where the directory
		// is rooted. So we return a generic error. It might be nice
		// to find some way to allow more information for debugging.
		return OpenFailed
	}
	ctx, cancel := context.WithCancel(p.Ctx)
	p.Results.SetCancel(util_capnp.Handle_ServerToClient(cancelHandle(cancel)))
	go func() {
		defer file.Close()
		defer stream.Done(ctx, func(filesystem.Directory_Entry_Stream_done_Params) error {
			return nil
		})

		maxBufSize := 1024

		for ctx.Err() == nil {
			fis, err := file.Readdir(maxBufSize)
			if err != nil {
				// TODO: can we communicate failures somehow? This
				// could mean EOF or a legitmate problem, but we don't
				// currently have a good way to convey the latter to the
				// caller
				return
			}

			stream.Push(ctx, func(p filesystem.Directory_Entry_Stream_push_Params) error {
				list, err := p.NewEntries(int32(len(fis)))
				if err != nil {
					return err
				}
				for i := range fis {
					fi := fis[i]
					ent := list.At(i)
					ent.SetName(fi.Name())
					ent.SetCanWrite(d.canWrite && fi.Mode()&0200 != 0)
					if fi.IsDir() {
						ent.Node().SetDir()
					} else {
						node := ent.Node()
						node.SetFile()
						node.File().SetIsExec(fi.Mode()&0100 != 0)
					}

				}
				return nil
			})
		}

	}()
	return nil
}

func (d *Directory) Walk(p filesystem.Directory_walk) error {
	name, err := p.Params.Name()
	if err != nil {
		return err
	}

	if !validFileName(name) {
		return IllegalFileName
	}

	path := d.path + "/" + name
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	node := &Node{
		path:     path,
		canWrite: d.canWrite && fi.Mode()&0200 != 0,
		typ:      fiNodeType(fi),
		isExec:   fi.Mode()&0100 != 0,
	}

	p.Results.SetNode(node.MakeClient())
	return nil
}

func (d *Directory) Create(p filesystem.RwDirectory_create) error {
	name, err := p.Params.Name()
	if err != nil {
		return err
	}
	isExec := p.Params.IsExec()

	if !validFileName(name) {
		return IllegalFileName
	}

	path := d.path + "/" + name

	mode := os.FileMode(0644)
	if isExec {
		mode |= 0111
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, mode)
	if err != nil {
		return OpenFailed
	}
	file.Close()

	p.Results.SetFile(filesystem.RwFile_ServerToClient(&File{
		Node{
			path:     path,
			canWrite: true,
			typ:      filesystem.Node_Type_file,
			isExec:   isExec,
		},
	}))
	return nil
}

func (d *Directory) MkDir(p filesystem.RwDirectory_mkDir) error {
	return NotImplemented
}

func (d *Directory) Delete(p filesystem.RwDirectory_delete) error {
	return NotImplemented
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

func (n *Node) MakeClient() filesystem.Node {
	var client capnp.Client
	switch n.typ {
	case filesystem.Node_Type_dir:
		d := &Directory{*n}
		if n.canWrite {
			client = filesystem.RwDirectory_ServerToClient(d).Client
		} else {
			client = filesystem.Directory_ServerToClient(d).Client
		}
	case filesystem.Node_Type_file:
		f := &File{*n}
		if n.canWrite {
			client = filesystem.RwFile_ServerToClient(f).Client
		} else {
			client = filesystem.File_ServerToClient(f).Client
		}
	}
	return filesystem.Node{Client: client}
}

func (f File) Write(p filesystem.RwFile_write) error {
	return NotImplemented
}

func (f *File) IsExec(p filesystem.File_isExec) error {
	p.Results.SetIsExec(f.isExec)
	return nil
}

func (f *File) Size(p filesystem.File_size) error {
	fi, err := os.Stat(f.path)
	if err != nil {
		// FIXME: censor error like with OpenFailed.
		return err
	}
	p.Results.SetSize(uint64(fi.Size()))
	return nil
}

func (f *File) SetExec(p filesystem.RwFile_setExec) error {
	exec := p.Params.Exec()
	fi, err := os.Stat(f.path)
	// FIXME: censor error like with OpenFailed.
	if err != nil {
		return err
	}
	if exec {
		// FIXME: censor error like with OpenFailed.
		return os.Chmod(f.path, fi.Mode()|0111)
	} else {
		// FIXME: censor error like with OpenFailed.
		return os.Chmod(f.path, fi.Mode()&^0111)
	}
}

func (f *File) Truncate(p filesystem.RwFile_truncate) error {
	// FIXME: cast/overflow issues.
	if err := os.Truncate(f.path, int64(p.Params.Size())); err != nil {
		return OpenFailed
	}
	return nil
}

func (f *File) Read(p filesystem.File_read) error {
	startAt := p.Params.StartAt()
	amount := p.Params.Amount()
	sink := p.Params.Sink()

	file, err := os.Open(f.path)
	if err != nil {
		return OpenFailed
	}

	ctx, cancel := context.WithCancel(p.Ctx)
	p.Results.SetCancel(util_capnp.Handle_ServerToClient(cancelHandle(cancel)))

	go func() {
		defer file.Close()
		wc := util.ByteStreamWriteCloser{ctx, sink}
		defer wc.Close()
		// XXX: The int64 cast here (and below) isn't really valid; should think
		// about what to do.
		_, err := file.Seek(int64(startAt), 0)
		r := io.Reader(file)
		if amount != 0 {
			r = io.LimitReader(r, int64(amount))
		}
		if err != nil {
			return
		}
		io.Copy(wc, r)
	}()
	return nil
}
