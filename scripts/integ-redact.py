#!/usr/bin/env python3
"""Redact integration diagnostics before console output or artifact upload."""
import os
import re
import signal
import sys
from urllib.parse import unquote


def redact(text, template=False):
    if not template:
        text = re.sub(r"\x1b\[[0-?]*[ -/]*[@-~]", "", text)
        # Decode escaped identifiers before matching, including double encoding.
        for _ in range(3):
            text = unquote(text)
    for variable in ("INTEG_BUCKET_NAME", "INTEG_REDACT_ROLE"):
        value = os.environ.get(variable, "")
        if value:
            text = text.replace(value, "redacted-identifier")
            if variable == "INTEG_REDACT_ROLE":
                role = value.rsplit("/", 1)[-1]
                text = text.replace(role, "redacted-role")
    text = re.sub(r"\d{12}", "000000000000", text)
    text = re.sub(r"(?:assumed-role|role)/[^\s\"'<>]+", "role/redacted-role", text)
    if not template:
        text = re.sub(r"https?://[^\s\"<>]+", "<url>", text)
    return text


if __name__ == "__main__":
    signal.signal(signal.SIGINT, signal.SIG_IGN)
    signal.signal(signal.SIGTERM, signal.SIG_IGN)
    for line in sys.stdin:
        print(redact(line, template="--template" in sys.argv), end="", flush=True)
