+++
date = "2015-06-20T11:56:35+02:00"
title = "Messages"

+++

## Introduction

As described in the terminology section, messages are JWS signed JSON payloads
that are added to a topic. A message has a "type" which defines its structure
and use. The server will validate the message type when they are added to
a topic.

## JWS requirements

The server will reject messages that do not fulfill the following requirements.

1. The JWS MUST use the Flattened JSON Serialization
1. The JWS MUST be encoded using UTF-8
1. The `alg` (algorithm) field in the protected JWS header MUST be `RS256`
1. The `nonce` field in the JWS protected header MUST be present and MUST be at
   least 256 bits and should be base64-encoded. The server will validate that it
   is of sufficient length (minimum 256/8 * 4/3 = 43 bytes, including padding).
1. The `kid` field in the JWS header must be present in the protected JWS
   header, except for the following resources where the `jwk` field must be
   present instead:
    * [identity.claim]({{< relref "messages/identity/claim.md" >}})
    * [root.config]({{< relref "messages/root/config.md" >}})

## Standard fields

The following fields are standard fields and must appear in all message payloads.

* `resource` (string). Defines the type of the message (`identity.claim` etc)
* `index` (number). Index of the message within the topic (zero-based).

## Auditor Signature

The auditor signature is a JWS signature with the following:

### Protected JWS Header fields

* `alg` will be `RS256`
* `kid` will be set and valid.
* `nonce` will be set and valid.

### Payload fields

 * `at`: (timestamp) The timestamp when this message was created, specified
   as milliseconds since 1970-01-01 00:00:00 UTC.
 * `index` (number) The message index within the topic that this signature covers.
 * `payload_hash`: (base64) The SHA256 of the JWS payload.
 * `header_hash`: (base64) The SHA256 of the protected JWS header.
 * `parent`: (base64) The SHA256 of the previous auditor signature.
 * `references`: (array of objects) Direct references used (for validating certificates etc).
   * `topic`: (ref) topic id
   * `index` (number) index used when validating.

#### example

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "ZLQXVIQ5GH3Q6Z4ZACEVACYF5KAQFGUOJV3SPEXDHGKPHTSSAL7A",
  "nonce": "yyq1pv0IsqDWjTYlHHzHBGqcnLJth2LwKXfbMJzKt+U="
}

{
  "at": 1436820304865,
  "hash": "OBXzBo4Rks4kIvcrdLm9XfHZ0A1K+UlPS++zymJ76Uk=",
  "parent": "w0rr0mamPbdHB2Momx9bN2J2a1FJk9tRcmvSOfYnmeU=",
  "references": [
    {
      "topic": "TBFXUGQ4HZH6KT7ZOPBSSCM63NJLZDNPYOIAAQAUR7DJWGVTDZBQ",
      "index": 5
    }
  ]
}
{{< /highlight >}}
