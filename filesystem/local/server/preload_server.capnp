@0xfa27741246fb0c93;

using Go = import "/go.capnp";

$Go.package("main");
$Go.import("zenhack.net/go/sandstorm-filesystem/filesystem/local/server");

using FileSystem = import "/filesystem.capnp";

interface Bootstrap {
  rootfs @0 () -> (dir :FileSystem.RwDirectory);
}
