#!/usr/bin/env bash

echo "Installing dependencies"
npm install

echo "Building the application"
npm run build

# package the application and get the name
PACKAGE=$(npm pack | tail -1)
echo "Generated package" $PACKAGE

# install the package globally so that the cli is accessible globally
sudo npm i -g $PACKAGE

# remove the generated package
rm $PACKAGE
