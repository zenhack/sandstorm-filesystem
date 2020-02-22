package httpfs

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"zenhack.net/go/sandstorm-filesystem/filesystem"
	"zenhack.net/go/sandstorm/exp/util/bytestream"

	"zombiezen.com/go/capnproto2"
)

var (
	InvalidArgument = errors.New("Invalid argument")
	TooManyEntries  = errors.New("Stream received too many entries")
)

// Copy StatInfo. Useful to when the original is going to be reclaimed.
func cloneInfo(info filesystem.StatInfo) filesystem.StatInfo {
	msg, _, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		panic(err)
	}
	msg.SetRoot(info.Struct.ToPtr())
	newInfo, err := filesystem.ReadRootStatInfo(msg)
	if err != nil {
		panic(err)
	}
	return newInfo
}

type FileSystem struct {
	Dir filesystem.Directory
}

type File struct {
	Node filesystem.Node
	Name string
	Info *FileInfo
	pos  int64
}

func (f *File) Close() error {
	return nil
}

func (f *File) Read(buf []byte) (n int, err error) {
	if f.Info.IsDir() {
		return 0, InvalidArgument
	}
	r, w := io.Pipe()
	file := filesystem.File{Client: f.Node.Client}
	file.Read(context.TODO(), func(p filesystem.File_read_Params) error {
		p.SetStartAt(f.pos)
		p.SetAmount(uint64(len(buf)))
		p.SetSink(bytestream.FromWriteCloser(w, nil))
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

func (f *File) Seek(offset int64, whence int) (int64, error) {
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

func (f *File) Stat() (os.FileInfo, error) {
	return f.Info, nil
}

type fiStream struct {
	isClosed bool
	limit    bool
	have     int
	buf      []os.FileInfo
	done     chan struct{}
	err      error
}

func (s *fiStream) Close() error {
	if !s.isClosed {
		s.isClosed = true
		s.err = io.ErrUnexpectedEOF
		s.done <- struct{}{}
	}
	return nil
}

func (s *fiStream) Done(ctx context.Context, p filesystem.Directory_Entry_Stream_done) error {
	if !s.isClosed {
		s.isClosed = true
		s.done <- struct{}{}
	}
	return nil
}

func (s *fiStream) Push(ctx context.Context, p filesystem.Directory_Entry_Stream_push) error {
	if s.isClosed {
		return s.err
	}
	entries, err := p.Args().Entries()
	if err != nil {
		return err
	}
	if s.limit && cap(s.buf)-s.have < entries.Len() {
		s.err = TooManyEntries
		s.isClosed = true
		s.done <- struct{}{}
		return s.err
	}
	for i := 0; i < entries.Len(); i++ {
		entry := entries.At(i)
		name, err := entry.Name()
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		s.buf = append(s.buf, &FileInfo{
			name: name,
			info: cloneInfo(info),
		})
		s.have++
	}
	return nil
}

func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	if !f.Info.IsDir() {
		return nil, InvalidArgument
	}
	ret := &fiStream{
		buf:  []os.FileInfo{},
		done: make(chan struct{}, 1),
	}
	if count > 0 {
		ret.buf = make([]os.FileInfo, 0, count)
		ret.limit = true
	}
	// FIXME: the remote could very easily cause us to hang here.
	filesystem.Directory{Client: f.Node.Client}.List(
		context.TODO(),
		func(p filesystem.Directory_list_Params) error {
			p.SetStream(filesystem.Directory_Entry_Stream_ServerToClient(ret, nil))
			return nil
		})
	<-ret.done
	return ret.buf, ret.err
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

func (fs *FileSystem) Open(name string) (http.File, error) {
	parts := strings.Split(name, "/")
	parts = parts[1:] // remove the empty string at the start
	if len(parts) != 0 && parts[0] == "fs" {
		// TODO(cleanup): This logic ought to go elsewhere.

		// strip off the path prefix
		parts = parts[1:]
	}
	toRelease := make([]capnp.ReleaseFunc, 0, len(parts))

	var node filesystem.Node
	var dir filesystem.Directory
	node = filesystem.Node{Client: fs.Dir.Client}
	dir.Client = node.Client

	defer func() {
		for _, release := range toRelease {
			release()
		}
	}()

	for _, nodeName := range parts {
		res, release := dir.Walk(context.TODO(), func(p filesystem.Directory_walk_Params) error {
			p.SetName(nodeName)
			return nil
		})
		node = res.Node()
		toRelease = append(toRelease, release)
		dir = filesystem.Directory{node.Client}
	}
	ret, _ := node.Stat(context.TODO(), func(p filesystem.Node_stat_Params) error {
		return nil
	})

	info, err := ret.Info().Struct()
	if err != nil {
		return nil, err
	}

	var retName string
	if len(parts) == 0 {
		retName = ""
	} else {
		retName = parts[len(parts)-1]
	}
	return &File{
		Node: node,
		Info: &FileInfo{
			name: retName,
			info: cloneInfo(info),
		},
	}, nil
}
