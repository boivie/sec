+++
date = "2015-06-20T11:56:35+02:00"
title = "Offer Identity"

+++

## Introduction

This is an initial topic message.

## Payload, mandatory fields

* `_type` (string): Set to "identity.offer"
* `at` (timestamp): The timestamp when this message was created, specified
   as milliseconds since 1970-01-01 00:00:00 UTC. Note that servers may reject
   a message with a timestamp that is too far in the past or future.
* `title` (string): Name of this identity.

## Payload, optional fields

* `ref` (string). Caller reference.

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "TN5XZI4WSLM6DD5NDUD6RZVSECROX6SBKHXWEBOPDJAKQ3RYO3UQ",
  "nonce": "tamyw6kc6buaotc0qkjqqtb9xqoyc5r9qxtqrfbski0="
}

{
  "_type": "identity.offer",
  "at": 1434806059000,
  "title": "John Doe",
  "ref": "employee:2394"
}
{{< /highlight >}}
