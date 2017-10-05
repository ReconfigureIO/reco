# reco

The Reconfigure.io tool.

## Installation
To install from source, check [Install from source](#installation-from-source) below.

### Linux
Paste into shell

```sh
curl -LO https://s3.amazonaws.com/reconfigure.io/reco/releases/reco-master-x86_64-linux.zip \
&& unzip -o reco-master-x86_64-linux.zip \
&& sudo mv reco /usr/local/bin
```

### OSX
Paste into shell

```sh
curl -LO https://s3.amazonaws.com/reconfigure.io/reco/releases/reco-master-x86_64-apple-darwin.zip \
&& unzip -o reco-master-x86_64-apple-darwin.zip \
&& sudo mv reco /usr/local/bin
```

### Windows
Launch Powershell as **administrator** and paste

```powershell
Invoke-WebRequest https://s3.amazonaws.com/reconfigure.io/reco/releases/reco-master-x86_64-pc-windows.zip -OutFile reco-master-x86_64-pc-windows.zip;
Expand-Archive -Path reco-master-x86_64-pc-windows.zip -DestinationPath C:\reco;
setx PATH "$env:path;C:\reco" -m;
```

reco will be available in further sessions of cmd or powershell

## Usage
```
reco is the CLI utility for Reconfigure.io

Usage:
  reco [command]

Development Commands:
  build       Manage builds
  check       Verify that the compiler can build your Go source
  deployment  Manage deployments
  graph       Manage graphs
  test        Manage tests

Other Commands:
  auth        Authenticate your account
  completion  Generate bash completion
  project     Manage projects
  version     Show app version

Flags:
  -h, --help             help for reco
      --project string   project to use. If unset, the active project is used
  -s, --source string    source directory (default is current directory ".")

Use "reco [command] --help" for more information about a command.
```

## Hidden Flags and Commands
The following flags are meant for internal use and thereby hidden.

### Commands
`reco config` shows path to the default configuration file

### Flags
`--config` specify a config file


## Configuration
The configuration file is optional and reco does not require it to work.

Environment Variables or a `reco.yml` with the following values set. You can get the path to the `reco.yml` by running `reco config`.

```
# optional
PLATFORM_SERVER # defaults to "https://api.reconfigure.io"
```

## Installation from source

Requires Go 1.6+.

### 1. Export `GOPATH` if not set set. You can verify with `go env GOPATH`.
```sh
export GOPATH=$HOME/go
```

### 2. Clone the repo into `GOPATH`
```sh
git clone https://github.com/ReconfigureIO/reco $GOPATH/src/github.com/ReconfigureIO/reco
```

### 3. Install dependencies

Requires Glide, install if required.
```sh
glide -version || curl https://glide.sh/get | sh
```

cd into the source directory and install dependencies
```sh
cd $GOPATH/src/github.com/ReconfigureIO/reco
glide install
```

### 4. Build reco
```sh
make install
```
`reco` should be available in your $PATH afterwards.
