#!/usr/bin/env bash

git commit -a -S -m "$1"
chglog add --version $2
chglog format -i ./changelog.yml -o CHANGELOG -t repo
git commit -a -S -m $2
git tag $2
git push
git push --tags