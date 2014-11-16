import base64
import hashlib
import json
import sys
import ssl
import os
from collections import OrderedDict

import jws
import Crypto.PublicKey.RSA as rsa


def extract_pem(cert):
        capturing = False
        ret = []
        for l in cert.splitlines():
            if l.startswith("-----BEGIN CERTIFICATE-----"):
                capturing = True
            if capturing:
                ret.append(l)
            if l.startswith("-----END CERTIFICATE-----"):
                return "".join(ret)


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


class Record(object):
    def __init__(self, id, contents):
        self.id = id
        self.contents = contents

    def __str__(self):
        return self.contents


class PartSigner(object):
    def __init__(self, priv_key, cert = False):
        self.pkey = rsa.importKey(priv_key)
        self.fingerprint = None
        self.last_parent = None
        self.records = []
        if cert:
            der = ssl.PEM_cert_to_DER_cert(extract_pem(cert.read()))
            self.fingerprint = b64_encode(hashlib.sha1(der).digest())

    def load(self, lines):
        lines = lines.strip()
        if lines:
            last = lines.splitlines()[-1].strip()
            h = hashlib.sha256(last[0:last.rfind('.')]).digest()
            self.last_parent = b64_encode(h)

    def generate(self, typ, header, obj):
        if typ == "claim":
            jwk = {"kty": "RSA",
                   "n": base64_bigint(self.pkey.n),
                   "e": base64_bigint(self.pkey.e)}
            jws_header  = {'alg': 'RS256', 'typ': typ,
                           'jwk': json.dumps(jwk)}
        else:
            jws_header  = {'alg': 'RS256', 'typ': typ,
                           'x5t': self.fingerprint}

        h2 = OrderedDict([])
        if self.last_parent:
            h2['parent'] = self.last_parent
        for k in sorted(header.keys()):
            h2[k] = header[k]

        o2 = OrderedDict([('header', h2)])
        for k in sorted(obj.keys()):
            o2[k] = obj[k]

        signature = jws.sign(jws_header, o2, self.pkey)
        part1 = jws._signing_input(jws_header, o2)
        contents = part1 + "." + signature
        record_id = b64_encode(hashlib.sha256(contents).digest())
        self.last_parent = record_id
        r = Record(record_id, contents)
        self.records.append(r)
        return r


if __name__ == "__main__":
    PRIVATE_KEY = os.environ["SEC_PRIVKEY"]
    CERT = os.environ.get("SEC_CERT")
    request_id = sys.argv[1]
    pk_fobj = open(PRIVATE_KEY)
    cert_fobj = None
    if CERT:
        cert_fobj = open(CERT)

    s = PartSigner(pk_fobj, cert_fobj)
    r1 = s.generate("created",
                    dict(),
                    dict(request_id=request_id))
    r2 = s.generate("offer",
                    dict(),
                    dict(template_id="Y4ZMBVOVC4QDPJXWTNPPSA5UE4",
                         fields=dict(name=dict(value="Victor Boivie"),
                                     type=dict(value="employee")),
                         extra=dict(employee_id="23056791")))
    r3 = s.generate("claim",
                    dict(refs=[r2.id]),
                    dict(additional_keys=[]))
    r4 = s.generate("nop",
                    dict(),
                    dict())
    r5 = s.generate("cert-draft",
                    dict(refs=[r2.id, r3.id]),
                    dict(url="http://www.example.com/proposal.zip",
                         sha256="44d28fcb404be11b58e950c9ccb170cefdd70643372521f3e80a20b5eca7fb63"))
    r6 = s.generate("cert-draft-verified",
                    dict(refs=[r5.id]),
                    dict())
    r7 = s.generate("cert",
                    dict(),
                    dict(url="http://www.example.com/cert.zip",
                         sha256="f69c6d1896d70d61506fdadfc0cd29601abe8c0894d3d73a420d01d4cf871617"))
    r8 = s.generate("close",
                    dict(),
                    dict())

    for r in s.records:
        print(r)
