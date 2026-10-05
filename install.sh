#!/bin/sh
set -eu

REPOSITORY="Dani-Bhai/dev-doctor"
VERSION=""
INSTALL_DIR=${INSTALL_DIR:-"${HOME:?HOME must be set}/.local/bin"}
FROM_BUILD=0

usage() {
	cat <<'EOF'
Install Dev Doctor from the latest GitHub release.

Usage: install.sh [--version TAG] [--install-dir DIR] [--local] [--help]

Options:
  --version TAG     Install a specific GitHub release tag (for example v0.1.0)
  --install-dir DIR Install into DIR (default: ~/.local/bin)
  --local           Install dist/devdoctor from this source checkout
  --help            Show this help
EOF
}

fail() {
	printf 'install.sh: %s\n' "$1" >&2
	exit 1
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--version)
		[ "$#" -ge 2 ] || fail "--version requires a tag"
		VERSION=$2
		shift 2
		;;
	--install-dir)
		[ "$#" -ge 2 ] || fail "--install-dir requires a directory"
		INSTALL_DIR=$2
		shift 2
		;;
	--local)
		FROM_BUILD=1
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		fail "unknown option: $1"
		;;
	esac
done

if [ -z "$INSTALL_DIR" ]; then
	fail "installation directory cannot be empty"
fi

temp_dir=$(mktemp -d "${TMPDIR:-/tmp}/devdoctor-install.XXXXXX") || fail "could not create a temporary directory"
trap 'rm -rf "$temp_dir"' EXIT HUP INT TERM

if [ "$FROM_BUILD" -eq 1 ]; then
	script_dir=$(CDPATH= cd "$(dirname "$0")" && pwd)
	[ -f "$script_dir/dist/devdoctor" ] || fail "dist/devdoctor is missing; run 'make build' first"
	cp "$script_dir/dist/devdoctor" "$temp_dir/devdoctor"
else
	case "$VERSION" in
	'') release_url="https://github.com/$REPOSITORY/releases/latest/download" ;;
	*[!A-Za-z0-9._-]*) fail "release tag contains unsupported characters" ;;
	*) release_url="https://github.com/$REPOSITORY/releases/download/$VERSION" ;;
	esac

	case "$(uname -s)" in
	Linux) os=linux ;;
	*) fail "Dev Doctor currently supports Linux only" ;;
	esac

	case "$(uname -m)" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) fail "unsupported CPU architecture: $(uname -m)" ;;
	esac

	asset="devdoctor_${os}_${arch}.tar.gz"
	if command -v curl >/dev/null 2>&1; then
		download() { curl -fsSL "$1" -o "$2"; }
	elif command -v wget >/dev/null 2>&1; then
		download() { wget -q "$1" -O "$2"; }
	else
		fail "curl or wget is required to download a release"
	fi

	command -v sha256sum >/dev/null 2>&1 || fail "sha256sum is required to verify the release download"
	download "$release_url/$asset" "$temp_dir/$asset" || fail "could not download $asset from $release_url"
	download "$release_url/SHA256SUMS" "$temp_dir/SHA256SUMS" || fail "could not download release checksums"

	expected=$(awk -v name="$asset" '$2 == name { print $1; exit }' "$temp_dir/SHA256SUMS")
	[ -n "$expected" ] || fail "no checksum found for $asset"
	actual=$(sha256sum "$temp_dir/$asset" | awk '{ print $1 }')
	[ "$actual" = "$expected" ] || fail "checksum verification failed for $asset"
	tar -xzf "$temp_dir/$asset" -C "$temp_dir" devdoctor || fail "could not unpack the release archive"
fi

mkdir -p "$INSTALL_DIR"
install -m 0755 "$temp_dir/devdoctor" "$INSTALL_DIR/devdoctor"

if [ "$INSTALL_DIR" = "${HOME}/.local/bin" ]; then
	add_path_config() {
		profile=$1
		if [ ! -f "$profile" ] || ! grep -Fq '# devdoctor PATH' "$profile"; then
			printf '\n# devdoctor PATH\ncase ":$PATH:" in\n  *":$HOME/.local/bin:"*) ;;\n  *) PATH="$HOME/.local/bin:$PATH" ;;\nesac\nexport PATH\n' >> "$profile"
		fi
	}

	add_path_config "$HOME/.profile"
	shell_name=${SHELL:-/bin/sh}
	shell_name=${shell_name##*/}
	case "$shell_name" in
	bash) add_path_config "$HOME/.bashrc" ;;
	zsh)
		add_path_config "$HOME/.zprofile"
		add_path_config "$HOME/.zshrc"
		;;
	fish)
		mkdir -p "$HOME/.config/fish/conf.d"
		cat > "$HOME/.config/fish/conf.d/devdoctor-path.fish" <<'EOF'
if not contains -- "$HOME/.local/bin" $PATH
    set -gx PATH "$HOME/.local/bin" $PATH
end
EOF
		;;
	esac
fi

printf 'Installed Dev Doctor at %s/devdoctor\n' "$INSTALL_DIR"
case ":$PATH:" in
*":$INSTALL_DIR:"*) printf 'Run it with: devdoctor\n' ;;
*) printf 'Open a new terminal or run: export PATH="%s:$PATH"\n' "$INSTALL_DIR" ;;
esac
