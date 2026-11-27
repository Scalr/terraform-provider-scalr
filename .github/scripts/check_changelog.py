"""
Fails when a change adds or edits a CHANGELOG entry outside [Unreleased]. Released sections are
stamped by the release, an entry that lands there, e.g. in a merge conflict, never ships with the
release of its change.
"""

import re
import subprocess
import sys

FILE = "CHANGELOG.md"
HUNK = re.compile(r"^@@ -\S+ \+(\d+)(?:,(\d+))? @@", re.MULTILINE)
LINK_REFERENCE = re.compile(r"^\[[^\]]*\]: ")


def main(base: str):
    diff = subprocess.run(
        ["git", "diff", "--no-ext-diff", "--no-color", "-U0", f"{base}...HEAD", "--", FILE],
        capture_output=True, text=True, check=True,
    ).stdout
    changed = set()
    for start, count in HUNK.findall(diff):
        changed.update(range(int(start), int(start) + (1 if count == "" else int(count))))

    problems = []
    section = None
    with open(FILE) as file:
        for number, line in enumerate(file.read().split("\n"), 1):
            if line.startswith("## ["):
                section = line
            elif (
                number in changed and section and not section.startswith("## [Unreleased]")
                and line.strip() and not line.startswith("#") and not LINK_REFERENCE.match(line)
            ):
                problems.append(f"{FILE}:{number} is under `{section}`: {line}")
    for problem in problems:
        print(f"::error file={FILE}::{problem}")
    if problems:
        print("Move these entries to the [Unreleased] section, the released sections do not change.")
        sys.exit(1)


if __name__ == "__main__":
    main(sys.argv[1])
