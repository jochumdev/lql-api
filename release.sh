#!/usr/bin/env bash

git commit -a -S -m "$1"
chglog add --version $2
git commit -a -S -m $2
git tag $2
git push
git push --tags