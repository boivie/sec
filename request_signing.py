#!/usr/bin/env python
import base64
import hashlib
import json
import sys
from collections import OrderedDict

import requests
import jws
import Crypto.PublicKey.RSA as rsa


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

priv = rsa.importKey(open(private_key_fname))

r = requests.post(base_url + "/request/").json()
req_id = r['id']
req_url = r['url']

h1  = {'alg': 'RS256', 'typ': "create",
       'x5t#S256': fingerprint}
p1 = OrderedDict([('header', {}),
                  ('request_id', req_id)])
s1 = jws._signing_input(h1, p1) + "." + jws.sign(h1, p1, priv)
id1 = b64_encode(hashlib.sha256(s1).digest())

h2  = {'alg': 'RS256', 'typ': "loginreq",
       'x5t#S256': fingerprint}
p2 = OrderedDict([('header', dict(parent=id1)),
                  ('identity', 'user@example.com'),
                  ('target', 'example.com'),
                  ('reference', '902ba2ac8')])
s2 = jws._signing_input(h2, p2) + "." + jws.sign(h2, p2, priv)

requests.post(req_url, json=dict(records=[s1, s2]))
print(req_url)
