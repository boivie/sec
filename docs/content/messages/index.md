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

A client should reject messages that do not fulfill the following requirements.

1. The `alg` (algorithm) field in the JWS header must be `RS256`
1. The `nonce` field in the JWS header must be present. It is recommended, but
   not enforced, that the nonce is 256 bits and base64-encoded.
1. Except for the [claim-identity]({{< relref "messages/claim-identity.md" >}})
   message type, the `kid` field in the JWS header must be present.

## Standard fields

Fields starting with an underscore are generic fields, in contrast to
type-specific fields which must not start with an underscore. The
allowed fields are:

 * `_type` (string). Defines the type of the message. The server receiving
   the message must handle the message type to allow it to be added to a
   topic. The supported types are specified in the
   [workflow]({{< relref "messages/workflow.md" >}}) page.
* `_parent` (base64). SHA256 of the parent message, used for chaining.

## Auditor Signature

The auditor signature is a JWS signature with the following:

### Header fields

* `alg` will be `RS256`
* `kid` will be set and valid.
* `nonce` will be set and valid.

### Payload fields

 * `at`: (timestamp) The timestamp when this message was created, specified
   as milliseconds since 1970-01-01 00:00:00 UTC.
 * `index` (number) The message index within the topic that this signature covers.
 * `hash`: (base64) The SHA256 of the entire message, in the JWS encoded form.
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
