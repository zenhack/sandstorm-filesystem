@0xe91f231103c0780e;

using Go = import "/go.capnp";
using Util = import "/util.capnp";

$Go.package("filesystem");
$Go.import("zenhack.net/go/sandstorm-filesystem");

interface Node {
	type @0 () -> (type :Type);
	enum Type {
		dir @0;
		file @1;
	}

	canWrite @1 () -> (canWrite :Bool);
}

interface Directory extends(Node) {
  list @0 () -> (list: List(Entry));
  struct Entry {
    name @0 :Text;
    file @1 :Node;
  }

	open @1 (name :Text) -> (node :Node);
}

interface RwDirectory extends(Directory) {
  create @0 (name :Text, type :Node.Type) -> (node :Node);
	# Create a node named `name` within the directory. `type`
	# indicates what type of node to create. The returned `node`
	# always implements the writable variant of that type.

	delete @1 (name :Text);
	# Delete the node in this directory named `name`.
}

interface File extends(Node) {
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

interface RwFile extends(File) {
  write @0 (startAt :UInt64) -> (sink :Util.ByteStream);
	# Return a ByteStream that can be used to write data to the file.
	# Writing starts at offset `startAt`.

  truncate @1 (size :UInt64);
	# Truncate the file to `size` bytes.
}
