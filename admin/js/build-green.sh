#!/bin/sh

set -e

mkdir build-green
for f in `ls *.json` node_modules; do
	ln -s ../$f build-green/$f
done
cp -R src build-green/src
cp -R public build-green/public
cp public-green/* build-green/public/
sed 's/#fc2c38/#4caf50/' public/index.html > build-green/public/index.html
sed 's/#fc2c38/#4caf50/' public/browserconfig.xml > build-green/public/browserconfig.xml

(cd build-green; REACT_APP_GREEN=1 react-scripts build)

rm -rf build
mv build-green/build build
rm -rf build-green
