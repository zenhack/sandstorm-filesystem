@0xf7eee63280552269;

using Go = import "/go.capnp";
$Go.package("main");
$Go.import("zenhack.net/go/sandstorm-filesystem");

using Powerbox = import "/powerbox.capnp";

const directoryReq :Powerbox.PowerboxDescriptor = (
	tags = [(
		id = 14857438204834472188, # Id for Directory
	)],
);

const rwDirectoryReq :Powerbox.PowerboxDescriptor = (
	tags = [(
		id = 16140382331059167228, # Id for RwDirectory
	)],
);
