+++
date = "2015-06-20T11:56:35+02:00"
title = "Offer Identity"

+++

## Introduction

This is an initial topic message.

## Payload, mandatory fields

* `resource` (string): Set to "identity.offer"
* `title` (string): Name of this identity.
* `path` (string): Specifies the signing path that is allowed to sign this
  identity.

## Example

Showing the protected JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "TN5XZI4WSLM6DD5NDUD6RZVSECROX6SBKHXWEBOPDJAKQ3RYO3UQ",
  "nonce": "tamyw6kc6buaotc0qkjqqtb9xqoyc5r9qxtqrfbski0="
}

{
  "resource": "identity.offer",
  "at": 1434806059000,
  "title": "John Doe",
  "path": "/employee/",
  "ref": "employee:2394"
}
{{< /highlight >}}
