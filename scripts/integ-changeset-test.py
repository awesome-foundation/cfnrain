#!/usr/bin/env python3
"""Check explicit integration change set names without AWS access."""
from pathlib import Path
import re

# Shell words preserve quoted paths and variables. Commands in these scripts
# put positional arguments before flags that take values.
word = r'(?:[^\s\\"\']+|"[^"]*"|\x27[^\x27]*\x27)+'
count = 0
for path in Path("scripts").glob("integ*.sh"):
    source = path.read_text().replace("\\\n", " ")
    patterns = [
        rf'--change-set-name\s+({word})',
        rf'\./cfnrain deploy\s+(?:--no-exec\s+)?(?:--changeset\s+)?({word})\s+({word})(?:\s+({word}))?',
        rf'\./cfnrain (?:ls|rm) -c\s+({word})\s+({word})',
    ]
    for index, pattern in enumerate(patterns):
        for match in re.finditer(pattern, source):
            if index == 0:
                name = match[1]
            elif index == 2 or "--changeset" in match[0]:
                name = match[2]
            else:
                name = match[3]
            if not name or name.startswith(("-", ">", "2>", "|", ";")):
                continue
            name = name.strip("\"'")
            assert name.startswith("success"), f"{path}: invalid change set name: {name}"
            count += 1
assert count, "No explicit change set names checked"
print(f"Checked {count} explicit change set names")
