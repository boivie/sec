+++
date = "2015-06-20T11:56:35+02:00"
title = "Request Signing"
messagetype="sign.request"

+++

## Introduction

{{% msginitial %}}

## Payload, mandatory fields

* `_type` (string): Set to "sign.request"
* `at` (timestamp): The timestamp when this message was created, specified
   as milliseconds since 1970-01-01 00:00:00 UTC. Note that servers may reject
   a message with a timestamp that is too far in the past or future.
* `message` (string): JWS signed message to sign.

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "WPNHSTXSFRL5S7VJ7ZLBKQUS3EUFZLVI4UP3RHIZ7GK53FB5H4QQ",
  "nonce": "cpeq00yf9xs8/qo4d3kwpgtg/iae7lnmc6smc7btgye="
}

{
  "_type": "sign.request",
  "at": 1434806059000,
  "message": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkw
              IiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.EkN-DOsnsuRjRO6B
              xXemmJDm3HbxrbRzXglbN2S4sOkopdU4IsDxTI8jO19W_A4K8ZPJijNLis4EZ
              sHeY559a4DFOd50_OqgHGuERTqYZyuhtF39yxJPAjUESwxk2J5k_4zM3O-vtd
              1Ghyo4IbqKKSy6J9mTniYJPenn5-HIirE"
}
{{< /highlight >}}
