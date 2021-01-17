#!/usr/bin/env bash

DATE=$(date "+%Y-%m-%dT%H:%M:%S%z")
TITLE=$(echo $1)
FILENAME=$(echo "$TITLE.md" | tr " " "-" | tr '[:upper:]' '[:lower:]')
hugo new content/items/$FILENAME
cat > content/items/$FILENAME <<EOF
---
title: "$TITLE"
date: $DATE
itemurl: "$2"
sites: "$(echo $(echo "$2" | awk -F/ '{print $3}'))"
tags: $3
draft: false
---
EOF
