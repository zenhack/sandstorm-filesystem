@0xe91f231103c0780e;

using Go = import "/go.capnp";
using Util = import "/util.capnp";

$Go.package("filesystem");
$Go.import("zenhack.net/go/sandstorm-filesystem");

# This file specifies a schema that sandstorm grains may use to share files
# and directories with one another.
#
# We follow the suggested conventions for doing requests/offers. Any of the
# interfaces in this file may be requested or offered; use the interface id
# as the tag, and leave the value null.
#
# NOTE: this is *unstable*. Backwards-incompatible changes may be made to this
# schema until we settle on a final-ish design.

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

interface Directory @0xce3039544779e0fc extends(Node) {
  # A (possibly read-only) directory.

  list @0 () -> (list: List(Entry));
  # List the contents of the directory. TODO: we probably want some way
  # to do pagination/otherwise not have to transfer all of the entries at
  # once.

  struct Entry {
    # A child of a directory.
    name @0 :Text;
    file @1 :Node;
  }

  open @1 (name :Text) -> (node :Node);
  # Open a file in this directory.
}

interface RwDirectory @0xdffe2836f5c5dffc extends(Directory) {
  # A directory, with write access.

  create @0 (name :Text, type :Node.Type) -> (node :Node);
  # Create a node named `name` within the directory. `type`
  # indicates what type of node to create. The returned `node`
  # always implements the writable variant of that type.

  delete @1 (name :Text);
  # Delete the node in this directory named `name`.
}

interface File @0xaa5b133d60884bbd extends(Node) {
  # A regular file

  size @0 () -> (size: UInt64);
  # Return the size of the file.

  read @1 (startAt :UInt64, amount :UInt64, sink :Util.ByteStream) ->
    (handle :Util.Handle);
  # Read `amount` bytes from the file into `sink`, starting at position
  # `startAt`. As a special case, if `amount` is 0, data will be read
  # until the end of the file is reached.
  #
  # Dropping the returned handle can be used to request that the transfer
  # be canceled.
}

interface RwFile @0xb4810121539f6e53 extends(File) {
  # A file, with write access.

  write @0 (startAt :UInt64) -> (sink :Util.ByteStream);
  # Return a ByteStream that can be used to write data to the file.
  # Writing starts at offset `startAt`.

  truncate @1 (size :UInt64);
  # Truncate the file to `size` bytes.
}
