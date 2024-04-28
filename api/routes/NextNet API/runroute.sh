#!/usr/bin/env bash
npx @usebruno/cli run "$1" --output /tmp/out.json
cat /tmp/out.json | less