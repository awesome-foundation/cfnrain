#!/usr/bin/env python3
"""Check the diagnostic forms reported by the integration review."""
import importlib.util
import os
import sys

sys.dont_write_bytecode = True

spec = importlib.util.spec_from_file_location("redactor", "scripts/integ-redact.py")
redactor = importlib.util.module_from_spec(spec)
spec.loader.exec_module(redactor)
os.environ["INTEG_BUCKET_NAME"] = "private-test-bucket"
os.environ["INTEG_REDACT_ROLE"] = "arn:aws:iam::123456789012:role/private-role"
for value in (
    "arn%3Aaws%3Asts%3A%3A123456789012%3Aassumed-role/private-role/session",
    "arn%253Aaws%253Aiam%253A%253A123456789012%253Arole/private-role",
    "\x1b[0m123456789012_x",
    "private-test-bucket private-role",
    "https://example.invalid/?X-Amz-Credential=secret",
):
    result = redactor.redact(value)
    for secret in ("123456789012", "private-role", "private-test-bucket", "secret", "\x1b"):
        assert secret not in result, result
assert redactor.redact("exit=1 duration=42\n") == "exit=1 duration=42\n"
assert redactor.redact("https://bucket-123456789012.s3.example/key", template=True) == "https://bucket-000000000000.s3.example/key"
