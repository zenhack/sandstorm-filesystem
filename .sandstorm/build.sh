#!/bin/bash
set -euo pipefail
# This script is run in the VM each time you run `vagrant-spk dev`.  This is
# the ideal place to invoke anything which is normally part of your app's build
# process - transforming the code in your repository into the collection of files
# which can actually run the service in production
#
# Some examples:
#
#   * For a C/C++ application, calling
#       ./configure && make && make install
#   * For a Python application, creating a virtualenv and installing
#     app-specific package dependencies:
#       virtualenv /opt/app/env
#       /opt/app/env/bin/pip install -r /opt/app/requirements.txt
#   * Building static assets from .less or .sass, or bundle and minify JS
#   * Collecting various build artifacts or assets into a deployment-ready
#     directory structure

# By default, this script does nothing.  You'll have to modify it as
# appropriate for your application.


export GOPATH=$HOME/go

[ -d $GOPATH/src/zenhack.net/go/ ] || mkdir -p $GOPATH/src/zenhack.net/go/
[ -L $GOPATH/src/zenhack.net/go/sandstorm-filesystem ] || \
	ln -s /opt/app $GOPATH/src/zenhack.net/go/sandstorm-filesystem

cd /opt/app
go get -d ./...
go build -v -i

exit 0
