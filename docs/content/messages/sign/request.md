+++
date = "2015-06-20T11:56:35+02:00"
title = "Request Signing"
messagetype="sign.request"

+++

## Introduction

Requests a message to be signed by a client.

TODO: Recipient handling - how to limit the identities that can sign it?

{{% msginitial %}}

## Payload, mandatory fields

* `resource` (string): Set to "sign.request"
* `message` (string): Human readable message to sign.

## Payload, optional fields

* `recipient` (string): Identity that should sign this message.

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "3GFoeJJJJwBAod4MMYvZssEogYoEkZjE66Lykow2Uc8e",
  "nonce": "cpeq00yf9xs8/qo4d3kwpgtg/iae7lnmc6smc7btgye="
}

{
  "resource": "sign.request",
  "at": 1434806059000,
  "message": "Pay $500?",
  "recipient": "FFt5jazN9tsDTZtD5MNZHXBJ4tBJ8efrGPx5wb1gBfbM",
  "ref": "transaction_id=33929202"
}
{{< /highlight >}}
