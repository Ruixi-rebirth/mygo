#!/usr/bin/env bash
set -e
echo "Building to get vendorHash..."
output=$(nix build 2>&1 || true)
hash=$(echo "$output" | grep -oP 'got:\s+\Ksha256-[A-Za-z0-9+/=]+')
if [ -z "$hash" ]; then
  echo "Build succeeded, no hash update needed."
  exit 0
fi
echo "New hash: $hash"
sed -i "s|vendorHash = \".*\"|vendorHash = \"$hash\"|" flake.nix
echo "Updated flake.nix. Rebuilding..."
nix build
echo "Done."
