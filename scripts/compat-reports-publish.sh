#!/usr/bin/env bash
# Install compatibility report files from CI artifact dir into docs/reports/.
# Used by .github/workflows/compat-reports-sync.yml — not a maintainer gate.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="${COMPAT_DIR:-${1:-$ROOT/.compat-sync}}"
DEST="$ROOT/docs/reports"

if [[ ! -d "$SRC" ]]; then
  echo "compat-reports-publish: source dir not found: $SRC" >&2
  exit 1
fi

mkdir -p "$DEST"

copy_if_present() {
  local name="$1"
  if [[ -f "$SRC/$name" ]]; then
    cp "$SRC/$name" "$DEST/$name"
    echo "updated $DEST/$name"
  fi
}

copy_if_present endpoint-probe.md
copy_if_present endpoint-probe.csv
copy_if_present webui-api-endpoints.txt
copy_if_present docs-reference-api-endpoints.txt
copy_if_present dockhand-last-tested.json

verified_date="$(date -u +"%B %-d, %Y")"
non_present="$ROOT/docs/non-present-endpoints.md"
if [[ -f "$non_present" ]]; then
  /usr/bin/python3 - "$non_present" "$verified_date" <<'PY'
import re
import sys
from pathlib import Path

path = Path(sys.argv[1])
date = sys.argv[2]
text = path.read_text(encoding="utf-8")
text, count = re.subn(
    r"(?m)^- Date: .*$",
    f"- Date: {date} (CI Acceptance Full / Release Watch)",
    text,
    count=1,
)
text = text.replace(
    "- Re-run the probe after Dockhand upgrades and update this file when status changes.",
    "- Compatibility reports refresh via **Compat Reports Sync** on green Acceptance Full / Release Watch runs.",
    1,
)
if count:
    path.write_text(text, encoding="utf-8")
    print(f"updated {path}")
PY
fi
