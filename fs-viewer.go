package main

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"zombiezen.com/go/capnproto2"

	"zenhack.net/go/sandstorm-filesystem/filesystem"
	grain_capnp "zenhack.net/go/sandstorm/capnp/grain"
	util_capnp "zenhack.net/go/sandstorm/capnp/util"
	"zenhack.net/go/sandstorm/grain"
	"zenhack.net/go/sandstorm/util"
)

var (
	lck sync.Mutex

	rootDir *CapnpHTTPFileSystem

	InvalidArgument = errors.New("Invalid argument")
)

type CapnpHTTPFileSystem struct {
	Dir filesystem.Directory
}

type CapnpHTTPFile struct {
	Node filesystem.Node
	Name string
	Info *FileInfo
	pos  int64
}

func (f *CapnpHTTPFile) Close() error {
	return f.Node.Client.Close()
}

func (f *CapnpHTTPFile) Read(buf []byte) (n int, err error) {
	if f.Info.IsDir() {
		return 0, InvalidArgument
	}
	r, w := io.Pipe()
	file := filesystem.File{Client: f.Node.Client}
	file.Read(context.TODO(), func(p filesystem.File_read_Params) error {
		p.SetStartAt(f.pos)
		p.SetAmount(uint64(len(buf)))
		p.SetSink(util_capnp.ByteStream_ServerToClient(&util.WriteCloserByteStream{w}))
		return nil
	})
	n, err = io.ReadFull(r, buf)
	if err == io.ErrUnexpectedEOF {
		// ReadFull expects to read the full buffer, but
		// a short read is OK for Read in general.
		err = nil
	}
	r.Close()
	return n, err
}

func (f *CapnpHTTPFile) Seek(offset int64, whence int) (int64, error) {
	if f.Info.IsDir() {
		return 0, InvalidArgument
	}
	oldPos := f.pos
	switch whence {
	case io.SeekStart:
		f.pos = offset
	case io.SeekCurrent:
		f.pos += offset
	case io.SeekEnd:
		f.pos = f.Info.Size() + offset
	default:
		return f.pos, InvalidArgument
	}
	if f.pos < 0 {
		f.pos = oldPos
		return f.pos, InvalidArgument
	}
	return f.pos, nil
}

func (f *CapnpHTTPFile) Stat() (os.FileInfo, error) {
	return f.Info, nil
}

type fiStream struct {
	isClosed bool
	limit    bool
	have     int
	buf      []os.FileInfo
	done     chan struct{}
}

func (s *fiStream) Done(p filesystem.Directory_Entry_Stream) {
	if !s.isClosed {
		s.isClosed = true
		s.buf = s.buf[:s.have]
		s.done <- struct{}{}
	}
}

func (s *fiStream) Push(p filesystem.Directory_Entry_Stream) error {
	// TODO: implement
	return nil
}

func (f *CapnpHTTPFile) Readdir(count int) ([]os.FileInfo, error) {
	if !f.Info.IsDir() {
		return nil, InvalidArgument
	}
	//dir := filesystem.Directory{Client: f.Node.Client}
	ret := &fiStream{
		buf:  []os.FileInfo{},
		done: make(chan struct{}),
	}
	if count > 0 {
		ret.buf = make([]os.FileInfo, count)
		ret.limit = true
	}
	// TODO: finish
	return nil, nil
}

type FileInfo struct {
	name string
	info filesystem.StatInfo
}

func (fi *FileInfo) Name() string {
	return fi.name
}

func (fi *FileInfo) Size() int64 {
	if fi.info.Which() == filesystem.StatInfo_Which_file {
		return fi.info.File().Size()
	} else {
		return 0
	}
}

func (fi *FileInfo) Mode() os.FileMode {
	mode := os.FileMode(0400)
	if fi.info.Executable() {
		mode |= 0100
	}
	if fi.info.Writable() {
		mode |= 0200
	}
	if fi.info.Which() == filesystem.StatInfo_Which_dir {
		mode |= os.ModeDir
	}
	return mode
}

func (fi *FileInfo) ModTime() (mtime time.Time) {
	// TODO: right now the schema doesn't carry this information;
	// might want to fix that.
	return
}

func (fi *FileInfo) IsDir() bool {
	return fi.Mode().IsDir()
}

func (fi *FileInfo) Sys() interface{} {
	return fi.info
}

func (fs *CapnpHTTPFileSystem) Open(name string) (http.File, error) {
	parts := strings.Split(name, "/")
	toClose := make([]filesystem.Node, len(parts))

	var node filesystem.Node
	var dir filesystem.Directory
	node = filesystem.Node{Client: fs.Dir.Client}

	for i, nodeName := range parts {
		node = dir.Walk(context.TODO(), func(p filesystem.Directory_walk_Params) error {
			p.SetName(nodeName)
			return nil
		}).Node()
		toClose[i] = node
		dir = filesystem.Directory{node.Client}
	}
	for i := range toClose[:len(toClose)-1] {
		toClose[i].Client.Close()
	}
	ret, err := node.Stat(context.TODO(), func(p filesystem.Node_stat_Params) error {
		return nil
	}).Struct()
	if err != nil {
		return nil, err
	}
	info, err := ret.Info()
	if err != nil {
		return nil, err
	}
	var retName string
	if len(parts) == 0 {
		retName = ""
	} else {
		retName = parts[len(parts)-1]
	}
	return &CapnpHTTPFile{
		Node: node,
		Info: &FileInfo{
			name: retName,
			info: info,
		},
	}, nil
}

func withLock(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		lck.Lock()
		defer lck.Unlock()
		h.ServeHTTP(w, req)
	})
}

func initHTTPFS() {
	r := mux.NewRouter()

	badReq := func(w http.ResponseWriter, ctx string, err error) {
		log.Print(ctx, ":", err)
		w.WriteHeader(400)
		w.Write([]byte("Bad Request"))
	}

	r.Methods("POST").Path("/filesystem-cap").
		HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			buf, err := ioutil.ReadAll(req.Body)
			if err != nil {
				badReq(w, "read body", err)
				return
			}

			sessionCtx := w.(grain.HasSessionContext).GetSessionContext()
			results, err := sessionCtx.ClaimRequest(
				context.TODO(),
				func(p grain_capnp.SessionContext_claimRequest_Params) error {
					p.SetRequestToken(string(buf))
					return nil
				}).Struct()
			if err != nil {
				badReq(w, "claim request", err)
				return
			}
			capability, err := results.Cap()
			if err != nil {
				log.Print("Error claiming network cap:", err)
				return
			}
			rootDir = &CapnpHTTPFileSystem{Dir: filesystem.Directory{
				Client: capnp.ToInterface(capability).Client(),
			}}
		})

	r.Methods("GET").PathPrefix("/fs/").
		HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		})

	http.Handle("/", withLock(r))
}
