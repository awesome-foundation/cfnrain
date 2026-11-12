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


def substitutions(line, start=0, nested=False):
    """Replace substitutions, checking each command in its own quote context."""
    output = []
    commands = []
    quote = None
    i = start
    while i < len(line):
        char = line[i]
        if char == "\\" and quote != "'":
            output.append(line[i:i + 2])
            i += 2
            continue
        if char == "`" and quote != "'":
            raise ValueError("unsupported backtick command substitution")
        if char == "$" and quote == "'":
            # Preserve literal dollars so shlex cannot turn them into trusted variables.
            output.append("literal-dollar-")
            i += 1
            continue
        if quote != "'" and line.startswith("$(", i):
            command, children, i = substitutions(line, i + 2, True)
            commands.extend(children)
            commands.append(command)
            output.append("command-output")
            continue
        if char in ("'", '"'):
            if quote is None:
                quote = char
            elif quote == char:
                quote = None
        elif quote is None:
            if char == "#" and (i == start or line[i - 1].isspace()):
                if nested:
                    raise ValueError("cannot parse comment in command substitution")
                output.append(line[i:])
                break
            if line.startswith(("<(", ">("), i):
                raise ValueError("unsupported process substitution")
            if char == ")" and nested:
                return "".join(output), commands, i + 1
            if char == "(" and nested:
                raise ValueError("unsupported parentheses in command substitution")
        output.append(char)
        i += 1
    if nested:
        raise ValueError("unclosed command substitution")
    return "".join(output), commands, i


def without_redirections(line):
    # Match shell words before shlex removes the quotes around literal > or <.
    word = r"(?:[^\s<>;&|\\\"']|\\.|\"(?:[^\"\\]|\\.)*\"|'[^']*')+"
    token = re.compile(r"(?P<redirect>(?:[0-9]+|&)?(?:>>?|<<<|<<-?|<>|<)(?:[&|])?)|" + word)
    output = []
    i = 0
    while i < len(line):
        if line[i] == "#" and (i == 0 or line[i - 1].isspace()):
            output.append(line[i:])
            break
        match = token.match(line, i)
        if not match:
            output.append(line[i])
            i += 1
            continue
        if match.group("redirect"):
            i = match.end()
            while i < len(line) and line[i].isspace():
                i += 1
            target = re.match(word, line[i:])
            if not target or line[i] == "#":
                raise ValueError("missing or unsupported redirection target")
            i += target.end()
            output.append(" ")
        else:
            output.append(match[0])
            i = match.end()
    return "".join(output)


def names(line):
    # Only explicit continuations are joined by the caller. Never consume the
    # first word of the next independent command as a positional argument.
    line, commands, _ = substitutions(line)
    for command in commands:
        yield from names(command)
    line = without_redirections(line)
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
                re.fullmatch(r"\$(?:CFNRAIN|rain|\{(?:CFNRAIN|rain)\})", word)):
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
            if arg in (";", "&&", "||", "|"):
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


def self_test():
    cases = [
        ('./cfnrain deploy t s bad --no-exec -y', ['bad']),
        ('./cfnrain deploy --changeset=true s bad', ['bad']),
        ('./cfnrain rm -y -c s bad', ['bad']),
        ('./cfnrain ls --changeset s bad', ['bad']),
        ('"$CFNRAIN" deploy t s bad', ['bad']),
        ('timeout 5 ./cfnrain deploy t s bad', ['bad']),
        ('./cfnrain deploy t s -y', []),
        ('aws cloudformation describe-change-set --change-set-name=bad', ['bad']),
        ('aws cloudformation describe-change-set --change-set-name "bad"', ['bad']),
        ('./cfnrain deploy t s "${STACK_PREFIX}cs"', ['${STACK_PREFIX}cs']),
        ('./cfnrain deploy t s "$cs"', ['$cs']),
        ('"$rain" deploy t s bad', ['bad']),
        ('"${rain}" deploy --changeset s bad', ['bad']),
        ("./cfnrain deploy t s '$STACK_PREFIX'", ['literal-dollar-STACK_PREFIX']),
        ("./cfnrain deploy t s success-ok --params '`literal`'", ['success-ok']),
        ('./cfnrain deploy t s success-ok --params "A=<(literal)"', ['success-ok']),
        ("./cfnrain deploy t s success-ok --params 'A=>(literal)'", ['success-ok']),
    ]
    for name in ('invalid-name', 'success-ok'):
        for redirect in ('> "$log"', '>>"$log"', '2> "$log"', '2>&1',
                         '&> "$log"', '&>>"$log"', '< "$input"',
                         '0<&3', '2>&-', '>| "$log"', '<> "$log"',
                         '>"a > b" 2>&1', '>"$log"'):
            cases.append((f'./cfnrain deploy t s {redirect} {name} --no-exec', [name]))
            if redirect.startswith(('>', '<', '&')):
                cases.append((f'./cfnrain deploy t s {name}{redirect} --no-exec', [name]))
        cases.extend([
            (f'result="$(./cfnrain deploy t "$(printf success-stack)" {name} --no-exec)"', [name]),
            (f'result="$(echo "$(./cfnrain deploy t s {name})")"', [name]),
            (f'result="$(./cfnrain deploy t s {name} --params \'A=f(x)\')"', [name]),
            (f'out="$(aws cloudformation describe-change-set --change-set-name {name} --query \'length(Changes)\')"', [name]),
            (f'./cfnrain deploy "$(./cfnrain deploy t s {name})" s {name}', [name, name]),
            (f'./cfnrain deploy t s {name} --params "A=>"', [name]),
        ])
    cases.append(('out="$(./cfnrain deploy t s "$(echo bad)" --no-exec)"', ['command-output']))
    for line, expected in cases:
        actual = list(names(line))
        assert actual == expected, (line, actual, expected)
    for line in (
        'out="$(./cfnrain deploy t s bad',
        'out="$(./cfnrain deploy t s bad --params (x))"',
        'out="$(./cfnrain deploy t s bad # comment)"',
        './cfnrain deploy t s >',
        './cfnrain deploy t s > # missing target',
        'diff <(./cfnrain ls -c s invalid-name) expected.txt',
        'cat input.txt >(./cfnrain ls -c s invalid-name)',
        'out=`./cfnrain ls -c s invalid-name`',
        'out="`./cfnrain ls -c s invalid-name`"',
    ):
        try:
            list(names(line))
        except ValueError:
            continue
        raise AssertionError(f"accepted ambiguous command: {line}")
    # Exercise prefix inference and the line filter through the CLI as well.
    import subprocess
    import tempfile
    invalid_sources = [
        "STACK_PREFIX=success-ok\n./cfnrain deploy t s '$STACK_PREFIX'",
        '"$rain" deploy t s invalid-name',
        '"${rain}" deploy t s invalid-name',
        'out=`./cfnrain ls -c s invalid-name`',
    ]
    for declaration in ('export', 'declare', 'readonly', 'local', 'declare -x'):
        invalid_sources.append(
            f'STACK_PREFIX=success-ok\n{declaration} STACK_PREFIX=bad\n'
            './cfnrain deploy t s "$STACK_PREFIX"')
    invalid_sources.append(
        'STACK_PREFIX=success-ok\nSTACK_PREFIX=bad extra\n'
        './cfnrain deploy t s "$STACK_PREFIX"')
    with tempfile.TemporaryDirectory() as directory:
        path = Path(directory) / 'integ-regression.sh'
        for source in invalid_sources:
            path.write_text('./cfnrain deploy t s success-baseline\n' + source + '\n')
            result = subprocess.run([sys.executable, __file__, str(path)],
                                    capture_output=True, text=True)
            assert result.returncode != 0 and f'{path}:' in result.stderr, (source, result)
    print(f"Passed {len(cases)} command cases, 9 parse error cases, and {len(invalid_sources)} CLI cases")


if sys.argv[1:] == ["--self-test"]:
    self_test()
    sys.exit(0)


# Derive safe prefixes from assignments, including shared sourced scripts.
prefixes = set()
sources = [p.read_text().replace("\\\n", " ") for p in paths]
assignments = []
for source in sources:
    for line in source.splitlines():
        if line.lstrip().startswith('#'):
            continue
        match = re.match(r'^\s*(\w+)=(.*)$', line)
        parsed = False
        if match:
            try:
                value = shlex.split(substitutions(match[2])[0], comments=True)
            except ValueError:
                value = []
            if len(value) == 1:
                assignments.append((match[1], value[0]))
                parsed = True
        if not parsed:
            # Declarations and other assignment forms must not retain earlier trust.
            for key in re.findall(r'(?:^|[\s;])(\w+)\+?=', line):
                assignments.append((key, ''))


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
        if not re.search(r'cfnrain|CFNRAIN|\$(?:rain\b|\{rain\})|--change-set-name', line):
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
