#!/usr/bin/env python3
"""Check explicit integration change set names without executing shell code."""
from pathlib import Path
import re
import shlex
import sys

# Unsupported indirect forms: "$bin" deploy, command arrays, --cli-input-json,
# and names populated by read. Keep integration change set commands explicit.
scripts = Path(__file__).resolve().parent
paths = [Path(p) for p in sys.argv[1:]] or sorted(scripts.glob("integ*.sh"))
variable = re.compile(r"^\$(?:\{(\w+)\}|(\w+))")
boolean = set("--no-exec -x -y --yes -d --detach --keep --experimental --express -t --termination-protection".split())
valued = set("--params --tags --s3-prefix --node-style --role-arn --config --region --profile".split())


def names(line):
    # Only explicit continuations are joined by the caller. Never consume the
    # first word of the next independent command as a positional argument.
    substitution = r"\$\(([^()]*)\)"
    for nested in re.findall(substitution, line):
        yield from names(nested)
    line = re.sub(substitution, "command-output", line)
    words = shlex.split(line, comments=True)
    for i, word in enumerate(words):
        if word.startswith("--change-set-name="):
            yield word.split("=", 1)[1]
        elif word == "--change-set-name":
            if i + 1 == len(words):
                raise ValueError("missing --change-set-name value")
            yield words[i + 1]
    for i, word in enumerate(words):
        if not (word.rsplit("/", 1)[-1] == "cfnrain" or
                re.fullmatch(r"\$\{?CFNRAIN\}?", word)):
            continue
        if i and words[i - 1] in ("build", "-o"):
            continue
        args = words[i + 1:]
        if not args:
            continue
        command = args.pop(0)
        markers = ("--changeset",) if command == "deploy" else ("-c", "--changeset")
        marked = any(a.split("=")[0] in markers for a in args)
        if command not in ("deploy", "rm", "ls"):
            if marked and command != "cat":
                raise ValueError("unrecognized change set command")
            continue
        if command != "deploy" and not marked:
            continue
        positional = []
        j = 0
        while j < len(args):
            arg = args[j]
            if re.match(r"^\d*[<>]", arg) or arg in (";", "&&", "||", "|"):
                break
            if arg.split("=")[0] in markers:
                if "=" in arg and arg.split("=", 1)[1] != "true":
                    raise ValueError("unsupported changeset flag value")
                marked = True
            elif arg in boolean or arg in ("$"+"{flags[@]}", "$"+"{validation_flags[@]}"):
                pass
            elif arg.split("=")[0] in valued or (command == "deploy" and arg == "-c"):
                if "=" not in arg:
                    j += 1
                    if j == len(args):
                        raise ValueError("missing option value")
            elif arg.startswith("-"):
                raise ValueError(f"unrecognized option {arg}")
            else:
                positional.append(arg)
            j += 1
        expected = 2 if marked else 3
        if len(positional) > expected or (marked and len(positional) != expected):
            raise ValueError(f"cannot parse change set arguments: {positional}")
        if len(positional) == expected:
            yield positional[-1]


# Derive safe prefixes from assignments, including shared sourced scripts.
prefixes = set()
sources = [p.read_text().replace("\\\n", " ") for p in paths]
assignments = []
for source in sources:
    for line in source.splitlines():
        match = re.match(r'^\s*(\w+)=(.*)$', line)
        if match:
            try:
                value = shlex.split(match[2], comments=True)
            except ValueError:
                continue
            if len(value) == 1:
                assignments.append((match[1], value[0]))


def safe(name):
    match = variable.match(name)
    return name.startswith("success") or bool(
        match and (match[1] or match[2]) in prefixes)


for _ in assignments:
    for key, value in assignments:
        values = [v for k, v in assignments if k == key]
        if all(safe(v) for v in values):
            prefixes.add(key)

count = 0
for path, source in zip(paths, sources):
    for number, line in enumerate(source.splitlines(), 1):
        # Embedded Python and heredocs are not shell commands.
        if not re.search(r'cfnrain|CFNRAIN|--change-set-name', line):
            continue
        try:
            for name in names(line):
                if not safe(name):
                    raise ValueError(f"invalid or unknown change set name: {name}")
                count += 1
        except ValueError as error:
            sys.exit(f"{path}:{number}: {error}")
if not count:
    sys.exit("No explicit change set names checked")
print(f"Checked {count} explicit change set names")
