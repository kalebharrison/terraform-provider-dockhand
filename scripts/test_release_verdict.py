#!/usr/bin/env python3
"""Unit tests for release verdict parsing."""

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from release_verdict import clear_to_tag_versions, is_cleared_for_version


class ReleaseVerdictTest(unittest.TestCase):
    def test_clear_verdict_detected(self) -> None:
        text = """
## Release v0.2.0 — lens review

### Release v0.2.0 — verdict

- **Clear to tag:** yes
- **Blocking findings:** none
"""
        self.assertTrue(is_cleared_for_version(text, "0.2.0"))
        self.assertIn("0.2.0", clear_to_tag_versions(text))

    def test_no_verdict(self) -> None:
        self.assertFalse(is_cleared_for_version("no verdict here", "0.2.0"))


if __name__ == "__main__":
    unittest.main()
