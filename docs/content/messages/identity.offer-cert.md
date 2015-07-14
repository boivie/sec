+++
date = "2015-06-20T11:56:35+02:00"
title = "Offer Certificate"

+++

## Introduction

NOTE: This message can only be added to a topic containing a
      [identity.claim]({{< relref "messages/identity.claim.md" >}})
      message.

The certificate's CN will be the topic's ID.
The issuer's cert must be present in a topic and will be validated.

## Payload, mandatory fields

* `_type` (string): Set to "identity.offer-cert"
* `_parent` (base64). SHA256 of the parent message.
* `cert` (base64) The base64 encoded DER certificate

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "3VTNN7TE3U5QJ7I3GRZZNV2HYP7FBPEIJ627NRYH2KKUND3VBPHQ",
  "nonce": "qspepbtqom3/ebh1nohk4ozj1tl7bqeohhydthixfdo="
}

{
  "_type": "identity.offer-cert",
  "_parent": "UgWZ3+FZ/li1HYyrZkSDRpyU0hLKd/Kq714YKY4xp6o=",
  "cert": "MIIFazCCA1OgAwIBAgIRAIIQz7DSQONZRGPgu2OCi"
}
{{< /highlight >}}
