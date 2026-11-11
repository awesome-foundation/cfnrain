#!/usr/bin/env python3
"""Check the actual integration command lines without executing their AWS operations."""

import itertools
import json
import os
from pathlib import Path
import re
import shlex
import subprocess
import sys

binary = str(Path("cfnrain").resolve())
env = dict(os.environ, AWS_EC2_METADATA_DISABLED="true")
paths = sys.argv[1:] or ["scripts/integ-p1.sh", "scripts/integ-p1-slow.sh"]
checked = 0

for path in paths:
    # These scripts use simple command lines. Do not eval shell source: extract
    # argv, discard redirections, and supply inert values for runtime variables.
    source = Path(path).read_text().replace("\\\n", " ")
    commands = 0
    for number, line in enumerate(source.splitlines(), 1):
        words = shlex.split(line, comments=True)
        if "./cfnrain" not in words:
            continue
        words = words[words.index("./cfnrain") + 1:]
        end = next((i for i, word in enumerate(words)
                    if re.match(r"^\d*[<>]", word)), len(words))
        words = words[:end]
        commands += 1
        modes = ("managed", "retained") if any("$mode" in w for w in words) else ("managed",)
        flags = ([], ["--keep"]) if "${flags[@]}" in words else ([],)
        for mode, extra in itertools.product(modes, flags):
            argv = []
            for word in words:
                if word == "${flags[@]}":
                    argv.extend(extra)
                else:
                    word = word.replace("$mode", mode)
                    argv.append(re.sub(r"\$\{[^}]+\}|\$[A-Za-z_][A-Za-z_0-9]*",
                                       "fixture", word))
            result = subprocess.run([binary, *argv, "--help"], env=env,
                                    capture_output=True, text=True, timeout=10)
            if result.returncode:
                sys.exit(f"{path}:{number}: flag parsing failed: {result.stderr}")

            # Help validates flag names, not deploy's Metadata.Rain precondition.
            # Inspect each checked-in deploy fixture with the real formatter too.
            if argv[0] == "deploy":
                for argument in argv[1:]:
                    fixture = Path(argument)
                    if fixture.suffix != ".yaml" or not fixture.is_file():
                        continue
                    template = json.loads(subprocess.check_output(
                        [binary, "fmt", "--json", str(fixture)], env=env, timeout=10))
                    has_rain = any("Rain" in r.get("Metadata", {})
                                   for r in template.get("Resources", {}).values())
                    if has_rain and "--experimental" not in argv:
                        sys.exit(f"{path}:{number}: Metadata.Rain deploy needs "
                                 "--experimental; deploy -x means --no-exec")
            checked += 1
    if not commands:
        sys.exit(f"{path}: no cfnrain commands checked")

print(f"P1 CLI flag checks passed ({checked} command variants)")
