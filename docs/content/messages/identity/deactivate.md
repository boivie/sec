+++
date = "2015-06-20T11:56:35+02:00"
title = "Deactivate"

+++

## Introduction

Indicates that the client has removed the identity from the target device.

NOTE: This message can only be added to a topic containing a
      [identity.claim]({{< relref "messages/identity.claim.md" >}})
      message.

## Payload, mandatory fields

* `resource` (string): Set to "identity.deactivate"

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "CRQnNvwodN9yKiCudYSWYzFRnRpWNpJHzkZfT4Qwh3HM",
  "nonce": "d/BrepGR9PDP0n8G+W2/FO7ICO+gaVNtVFnJi6jccxQ="
}

{
  "resource": "identity.deactivate",
  "topic": "32cBWiURdyAaTqpi7aVyqD4Cud2PyQbX8hb5QAeRvy6U",
  "index": 3,
  "parent": "eF06e11Y8yaiIoED/T0kLsts51APLtpPA4jPvathowk=",
  "at": 1434806059000
}
{{< /highlight >}}
