@0xb2da01084b8a5f1f;

# This file is a hack to work around an issue with the go capnp package; I hope
# to find a proper solution soon.
#
# In particular, in order to export something that satisifies more than one
# interface, we need to have a *_ServerToClient function for something that
# is a subtype of both.
#
# This file provides stubs for some of those.

using Go = import "/go.capnp";
$Go.package("filesystem");
$Go.import("zenhack.net/go/sandstorm-filesystem/filesystem");

using Filesystem = import "filesystem.capnp";
using Grain = import "/grain.capnp";

interface PersistentNode extends (Grain.AppPersistent, Filesystem.Node) {}
interface PersistentFile extends (Grain.AppPersistent, Filesystem.File) {}
interface PersistentRwFile extends (Grain.AppPersistent, Filesystem.RwFile) {}
interface PersistentDirectory extends (Grain.AppPersistent, Filesystem.Directory) {}
interface PersistentRwDirectory extends (Grain.AppPersistent, Filesystem.RwDirectory) {}
