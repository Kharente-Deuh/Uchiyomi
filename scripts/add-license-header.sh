#!/usr/bin/env sh
# SPDX-License-Identifier: AGPL-3.0-or-later

set -eu

spdx='SPDX-License-Identifier: AGPL-3.0-or-later'
fixed=''

for file in "$@"; do
  [ -f "$file" ] || continue

  if head -n 5 "$file" | grep -qF "$spdx"; then
    continue
  fi

  blank='yes'
  case "$file" in
    *.vue)
      header="<!-- $spdx -->"
      blank=''
      ;;
    *.css | *.scss)
      header="/* $spdx */"
      blank=''
      ;;
    *.go | *.js | *.mjs | *.cjs | *.ts | *.tsx)
      header="// $spdx"
      ;;
    *.sh)
      header="# $spdx"
      ;;
    *)
      continue
      ;;
  esac

  tmp=$(mktemp)
  rest=1

  if head -n 1 "$file" | grep -q '^#!'; then
    head -n 1 "$file" >"$tmp"
    rest=2
  else
    : >"$tmp"
  fi

  printf '%s\n' "$header" >>"$tmp"
  if [ -n "$blank" ]; then
    printf '\n' >>"$tmp"
  fi
  tail -n "+$rest" "$file" >>"$tmp"

  cat "$tmp" >"$file"
  rm -f "$tmp"
  fixed="$fixed  $file
"
done

if [ -n "$fixed" ]; then
  printf 'SPDX header added to:\n%s' "$fixed"
fi
