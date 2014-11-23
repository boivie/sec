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

private_key_fname = sys.argv[1]
offer_url = sys.argv[2]
priv = rsa.importKey(open(private_key_fname))

r = requests.get(offer_url)
records = [bytes(c) for c in r.json()["records"]]
last_offer = None
last_cert = None
last_record_id = None
for rjws in records:
    last_record_id = b64_encode(hashlib.sha256(rjws).digest())

    (header, payload, sig) = rjws.split('.')
    header = jws.utils.decode(header)
    if header["typ"] == "offer":
        last_offer = (last_record_id, payload)
    if header["typ"] == "cert":
        last_cert = (last_record_id, payload)
    if header["typ"] == "close":
        print("This request is already closed")
        sys.exit(1)

if last_cert:
    payload = json.loads(b64_decode(payload))
    jws_header  = {'alg': 'RS256', 'typ': "accept",
                   'x5t#S256': payload['fingerprint']}
    h2 = OrderedDict([('parent', last_record_id),
                      ('refs', [last_cert[0]])])
    o2 = OrderedDict([('header', h2)])
    signature = jws.sign(jws_header, o2, priv)
    part1 = jws._signing_input(jws_header, o2)
    signed = part1 + "." + signature
    r = requests.post(offer_url, json=dict(records=[signed]))
    if r.status_code != 200:
        print("Failed to accept cert")
        sys.exit(1)

elif last_offer:
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

    r = requests.post(offer_url, json=dict(records=[signed]))
    if r.status_code != 200:
        print("Failed to claim offer")
        sys.exit(1)
else:
    print("No offer or cert found")
    sys.exit(1)
