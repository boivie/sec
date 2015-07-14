+++
date = "2015-06-20T11:56:35+02:00"
title = "Accept Certificate"

+++

## Introduction

NOTE: This message can only be added to a topic containing a
      [identity.offer-cert]({{< relref "messages/identity.offer-cert.md" >}})
      message.

## Payload, mandatory fields

* `_type` (string): Set to "identity.accept-cert"
* `_parent` (base64). SHA256 of the parent message.

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "GIUWFIKBOBLCZNF4BOUPWFIYD5ADDBGA66D2D6M75U3GOERBYAOQ",
  "nonce": "mnjjr2j+4dyimxw1hm7boxm+jknfbzjihi4/oxftv6y="
}

{
  "_type": "identity.accept-cert",
  "_parent": "Ll5+74Nwim0bRHINsrB7ZFbfozY23dBhU1M1S6zHAvM="
}
{{< /highlight >}}
