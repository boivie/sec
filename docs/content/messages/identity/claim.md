+++
date = "2015-06-20T11:56:35+02:00"
title = "Claim Identity"

+++

## Introduction

{{% started %}}[identity.offer]({{< relref "messages/identity/offer.md" >}}){{% /started %}}

NOTE: This message must be signed by the provided public key. It will however
      not set the `kid` JWS header property. Instead, the `jwk` JWS header
      property will be set to provide the public key to be signed.

## Payload, mandatory fields

* `resource` (string): Set to "identity.claim"
* `algorithm` (ref) The certificate algorithm used. Can be one of:
  * `RSA2048` 2048 bit RSA certificate.
* `at` (timestamp): The timestamp when this message was created, specified
   as milliseconds since 1970-01-01 00:00:00 UTC. Note that servers may reject
   a message with a timestamp that is too far in the past or future.
* `title` (string): Name of this identity

## Payload, optional fields

* `ref` (string). Caller reference.

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "jwk": {
    "kty": "RSA",
    "n": "vrjOfz9Ccdgx5nQudyhdoR17V-IubWMeOZCwX_jj0hgAsz2J_pqYW08
          PLbK_PdiVGKPrqzmDIsLI7sA25VEnHU1uCLNwBuUiCO11_-7dYbsr4iJmG0Q
          u2j8DsVyT1azpJC_NG84Ty5KKthuCaPod7iI7w0LK9orSMhBEwwZDCxTWq4a
          YWAchc8t-emd9qOvWtVMDC2BXksRngh6X5bUYLy6AyHKvj-nUy1wgzjYQDwH
          MTplCoLtU-o-8SNnZ1tmRoGE9uJkBLdh5gFENabWnU5m1ZqZPdwS-qo-meMv
          VfJb6jJVWRpl2SUtCnYG2C32qvbWbjZ_jBPD5eunqsIo1vQ",
    "e": "AQAB"
  },
  "nonce": "nliwch3bv2pallp95vrukktzjcd+dz7tpdybya0ijmc="
}

{
  "resource": "identity.claim",
  "index": 1,
  "algorithm": "RSA2048",
  "at": 1434806059000,
  "title": "John Doe, jdoe@example.com"
}
{{< /highlight >}}
