import base64
import hashlib
import json
import sys
import ssl
import os

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

            
def base64_bigint(i):
    b = bytearray()
    while i:
        b.append(i & 0xFF)
        i >>= 8
    return base64.b64encode(bytes(b))


class PartSigner(object):
    def __init__(self, priv_key, cert = False):
        self.pkey = rsa.importKey(priv_key)
        self.fingerprint = None
        self.last_parent = None
        if cert:
            der = ssl.PEM_cert_to_DER_cert(extract_pem(cert.read()))
            self.fingerprint = hashlib.sha1(der).hexdigest()

    def load(self, lines):
        lines = lines.strip()
        if lines:
            last = lines.splitlines()[-1].strip()
            h = hashlib.sha256(last[0:last.rfind('.')]).digest()
            self.last_parent = base64.b64encode(h)

    def generate(self, section_type, obj):
        if section_type == "public_keys":
            jwk = {"kty": "RSA",
                   "n": base64_bigint(self.pkey.n),
                   "e": base64_bigint(self.pkey.e),
                   "d": base64_bigint(self.pkey.d)}
            header  = {'alg': 'RS256', 'typ': section_type,
                       'jwk': jwk}
        else:
            header  = {'alg': 'RS256', 'typ': section_type,
                       'x5t': self.fingerprint}
        if self.last_parent:
            obj['parent'] = self.last_parent
        signature = jws.sign(header, obj, self.pkey)
        part1 = jws._signing_input(header, obj)
        self.last_parent = base64.b64encode(hashlib.sha256(part1).digest())
        return part1 + "." + signature


if __name__ == "__main__":
    PRIVATE_KEY = os.environ["SEC_PRIVKEY"]
    CERT = os.environ.get("SEC_CERT")
    invitation_id = sys.argv[1]
    pk_fobj = open(PRIVATE_KEY)
    cert_fobj = None
    if CERT:
        cert_fobj = open(CERT)

    s = PartSigner(pk_fobj, cert_fobj)
    print(s.generate("invitation",
                     dict(invitation_id=invitation_id,
                          template_id="Y4ZMBVOVC4QDPJXWTNPPSA5UE4",
                          fields=dict(name=dict(value="Victor Boivie"),
                                      type=dict(value="employee")),
                          extra=dict(employee_id="23056791"))))
    print(s.generate("msg",
                     dict(message="This is a simple message")))
