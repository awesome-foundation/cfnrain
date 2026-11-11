#!/usr/bin/env python3
"""Forward cancellation to the test process and drain its cleanup diagnostics."""
import os
import signal
import subprocess
import sys


def main():
    child = None
    pending = None
    forwarded = False

    def deliver():
        nonlocal forwarded
        forwarded = True
        try:
            os.killpg(child.pid, pending)
        except ProcessLookupError:
            pass

    def forward(signum, _frame):
        nonlocal pending
        # A second cancellation must not interrupt cleanup already in progress.
        if pending is not None:
            return
        pending = signum
        if child is not None:
            deliver()

    signal.signal(signal.SIGINT, forward)
    signal.signal(signal.SIGTERM, forward)
    env = dict(os.environ, INTEG_REDACTING="1")
    child = subprocess.Popen(sys.argv[1:], env=env, stdout=subprocess.PIPE,
                             stderr=subprocess.STDOUT, start_new_session=True)
    # Separate sessions prevent a terminal signal from killing the redactor or
    # delivering the same cancellation twice to the test process.
    redactor = subprocess.Popen([sys.executable, "scripts/integ-redact.py"],
                                stdin=child.stdout, start_new_session=True)
    child.stdout.close()
    if pending is not None and not forwarded:
        deliver()
    result = child.wait()
    redactor_result = redactor.wait()
    return (128 - result if result < 0 else result) or redactor_result


if __name__ == "__main__":
    sys.exit(main())
