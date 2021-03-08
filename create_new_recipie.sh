#!/usr/bin/env sh

mkdir -p recipe "$1"

cat >> recipes/"$1"/recipe.yml << EOF
Title: GIVE ME A TITLE
Description: GIVE ME A DESCRIPTION
Author: WHAT IS YOUR NAME
Ingredients:
  - This
  - is
  - a
  - list
Instructions:
  - so
  - is
  - this
EOF