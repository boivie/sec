+++
date = "2015-06-20T11:56:35+02:00"
title = "Revoked"

+++

## Introduction

Indicates that the identity has been revoked by the issuer.

NOTE: This message can only be added to a topic containing a
      [identity.offer]({{< relref "messages/identity.offer.md" >}})
      message.

NOTE: This is the final message of a topic. No more messages will be allowed
      in a topic where this message is present.

## Payload, mandatory fields

* `resource` (string): Set to "identity.revoked"

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "6hreuczTmcpaDRXbrTtet5xcPxqZapwZfQrzGej9H2Q8",
  "nonce": "d/BrepGR9PDP0n8G+W2/FO7ICO+gaVNtVFnJi6jccxQ="
}

{
  "resource": "identity.reject",
  "topic": "CqiGhjqtC3Sre6cVfC7j6gFb87pn74HfgcukE2TMTf77",
  "index": 3,
  "parent": "eF06e11Y8yaiIoED/T0kLsts51APLtpPA4jPvathowk=",
  "at": 1434806059000
}
{{< /highlight >}}
