#!/usr/bin/env bash
# Install run-skill-script from GitHub Releases into ~/.local/bin by default.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/towry/run-skill-script/main/install.sh | bash
#   RUN_SKILL_SCRIPT_VERSION=v0.1.0 bash install.sh
#   PREFIX=/usr/local bash install.sh
set -euo pipefail

REPO="${RUN_SKILL_SCRIPT_REPO:-towry/run-skill-script}"
PREFIX="${PREFIX:-${HOME}/.local}"
BIN_DIR="${RUN_SKILL_SCRIPT_BIN_DIR:-${PREFIX}/bin}"
VERSION="${RUN_SKILL_SCRIPT_VERSION:-}"
BINARY_NAME="run-skill-script"

die() {
	echo "install.sh: $*" >&2
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || die "need $1"
}

os_name() {
	case "$(uname -s)" in
	Linux) echo linux ;;
	Darwin) echo darwin ;;
	*) die "unsupported OS: $(uname -s). Supported: linux, darwin." ;;
	esac
}

arch_name() {
	case "$(uname -m)" in
	x86_64 | amd64) echo amd64 ;;
	arm64 | aarch64) echo arm64 ;;
	*) die "unsupported architecture: $(uname -m). Supported: amd64, arm64." ;;
	esac
}

latest_tag() {
	local url="https://api.github.com/repos/${REPO}/releases/latest"
	local json
	json="$(curl -fsSL "${url}")" || die "failed to fetch ${url}"
	local tag
	tag="$(printf '%s' "${json}" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
	[[ -n "${tag}" && "${tag}" != "null" ]] || die "no GitHub release found for ${REPO}. Publish a vX.Y.Z tag first."
	printf '%s\n' "${tag}"
}

need curl
need tar
need mktemp
need uname

OS="$(os_name)"
ARCH="$(arch_name)"
if [[ -z "${VERSION}" ]]; then
	VERSION="$(latest_tag)"
fi
case "${VERSION}" in
v*) ;;
*) VERSION="v${VERSION}" ;;
esac

ARCHIVE="${BINARY_NAME}_${OS}_${ARCH}.tar.gz"
BASE="https://github.com/${REPO}/releases/download/${VERSION}"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

echo "installing ${BINARY_NAME} ${VERSION} (${OS}/${ARCH}) to ${BIN_DIR}"
curl -fsSL "${BASE}/${ARCHIVE}" -o "${TMP}/${ARCHIVE}"
curl -fsSL "${BASE}/checksums.txt" -o "${TMP}/checksums.txt"

if command -v sha256sum >/dev/null 2>&1; then
	(cd "${TMP}" && sha256sum -c --ignore-missing checksums.txt)
elif command -v shasum >/dev/null 2>&1; then
	want="$(awk -v f="${ARCHIVE}" '$2 == f { print $1; exit }' "${TMP}/checksums.txt")"
	[[ -n "${want}" ]] || die "checksums.txt has no entry for ${ARCHIVE}"
	got="$(shasum -a 256 "${TMP}/${ARCHIVE}" | awk '{ print $1 }')"
	[[ "${got}" == "${want}" ]] || die "checksum mismatch for ${ARCHIVE}"
else
	die "need sha256sum or shasum to verify the download"
fi

tar -xzf "${TMP}/${ARCHIVE}" -C "${TMP}"
[[ -f "${TMP}/${BINARY_NAME}" ]] || die "archive ${ARCHIVE} did not contain ${BINARY_NAME}"

mkdir -p "${BIN_DIR}"
install -m 0755 "${TMP}/${BINARY_NAME}" "${BIN_DIR}/${BINARY_NAME}"

echo "installed ${BIN_DIR}/${BINARY_NAME}"
"${BIN_DIR}/${BINARY_NAME}" --version

case ":${PATH}:" in
*:"${BIN_DIR}":*) ;;
*)
	echo "add ${BIN_DIR} to PATH, for example:"
	echo "  export PATH=\"${BIN_DIR}:\$PATH\""
	;;
esac
