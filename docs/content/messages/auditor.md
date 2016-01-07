# Auditor Signature

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
