#!/bin/bash

if [ -f "$HOME/.bashrc" ]; then
    . "$HOME/.bashrc"
fi

. /opt/nvm/nvm.sh

exec "$@"
