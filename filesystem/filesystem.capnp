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

# General notes:
#
# * All file names that appear in the API are *single* path segments --
#   implementations of these interfaces must not interpret multi-part paths,
#   and must throw exceptions when given filenames containing slashes.
# * The filenames "", "." and ".." are illegal, and implementations must
#   throw an exception if given one of these. Note that ".." in particular
#   is very capability-unfriendly; traversing to a parent directory is very
#   purposfully not facilitated by the API.

using Util = import "/util.capnp";

# Note to non go users: you can just delete this if you don't want to
# install the go plugin:
using Go = import "/go.capnp";
$Go.package("filesystem");
$Go.import("zenhack.net/go/sandstorm-filesystem/filesystem");

interface Node @0x955400781a01b061 {
  # A node in the filesystem. This is either a file or a directory.

  stat @0 () -> (info :StatInfo);
  # Report information about the node.
}

struct StatInfo {
  union {
    dir @0 :Void;
    file :group {
      size @1 :Int64;
    }
  }
  executable @2 :Bool;
  writable @3 :Bool;
}

interface Directory @0xce3039544779e0fc extends(Node) {
  # A (possibly read-only) directory.

  list @0 (stream :Entry.Stream);
  # List the contents of the directory. Entries are pushed into `stream`.
  #
  # `list` will not return until all of the entries have been pushed into
  # `stream` (or an error has occurred).

  struct Entry {
    # Information about a child of a directory.
    name @0 :Text;
    info @1 :StatInfo;

    interface Stream {
      # A stream of directories, for use with `list`, above.
      push @0 (entries :List(Entry));
      done @1 ();
    }
  }

  walk @1 (name :Text) -> (node :Node);
  # Open a file in this directory.
}

interface RwDirectory @0xdffe2836f5c5dffc extends(Directory) {
  # A directory, with write access.

  create @0 (name :Text, executable :Bool) -> (file :RwFile);
  # Create a file in the current directory.

  mkdir @1 (name :Text) -> (dir :RwDirectory);
  # Create a sub-directory in the current directory.

  delete @2 (name :Text);
  # Delete the node in this directory named `name`. If it is a directory,
  # it must be empty.
}

interface File @0xaa5b133d60884bbd extends(Node) {
  # A regular file

  read @0 (startAt :Int64, amount :UInt64, sink :Util.ByteStream);
  # Read `amount` bytes from the file into `sink`, starting at position
  # `startAt`. As a special case, if `amount` is 0, data will be read
  # until the end of the file is reached.
  #
  # If there are fewer than `amount` bytes, available, data will be read
  # until the end of the file.
  #
  # `read` will not return until all of the data has been written into
  # `sink` (or an error has occurred).
}

interface RwFile @0xb4810121539f6e53 extends(File) {
  # A file, with write access.

  write @0 (startAt :Int64) -> (sink :Util.ByteStream);
  # Return a ByteStream that can be used to write data to the file.
  # Writing starts at offset `startAt`. `-1` denotes the end of the file.

  truncate @1 (size :UInt64);
  # Truncate the file to `size` bytes.

  setExec @2 (exec :Bool);
  # Set the executable bit to `exec`.
}

# vim: set ts=2 sw=2 et :
