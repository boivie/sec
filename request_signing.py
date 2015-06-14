#!/usr/bin/env python
import base64
import hashlib
import sys
from collections import OrderedDict

import requests
import jwt
from cryptography.hazmat.primitives.serialization import load_pem_private_key
from cryptography.hazmat.backends import default_backend


def b64_encode(source):
    if not isinstance(source, bytes):
        source = source.encode('ascii')

    encoded = base64.urlsafe_b64encode(source).replace(b'=', b'')
    return str(encoded.decode('ascii'))


def b64_decode(source):
    missing_padding = 4 - len(source) % 4
    if missing_padding:
        source += b'=' * missing_padding
    decoded = base64.urlsafe_b64decode(source)
    return decoded


class bcolors:
    HEADER = '\033[95m'
    BLUE = '\033[94m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    RED = '\033[91m'
    ENDC = '\033[0m'

private_key_fname = sys.argv[1]
fingerprint = sys.argv[2]
base_url = sys.argv[3]

payload = {'iss': 'victor', 'exp': 15,
           'claim': 'insanity'}

priv = load_pem_private_key(open(private_key_fname, 'r').read(),
                            password=None, backend=default_backend())
r = requests.post(base_url + "/request/").json()
req_id = r['id']
req_url = r['url']


payload = {
    "aud": "Sec",
    "iss": "host.example.com",
    "typ": "boivie/sec/login/v1",
    "request": {
        "reference": "902ba2ac8",
        "username": "user@host.example.com"
    }
}
jwt_message = jwt.encode(payload, priv, algorithm='RS256')
print(jwt_message)
requests.post(req_url, json=dict(jwt=jwt_message))
print(req_url)
