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
	bridge_capnp "zenhack.net/go/sandstorm/capnp/sandstormhttpbridge"
	"zenhack.net/go/sandstorm/exp/util/bytestream"

	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/server"
)

var (
	InvalidArgument = errors.New("Invalid argument")
	IllegalFileName = errors.New("Illegal file name")
	OpenFailed      = errors.New("Open failed")
	NotImplemented  = errors.New("Not implemented")
)

func NewNode(path string) (*Node, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &Node{
		Path:       path,
		IsDir:      fi.IsDir(),
		Writable:   fi.Mode()&0200 != 0,
		Executable: fi.Mode()&0100 != 0,
	}, nil
}

type Node struct {
	IsDir      bool
	Writable   bool
	Executable bool
	Path       string
}

func (n *Node) Save(ctx context.Context, p grain_capnp.AppPersistent_save) error {
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}
	res, err := p.AllocResults()
	if err != nil {
		return err
	}
	u8list, err := capnp.NewData(res.Struct.Segment(), data)
	if err != nil {
		return err
	}
	res.SetObjectId(u8list.List.ToPtr())
	return nil
}

func (n *Node) Restore(p bridge_capnp.AppHooks_restore) error {
	ptr, err := p.Args().ObjectId()
	if err != nil {
		return err
	}
	err = json.Unmarshal(ptr.Data(), n)
	if err != nil {
		return err
	}
	res, err := p.AllocResults()
	if err != nil {
		return err
	}
	capId := res.Struct.Segment().Message().AddCap(n.MakeClient().Client)
	res.SetCap(capnp.NewInterface(res.Struct.Segment(), capId).ToPtr())
	return nil
}

func (n *Node) Stat(ctx context.Context, p filesystem.Node_stat) error {
	fi, err := os.Stat(n.Path)
	if err != nil {
		// TODO: think about the right way to handle this.
		return err
	}
	res, err := p.AllocResults()
	if err != nil {
		return err
	}
	info, err := res.NewInfo()
	if err != nil {
		return err
	}
	if n.IsDir {
		info.SetDir()
	} else {
		info.SetFile()
		info.File().SetSize(fi.Size())
	}
	info.SetWritable(n.Writable)
	info.SetExecutable(n.Executable)
	return nil
}

func (d *Node) List(ctx context.Context, p filesystem.Directory_list) error {
	stream := p.Args().Stream()
	file, err := os.Open(d.Path)
	if err != nil {
		// err might contain private info, e.g. where the directory
		// is rooted. So we return a generic error. It might be nice
		// to find some way to allow more information for debugging.
		return OpenFailed
	}
	defer file.Close()
	maxBufSize := 1024

	for ctx.Err() == nil {
		fis, err := file.Readdir(maxBufSize)
		if err != nil && err != io.EOF {
			return err
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
				info, err := ent.NewInfo()
				if err != nil {
					return err
				}
				info.SetWritable(d.Writable && (fi.Mode()&0200 != 0))
				info.SetExecutable(fi.Mode()&0100 != 0)
				if fi.IsDir() {
					info.SetDir()
				} else {
					info.SetFile()
					info.File().SetSize(fi.Size())
				}
			}
			return nil
		})

		if err == io.EOF {
			break
		}
	}

	stream.Done(ctx, func(filesystem.Directory_Entry_Stream_done_Params) error {
		return nil
	})

	return nil
}

func (d *Node) Walk(ctx context.Context, p filesystem.Directory_walk) error {
	name, err := p.Args().Name()
	if err != nil {
		return err
	}

	if !validFileName(name) {
		return IllegalFileName
	}

	path := d.Path + "/" + name
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	node := &Node{
		Path:       path,
		IsDir:      fi.IsDir(),
		Writable:   d.Writable && fi.Mode()&0200 != 0,
		Executable: fi.Mode()&0100 != 0,
	}

	res, err := p.AllocResults()
	if err != nil {
		return err
	}

	res.SetNode(node.MakeClient())
	return nil
}

func (d *Node) Create(ctx context.Context, p filesystem.RwDirectory_create) error {
	name, err := p.Args().Name()
	if err != nil {
		return err
	}
	if !validFileName(name) {
		return IllegalFileName
	}

	node := Node{
		Path:       d.Path + "/" + name,
		Executable: p.Args().Executable(),
		Writable:   true,
	}

	mode := os.FileMode(0644)
	if node.Executable {
		mode |= 0111
	}

	file, err := os.OpenFile(node.Path, os.O_RDWR|os.O_CREATE, mode)
	if err != nil {
		return OpenFailed
	}
	file.Close()

	res, err := p.AllocResults()
	if err != nil {
		return err
	}
	res.SetFile(filesystem.RwFile{
		Client: node.MakeClient().Client,
	})
	return nil
}

func (d *Node) Mkdir(ctx context.Context, p filesystem.RwDirectory_mkdir) error {
	name, err := p.Args().Name()
	if err != nil {
		return err
	}
	if !validFileName(name) {
		return IllegalFileName
	}
	return os.Mkdir(d.Path+"/"+name, 0700)
}

func (d *Node) Delete(ctx context.Context, p filesystem.RwDirectory_delete) error {
	return NotImplemented
}

func validFileName(name string) bool {
	return name != "" &&
		name != "." &&
		name != ".." &&
		!strings.Contains(name, "/")
}

func (n *Node) MakeClient() filesystem.Node {
	var methods []server.Method
	if n.IsDir {
		if n.Writable {
			methods = filesystem.RwDirectory_Methods(nil, n)
		} else {
			methods = filesystem.Directory_Methods(nil, n)
		}
	} else {
		if n.Writable {
			methods = filesystem.RwFile_Methods(nil, n)
		} else {
			methods = filesystem.File_Methods(nil, n)
		}
	}
	return filesystem.Node{
		Client: capnp.NewClient(server.New(
			append(
				methods,
				grain_capnp.AppPersistent_Methods(nil, n)...,
			),
			n,
			nil,
			nil,
		)),
	}
}

func (f *Node) Write(ctx context.Context, p filesystem.RwFile_write) error {
	startAt := p.Args().StartAt()

	if startAt <= -2 {
		return InvalidArgument
	}

	file, err := os.OpenFile(f.Path, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	if startAt == -1 {
		_, err = file.Seek(0, 2)
	} else {
		_, err = file.Seek(startAt, 0)
	}
	if err != nil {
		file.Close()
		return err
	}
	bs := bytestream.FromWriteCloser(file, nil)
	res, err := p.AllocResults()
	if err != nil {
		return err
	}
	res.SetSink(bs)
	return nil
}

func (f *Node) SetExec(ctx context.Context, p filesystem.RwFile_setExec) error {
	exec := p.Args().Exec()
	fi, err := os.Stat(f.Path)
	// FIXME: censor error like with OpenFailed.
	if err != nil {
		return err
	}
	if exec {
		// FIXME: censor error like with OpenFailed.
		return os.Chmod(f.Path, fi.Mode()|0111)
	} else {
		// FIXME: censor error like with OpenFailed.
		return os.Chmod(f.Path, fi.Mode()&^0111)
	}
}

func (f *Node) Truncate(ctx context.Context, p filesystem.RwFile_truncate) error {
	// FIXME: cast/overflow issues.
	if err := os.Truncate(f.Path, int64(p.Args().Size())); err != nil {
		return OpenFailed
	}
	return nil
}

func (f *Node) Read(ctx context.Context, p filesystem.File_read) error {
	startAt := p.Args().StartAt()
	if startAt < 0 {
		return InvalidArgument
	}

	amount := int64(p.Args().Amount())
	if amount < 0 {
		// The go api expects a signed value, so if we get something
		// greater than an int64 can represent, we just say "read the
		// whole thing." That's a stupid amount of data, so it's always
		// going to do the same thing anyway.
		amount = 0
	}
	sink := p.Args().Sink()

	file, err := os.Open(f.Path)
	if err != nil {
		return OpenFailed
	}
	defer file.Close()

	wc := bytestream.ToWriteCloser(ctx, sink)
	_, err = file.Seek(startAt, 0)
	r := io.Reader(file)
	if amount != 0 {
		r = io.LimitReader(r, amount)
	}
	if err != nil {
		return err
	}
	_, err = io.Copy(wc, r)
	if err != nil {
		return err
	}
	wc.Close()
	return nil
}
