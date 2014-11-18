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


def base64_bigint(source):
    result_reversed = []
    while source:
        source, remainder = divmod(source, 256)
        result_reversed.append(remainder)

    return b64_encode(bytes(bytearray(reversed(result_reversed))))


class bcolors:
    HEADER = '\033[95m'
    BLUE = '\033[94m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    RED = '\033[91m'
    ENDC = '\033[0m'

offer_url = sys.argv[1]
private_key_fname = sys.argv[2]
priv = rsa.importKey(open(private_key_fname))

r = requests.get(offer_url)
lines = [bytes(l) for l in r.text.strip().splitlines()]
last_offer = None
last_record_id = None
for line in lines:
    last_record_id = b64_encode(hashlib.sha256(line.strip()).digest())

    (header, payload, sig) = line.split('.')
    header = jws.utils.decode(header)
    if header["typ"] == "offer":
        last_offer = (last_record_id, payload)
    if header["typ"] == "close":
        print("This request is already closed")
        sys.exit(1)

if not last_offer:
    print("No offer found")
    sys.exit(1)

jwk = {"kty": "RSA",
       "n": base64_bigint(priv.n),
       "e": base64_bigint(priv.e)}
jws_header  = {'alg': 'RS256', 'typ': "claim",
               'jwk': json.dumps(jwk)}

h2 = OrderedDict([('parent', last_record_id),
                  ('refs', [last_offer[0]])])
o2 = OrderedDict([('header', h2),
                  ('fields', {})])
signature = jws.sign(jws_header, o2, priv)
part1 = jws._signing_input(jws_header, o2)

signed = part1 + "." + signature

r = requests.post(offer_url, data=signed)
if r.status_code != 200:
    print("Failed to claim offer")
    sys.exit(1)
