#!/usr/bin/env python3
"""Check A6 values after original-style formatting, without AWS access."""
import json
import pathlib
import sys

template = json.loads(pathlib.Path(sys.argv[1]).read_text())
assert template["Metadata"] == {"M": {"x": "y"}}, "A6 explicit key value"
assert template["Resources"] == {
    "Handle": {"Type": "AWS::CloudFormation::WaitConditionHandle"}
}, "A6 resource"
assert template["Outputs"] == {
    "QuotedOn": {"Value": "ON"},
    "QuotedOff": {"Value": "OFF"},
    "Empty": {"Value": {"Fn::Join": ["", ["prefix", {"Fn::Sub": ""}]]}},
    "Anchor": {"Value": "anchored"},
    "Alias": {"Value": "anchored"},
    "Multiline": {"Value": "first\tcolumn\nsecond\n"},
}, "A6 output values"
