#!/bin/bash
set -ex

echo "Run go unit test"
make test

echo "pre-build master go binary!"
make upload