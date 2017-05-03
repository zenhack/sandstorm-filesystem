@0xe91f231103c0780e;
# This file specifies a schema that sandstorm grains may use to share files
# and directories with one another.
#
# We follow the suggested conventions for doing requests/offers. Any of the
# interfaces in this file may be requested or offered; use the interface id
# as the tag, and leave the value null.
#
# NOTE: this is *unstable*. Backwards-incompatible changes may be made to this
# schema until we settle on a final-ish design.

using Util = import "/util.capnp";

# Note to non go users: you can just delete this if you don't want to
# install the go plugin:
using Go = import "/go.capnp";
$Go.package("filesystem");
$Go.import("zenhack.net/go/sandstorm-filesystem");

interface Node @0x955400781a01b061 {
  # A node in the filesystem. This is either a file or a directory.

  type @0 () -> (type :Type);
  # Report the type of the node.

  enum Type {
    dir @0;
    file @1;
  }

  canWrite @1 () -> (canWrite :Bool);
  # Report whether the node is writable. If it is, it must implement
  # one of the Rw* interfaces below.
}

enum Whence {
  start @0;
  end @1;
}

interface Directory @0xce3039544779e0fc extends(Node) {
  # A (possibly read-only) directory.


  list @0 (stream :Entry.Stream) -> (cancel :Util.Handle);
  # List the contents of the directory. Entries are pushed into `stream`.
  # The returned handle may be dropped to request canceling the stream.

  struct Entry {
    # Information about a child of a directory.
    name @0 :Text;
    canWrite @1 :Bool;
    node :union {
      dir @2 :Void;
      file :group {
        isExec @3 :Bool;
      }
    }

    interface Stream {
      push @0 (entries :List(Entry));
      done @1 ();
    }
  }


  walk @1 (name :Text) -> (node :Node);
  # Open a file in this directory.
}

interface RwDirectory @0xdffe2836f5c5dffc extends(Directory) {
  # A directory, with write access.

  create @0 (name :Text, isExec :Bool) -> (file :RwFile);
  # Create a file named `name` in the current directory. `isExec`
  # indicates whether the file should be executable.

  mkDir @1 (name :Text) -> (dir :RwDirectory);
  # Create a subdirectory named `name` in the current directory.

  delete @2 (name :Text);
  # Delete the node in this directory named `name`.
}

interface File @0xaa5b133d60884bbd extends(Node) {
  # A regular file

  size @0 () -> (size :UInt64);
  # Return the size of the file.

  read @1 (startAt :Int64, amount :UInt64, sink :Util.ByteStream)
    -> (cancel :Util.Handle);
  # Read `amount` bytes from the file into `sink`, starting at position
  # `startAt`. As a special case, if `amount` is 0, data will be read
  # until the end of the file is reached.
  #
  # If there are fewer than `amount` bytes, available, data will be read
  # until the end of the file.
  #
  # Dropping the returned handle can be used to request that the transfer
  # be canceled.

  isExec @2 () -> (isExec :Bool);
  # Reports whether or not the file is executable.
}

interface RwFile @0xb4810121539f6e53 extends(File) {
  # A file, with write access.

  write @0 (startAt :Int64, cancel :Util.Handle)
    -> (sink :Util.ByteStream);
  # Return a ByteStream that can be used to write data to the file.
  # Writing starts at offset `startAt`. `-1` denotes the end of the file.

  truncate @1 (size :UInt64);
  # Truncate the file to `size` bytes.

  setExec @2 (exec :Bool);
  # Set the executable bit to `exec`.
}

# vim: set ts=2 sw=2 et :
