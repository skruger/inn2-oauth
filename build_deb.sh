#!/usr/bin/env bash
set -euo pipefail

# build_deb.sh - compile inn2-oauth and produce a Debian package that depends on `inn2`
# Usage: ./build_deb.sh [version]  (version defaults to git describe or 0.1.0)

PKGNAME=inn2-oauth
ARCH=${ARCH:-amd64}
VERSION=${1:-$(git describe --tags --always 2>/dev/null || echo "0.1.0")}
INSTALL_PATH=/usr/local/sbin
CONFIG_SRC=conf/inn2-oauth.yaml
CONFIG_DEST=/etc/news/inn2-oauth.yaml

echo "Packaging $PKGNAME version $VERSION for $ARCH"

WORKDIR=$(mktemp -d)
PKGDIR=$WORKDIR/${PKGNAME}_${VERSION}_${ARCH}

mkdir -p "$PKGDIR/DEBIAN"
mkdir -p "$PKGDIR${INSTALL_PATH}"
mkdir -p "$(dirname "$PKGDIR$CONFIG_DEST")"

echo "Building binary..."
# Build a static linux binary for the target arch
CGO_ENABLED=0 GOOS=linux GOARCH=$ARCH go build -o "$PKGDIR${INSTALL_PATH}/" ./...

chmod 0755 "$PKGDIR${INSTALL_PATH}/${PKGNAME}"

if [ -f "$CONFIG_SRC" ]; then
  echo "Including config $CONFIG_SRC -> $CONFIG_DEST"
  cp "$CONFIG_SRC" "$PKGDIR$CONFIG_DEST"
  chmod 0644 "$PKGDIR$CONFIG_DEST"
else
  echo "Warning: config $CONFIG_SRC not found; creating empty placeholder at $CONFIG_DEST"
  mkdir -p "$(dirname "$PKGDIR$CONFIG_DEST")"
  echo "# inn2-oauth config" > "$PKGDIR$CONFIG_DEST"
  chmod 0644 "$PKGDIR$CONFIG_DEST"
fi

# Populate control file
MAINTAINER="$(git config user.name 2>/dev/null || echo "packager") <$(git config user.email 2>/dev/null || echo "packager@example.invalid")>"
cat > "$PKGDIR/DEBIAN/control" <<EOF
Package: $PKGNAME
Version: $VERSION
Section: utils
Priority: optional
Architecture: $ARCH
Maintainer: $MAINTAINER
Depends: inn2
Description: inn2-oauth - OAuth client bridge for INN2
 A small program to obtain access tokens for inn2 via OAuth.
EOF

chmod 0644 "$PKGDIR/DEBIAN/control"

echo "${CONFIG_DEST}" >> "${PKGDIR}/DEBIAN/conffiles"

echo "Building .deb package (you may be prompted for fakeroot if available)..."
if command -v fakeroot >/dev/null 2>&1; then
  fakeroot dpkg-deb --build "$PKGDIR"
else
  dpkg-deb --build "$PKGDIR"
fi

OUT_DEB="${PKGNAME}_${VERSION}_${ARCH}.deb"
mv "${PKGDIR}.deb" "$OUT_DEB" 2>/dev/null || mv "$WORKDIR/${PKGNAME}_${VERSION}_${ARCH}.deb" "$OUT_DEB" 2>/dev/null || true

if [ -f "$OUT_DEB" ]; then
  echo "Package created: $OUT_DEB"
else
  echo "Packaging failed: expected $OUT_DEB to be created" >&2
  ls -la "$WORKDIR"
  exit 1
fi

# Cleanup
rm -rf "$WORKDIR"

echo "Done."

