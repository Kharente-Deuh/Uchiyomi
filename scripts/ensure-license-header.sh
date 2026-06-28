#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
#
# Ensure every source file passed as an argument carries the AGPL SPDX header.
# Idempotent: files that already declare the identifier are left untouched.
# Comment style matches repo convention:
#   .ts/.tsx/.js/.mjs/.cts/.mts  -> // ... followed by a blank line (first lines)
#   .scss/.css                   -> /* ... */ (first line)
#   .vue                         -> <!-- ... --> (first line, above <script>)
#
# JS/TS files get a blank line between the header and the code so the import
# sorter (perfectionist) treats the header as a detached top-of-file comment
# and never hoists an import above it.
set -euo pipefail

SLASH='// SPDX-License-Identifier: AGPL-3.0-or-later'
BLOCK='/* SPDX-License-Identifier: AGPL-3.0-or-later */'
HTML='<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->'

for f in "$@"; do
  [ -f "$f" ] || continue
  grep -q 'SPDX-License-Identifier' "$f" && continue
  tmp="$(mktemp)"
  case "$f" in
    *.scss|*.css)
      { printf '%s\n' "$BLOCK"; cat "$f"; } > "$tmp" ;;
    *.vue)
      { printf '%s\n' "$HTML"; cat "$f"; } > "$tmp" ;;
    *.ts|*.tsx|*.js|*.mjs|*.cts|*.mts)
      # header, one blank line, then the file with any leading blanks dropped
      awk -v hdr="$SLASH" 'BEGIN { print hdr; print "" }
        !started && $0 ~ /^[[:space:]]*$/ { next }
        { started = 1; print }' "$f" > "$tmp" ;;
    *)
      rm -f "$tmp"; continue ;;
  esac
  mv "$tmp" "$f"
  echo "license-header: added to $f"
done
