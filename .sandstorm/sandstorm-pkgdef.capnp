@0xfa677f714b89876c;

using Spk = import "/sandstorm/package.capnp";
using Util = import "/sandstorm/util.capnp";
# This imports:
#   $SANDSTORM_HOME/latest/usr/include/sandstorm/package.capnp
# Check out that file to see the full, documented package definition format.

const pkgdef :Spk.PackageDefinition = (
  # The package definition. Note that the spk tool looks specifically for the
  # "pkgdef" constant.

  id = "qw9ge09e8m0v22hkfm6jf6nekfwh7yz64eamsa6ddg064krza3gh",
  # Your app ID is actually its public key. The private key was placed in
  # your keyring. All updates must be signed with the same key.

  manifest = (
    # This manifest is included in your app package to tell Sandstorm
    # about your app.

    appTitle = (defaultText = "Filesystem Demo App"),

    appVersion = 0,  # Increment this for every release.

    appMarketingVersion = (defaultText = "0.0.0"),
    # Human-readable representation of appVersion. Should match the way you
    # identify versions of your app in documentation and marketing.

    actions = [
      # Define your "new document" handlers here.
      ( nounPhrase = (defaultText = "Local Filesystem"),
        command = (
          argv = ["/sandstorm-http-bridge", .portStr, .myExe, "localfs"],
          environ = .myEnviron,
        ),
      ),
      ( nounPhrase = (defaultText = "Filesystem Viewer"),
        command = (
          argv = ["/sandstorm-http-bridge", .portStr, .myExe, "httpview"],
          environ = .myEnviron,
        ),
      ),
      ( nounPhrase = (defaultText = "ZipFile uploader"),
        command = (
          argv = ["/sandstorm-http-bridge", .portStr, .myExe, "zip-uploader"],
          environ = .myEnviron,
        ),
      ),
    ],

    continueCommand = (
      argv = ["/sandstorm-http-bridge", .portStr, .myExe, "restore"],
      environ = .myEnviron,
    ),
    # This is the command called to start your app back up after it has been
    # shut down for inactivity. Here we're using the same command as for
    # starting a new instance, but you could use different commands for each
    # case.

    metadata = (
      # Data which is not needed specifically to execute the app, but is useful
      # for purposes like marketing and display.  These fields are documented at
      # https://docs.sandstorm.io/en/latest/developing/publishing-apps/#add-required-metadata
      # and (in deeper detail) in the sandstorm source code, in the Metadata section of
      # https://github.com/sandstorm-io/sandstorm/blob/master/src/sandstorm/package.capnp
      icons = (
        # Various icons to represent the app in various contexts.
        #appGrid = (svg = embed "path/to/appgrid-128x128.svg"),
        #grain = (svg = embed "path/to/grain-24x24.svg"),
        #market = (svg = embed "path/to/market-150x150.svg"),
        #marketBig = (svg = embed "path/to/market-big-300x300.svg"),
      ),

      website = "https://github.com/zenhack/sandstorm-filesystem",
      # This should be the app's main website url.

      codeUrl = "https://github.com/zenhack/sandstorm-filesystem",
      # URL of the app's source code repository, e.g. a GitHub URL.
      # Required if you specify a license requiring redistributing code, but optional otherwise.

      license = (openSource = apache2),
      # The license this package is distributed under.  See
      # https://docs.sandstorm.io/en/latest/developing/publishing-apps/#license

      categories = [],
      # A list of categories/genres to which this app belongs, sorted with best fit first.
      # See the list of categories at
      # https://docs.sandstorm.io/en/latest/developing/publishing-apps/#categories

      author = (
        # Fields relating to the author of this app.

        contactEmail = "ian@zenhack.net",
        # Email address to contact for any issues with this app. This includes end-user support
        # requests as well as app store administrator requests, so it is very important that this be a
        # valid address with someone paying attention to it.

        pgpSignature = embed "pgp-signature",
        # PGP signature attesting responsibility for the app ID. This is a binary-format detached
        # signature of the following ASCII message (not including the quotes, no newlines, and
        # replacing <app-id> with the standard base-32 text format of the app's ID):
        #
        # "I am the author of the Sandstorm.io app with the following ID: <app-id>"
        #
        # You can create a signature file using `gpg` like so:
        #
        #     echo -n "I am the author of the Sandstorm.io app with the following ID: <app-id>" | gpg --sign > pgp-signature
        #
        # Further details including how to set up GPG and how to use keybase.io can be found
        # at https://docs.sandstorm.io/en/latest/developing/publishing-apps/#verify-your-identity

        # upstreamAuthor = "Example App Team",
        # Name of the original primary author of this app, if it is different from the person who
        # produced the Sandstorm package. Setting this implies that the author connected to the PGP
        # signature only "packaged" the app for Sandstorm, rather than developing the app.
        # Remove this line if you consider yourself as the author of the app.
      ),

      pgpKeyring = embed "pgp-keyring",
      # A keyring in GPG keyring format containing all public keys needed to verify PGP signatures in
      # this manifest (as of this writing, there is only one: `author.pgpSignature`).
      #
      # To generate a keyring containing just your public key, do:
      #
      #     gpg --export <key-id> > keyring
      #
      # Where `<key-id>` is a PGP key ID or email address associated with the key.

      #description = (defaultText = embed "path/to/description.md"),
      # The app's description in Github-flavored Markdown format, to be displayed e.g.
      # in an app store. Note that the Markdown is not permitted to contain HTML nor image tags (but
      # you can include a list of screenshots separately).

      shortDescription = (defaultText = "Demo filesystem sharing"),
      # A very short (one-to-three words) description of what the app does. For example,
      # "Document editor", or "Notetaking", or "Email client". This will be displayed under the app
      # title in the grid view in the app market.

      screenshots = [
        # Screenshots to use for marketing purposes.  Examples below.
        # Sizes are given in device-independent pixels, so if you took these
        # screenshots on a Retina-style high DPI screen, divide each dimension by two.

        #(width = 746, height = 795, jpeg = embed "path/to/screenshot-1.jpeg"),
        #(width = 640, height = 480, png = embed "path/to/screenshot-2.png"),
      ],
      #changeLog = (defaultText = embed "path/to/sandstorm-specific/changelog.md"),
      # Documents the history of changes in Github-flavored markdown format (with the same restrictions
      # as govern `description`). We recommend formatting this with an H1 heading for each version
      # followed by a bullet list of changes.
    ),
  ),

  sourceMap = (
    # Here we defined where to look for files to copy into your package. The
    # `spk dev` command actually figures out what files your app needs
    # automatically by running it on a FUSE filesystem. So, the mappings
    # here are only to tell it where to find files that the app wants.
    searchPath = [
      ( sourcePath = "." ),   # Search this directory first.
      ( sourcePath = ".." ),  # Search the root of the repository next.
      ( sourcePath = "/",     # Then search the system root directory.
        hidePaths = [ "home", "proc", "sys",
                      "etc/passwd", "etc/hosts", "etc/host.conf",
                      "etc/nsswitch.conf", "etc/resolv.conf" ]
        # You probably don't want the app pulling files from these places,
        # so we hide them. Note that /dev, /var, and /tmp are implicitly
        # hidden because Sandstorm itself provides them.
      )
    ]
  ),

  fileList = "sandstorm-files.list",
  # `spk dev` will write a list of all the files your app uses to this file.
  # You should review it later, before shipping your app.

  alwaysInclude = [],
  # Fill this list with more names of files or directories that should be
  # included in your package, even if not listed in sandstorm-files.list.
  # Use this to force-include stuff that you know you need but which may
  # not have been detected as a dependency during `spk dev`. If you list
  # a directory here, its entire contents will be included recursively.

  bridgeConfig = (
    expectAppHooks = true,
    viewInfo = (),
  ),
);

const portStr :Text = "8000";
const myExe :Text = "/app.stripped";

const myEnviron :List(Util.KeyValue) = [
    # Note that this defines the *entire* environment seen by your app.
    (key = "PATH", value = "/usr/local/bin:/usr/bin:/bin"),
    (key = "LISTEN_PORT", value = .portStr),
  ];
