#!/usr/bin/env python3
"""Reproduce cancellation during a redacted test run without invoking AWS."""
import os
from pathlib import Path
import selectors
import signal
import subprocess
import sys
import tempfile


def ready(process):
    with selectors.DefaultSelector() as selector:
        selector.register(process.stdout, selectors.EVENT_READ)
        assert selector.select(5), "child did not become ready"
        assert process.stdout.readline() == "ready\n"


with tempfile.TemporaryDirectory() as directory:
    worker = Path(directory) / "worker.sh"
    worker.write_text('''#!/bin/bash
set -eu
marker="$1"
cleanup() {
  status=$?
  trap - EXIT
  for step in first second third; do
    echo "cleanup $step 123456789012" >&2
    sleep 0.05
  done
  echo complete > "$marker"
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
echo ready
while :; do sleep 0.05; done
''')
    for signum in (signal.SIGINT, signal.SIGTERM):
        for group in (False, True):
            marker = Path(directory) / f"{signum}-{group}"
            process = subprocess.Popen(
                [sys.executable, "scripts/integ-run-redacted.py", "bash", str(worker), str(marker)],
                stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, start_new_session=True)
            try:
                ready(process)
                (os.killpg if group else os.kill)(process.pid, signum)
                stdout, stderr = process.communicate(timeout=5)
                assert process.returncode == 128 + signum, (process.returncode, stderr)
                assert marker.read_text() == "complete\n"
                assert stdout.count("cleanup ") == 3, stdout
                assert "123456789012" not in stdout + stderr
            finally:
                if process.poll() is None:
                    os.killpg(process.pid, signal.SIGKILL)
                    process.wait()
        # Even a signal sent directly to the redactor must not close its pipe.
        process = subprocess.Popen([sys.executable, "scripts/integ-redact.py"],
                                   stdin=subprocess.PIPE, stdout=subprocess.PIPE, text=True)
        process.stdin.write("ready\n")
        process.stdin.flush()
        ready(process)
        process.send_signal(signum)
        stdout, _ = process.communicate("cleanup 123456789012\n", timeout=5)
        assert process.returncode == 0
        assert stdout == "cleanup 000000000000\n"
print("SIGINT and SIGTERM: cleanup completed, output drained, exit codes preserved")
