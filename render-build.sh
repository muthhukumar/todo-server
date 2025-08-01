#!/usr/bin/env bash
# exit on error
set -o errexit

# STORAGE_DIR=/opt/render/project/go/src/github.com/muthhukumar/todo-server/.render

# if [[ ! -d $STORAGE_DIR/chrome ]]; then
#   echo "...Downloading Chrome"
#   mkdir -p $STORAGE_DIR/chrome
#   cd $STORAGE_DIR/chrome
#   wget -P ./ https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
#   dpkg -x ./google-chrome-stable_current_amd64.deb $STORAGE_DIR/chrome
#   rm ./google-chrome-stable_current_amd64.deb
#   cd $HOME/project/go/src/github.com/muthhukumar/todo-server
# else
#   echo "...Using Chrome from cache"
# fi

go build ./cmd/todo-server/main.go
