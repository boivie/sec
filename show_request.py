#!/usr/bin/env python
import base64
import hashlib
import sys

import jws
import requests


def b64_encode(source):
    if not isinstance(source, bytes):
        source = source.encode('ascii')

    encoded = base64.urlsafe_b64encode(source).replace(b'=', b'')
    return str(encoded.decode('ascii'))


class bcolors:
    HEADER = '\033[95m'
    BLUE = '\033[94m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    RED = '\033[91m'
    ENDC = '\033[0m'

r = requests.get(sys.argv[1])
lines = [bytes(l) for l in r.text.strip().splitlines()]

prev_parent = None
for line in lines:
    (header, payload, sig) = line.split('.')
    record_id = b64_encode(hashlib.sha256(line.strip()).digest())
    fmt = bcolors.YELLOW + "record %s" + bcolors.ENDC
    print(fmt % record_id)
    header = jws.utils.decode(header)
    obj = jws.utils.decode(payload)
    obj_header = obj["header"]
    if prev_parent and obj_header["parent"] != prev_parent:
        fmt = bcolors.RED + "  invalid parent %s" + bcolors.ENDC
        print(fmt % obj_header["parent"])
    prev_parent = record_id
    fmt = bcolors.BLUE + "  type %s" + bcolors.ENDC
    print(fmt % header["typ"])

    refs = obj_header.get("refs") or []
    for ref in refs:
        fmt = bcolors.GREEN + "   ref %s" + bcolors.ENDC
        print(fmt % ref)

    fingerprint = header.get("x5t")
    if fingerprint:
        print("author %s" % fingerprint)
    print("")
    #jws.verify(header, claim, sig, key)
    del obj["header"]
    if obj:
        for key in obj:
            print("  %10s: %s" % (key, obj[key]))
        print("")
