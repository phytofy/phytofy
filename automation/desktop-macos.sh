#!/bin/sh

curl -L "https://github.com/$GH_REPOSITORY/releases/download/$RELEASE_VERSION/phytofy-cli.macos-amd64.$RELEASE_VERSION" --output desktop/phytofy-cli.exe
chmod +x desktop/phytofy-cli.exe

npm i -g electron@11.1.1 electron-builder@22.9.1
cp ui/public/img/icons/icon512.png desktop/icon.png
sed -i '' "s/0\.0\.0/$RELEASE_VERSION/" desktop/package.json

cd desktop
npm ci
npm run package-macos

brew install gh

echo $GH_API_TOKEN | gh auth login --hostname github.com --with-token
gh release upload $RELEASE_VERSION "release/OSRAM - PHYTOFY RL v1 UI-$RELEASE_VERSION.dmg"
