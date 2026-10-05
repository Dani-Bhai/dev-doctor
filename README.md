# Dev Doctor

Dev Doctor checks a Linux development environment for common issues with Node.js, Java, Python, Go, Git, Docker, disk space, memory, and swap.

## Install from a GitHub release

After a release has been published, install the latest version with:

```sh
curl -fsSL https://raw.githubusercontent.com/Dani-Bhai/dev-doctor/master/install.sh | sh
```

The installer downloads the matching Linux binary, verifies its SHA-256 checksum, and places it in `~/.local/bin`. It adds that directory to common shell startup files. Open a new terminal after installation, or run `export PATH="$HOME/.local/bin:$PATH"` in the current one.

To install a specific release, pass its tag to the installer:

```sh
curl -fsSL https://raw.githubusercontent.com/Dani-Bhai/dev-doctor/master/install.sh | sh -s -- --version v0.1.0
```

Supported release targets are Linux `amd64` and `arm64`.

## Build and install from source

Go 1.27.1 or newer is required to build from source.

```sh
make build
./install.sh --local
```

The local install places the binary in `~/.local/bin`. You can choose a different directory with `INSTALL_DIR=/path/to/bin ./install.sh --local`.

## Usage

```sh
devdoctor         # compact summary
devdoctor all     # full details
devdoctor help    # commands and usage
```

## Publishing a release

Push a version tag such as `v0.1.0`. GitHub Actions builds Linux `amd64` and `arm64` archives, generates `SHA256SUMS`, and attaches them to a GitHub release. The installer then downloads the latest release by default.
