#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
#
# Ensure every source file passed as an argument carries the AGPL SPDX header.
# Idempotent: files that already declare the identifier are left untouched.
# Comment style matches repo convention:
#   .ts/.tsx/.js/.mjs/.cts/.mts  -> // ... (first line)
#   .scss/.css                   -> /* ... */ (first line)
#   .vue                         -> <!-- ... --> (first line, above <script>)
set -euo pipefail

SLASH='// SPDX-License-Identifier: AGPL-3.0-or-later'
BLOCK='/* SPDX-License-Identifier: AGPL-3.0-or-later */'
HTML='<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->'

for f in "$@"; do
  [ -f "$f" ] || continue
  grep -q 'SPDX-License-Identifier' "$f" && continue
  tmp="$(mktemp)"
  case "$f" in
    *.scss|*.css) header="$BLOCK" ;;
    *.vue)        header="$HTML" ;;
    *.ts|*.tsx|*.js|*.mjs|*.cts|*.mts) header="$SLASH" ;;
    *) rm -f "$tmp"; continue ;;
  esac
  { printf '%s\n' "$header"; cat "$f"; } > "$tmp"
  mv "$tmp" "$f"
  echo "license-header: added to $f"
done
