#!/usr/bin/env python3
"""
Verify docs/examples parity:
- every docs/resources/*.md has examples/resources/dockhand_<name>/resource.tf
- every docs/data-sources/*.md has examples/data-sources/dockhand_<name>/data-source.tf
"""

from __future__ import annotations

import sys
from pathlib import Path


def stem_names(path: Path) -> list[str]:
    return sorted(p.stem for p in path.glob("*.md") if p.is_file())


def main() -> int:
    root = Path(__file__).resolve().parent.parent

    resources = stem_names(root / "docs" / "resources")
    data_sources = stem_names(root / "docs" / "data-sources")

    missing_resource_examples: list[str] = []
    for name in resources:
        expected = root / "examples" / "resources" / f"dockhand_{name}" / "resource.tf"
        if not expected.exists():
            missing_resource_examples.append(str(expected.relative_to(root)))

    missing_data_source_examples: list[str] = []
    for name in data_sources:
        expected = root / "examples" / "data-sources" / f"dockhand_{name}" / "data-source.tf"
        if not expected.exists():
            missing_data_source_examples.append(str(expected.relative_to(root)))

    if not missing_resource_examples and not missing_data_source_examples:
        print("docs/example coverage check passed")
        return 0

    print("docs/example coverage check failed:")
    if missing_resource_examples:
        print("- Missing resource examples:")
        for p in missing_resource_examples:
            print(f"  - {p}")
    if missing_data_source_examples:
        print("- Missing data source examples:")
        for p in missing_data_source_examples:
            print(f"  - {p}")

    return 1


if __name__ == "__main__":
    sys.exit(main())
