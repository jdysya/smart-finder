#!/bin/sh
set -e

echo "Building frontend..."
(cd client-front && pnpm build)

echo "Copying files to client/web..."
rm -rf client/web
mkdir -p client/web
cp -r client-front/out/* client/web/

echo "Done."
