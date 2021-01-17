#!/usr/bin/env bash

##
# @Description: Add new page to Hugo items.
##

function show_usage() {
    printf "Usage: $0 [options [parameters]]\n"
    printf "\n"
    printf "Options:\n"
    printf " -t|--title, Provide Title\n"
    printf " -u|--url, Provide URL\n"
    printf " -g|--tags, Provide Tags\n"
    printf " -h|--help, Print help\n"

    return 0
}

if [[ $# -eq 0 ]]; then
    echo "Arguments are invalid or missing"
    show_usage
    exit 1
fi

while [ ! -z "$1" ]; do
    case $1 in
    --title | -t)
        shift
        TITLE=$1
        echo "Title: $TITLE"
        ;;
    --url | -u)
        shift
        URL=$1
        echo "URL: $URL"
        ;;
    --tags | -g)
        shift
        TAGS=($(echo $1 | tr "," "\n"))
        for i in "${TAGS[@]}"; do
            echo $i
        done
        ;;
    *)
        show_usage
        ;;
    esac
    shift
done

TAGSFMT=[

length=${#TAGS[@]}
current=0

for VALUE in "${TAGS[@]}"; do
    current=$((current + 1))
    if [[ "$current" -eq "$length" ]]; then
        TAGSFMT=$TAGSFMT\"
        TAGSFMT="$TAGSFMT$VALUE"
        TAGSFMT=$TAGSFMT\"
        TAGSFMT=$TAGSFMT]
    else
        TAGSFMT=$TAGSFMT\"
        TAGSFMT=$TAGSFMT$VALUE
        TAGSFMT=$TAGSFMT\"
        TAGSFMT=$TAGSFMT,
    fi
done

DATE=$(date "+%Y-%m-%dT%H:%M:%S%z")
FILENAME=$(echo "$TITLE.md" | tr " " "-" | tr '[:upper:]' '[:lower:]')
hugo new items/$FILENAME
cat >content/items/$FILENAME <<EOF
---
title: "$TITLE"
date: $DATE
itemurl: "$URL"
sites: "$(echo $(echo "$URL" | awk -F/ '{print $3}'))"
tags: $TAGSFMT
draft: false
---
EOF
