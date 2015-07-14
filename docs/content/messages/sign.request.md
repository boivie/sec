+++
date = "2015-06-20T11:56:35+02:00"
title = "Request Signing"
messagetype="sign.request"

+++

## Introduction

{{% msginitial %}}

## Payload, mandatory fields

* `_type` (string): Set to "sign.request"
* `at` (timestamp): The timestamp when this message was created, specified
   as milliseconds since 1970-01-01 00:00:00 UTC. Note that servers may reject
   a message with a timestamp that is too far in the past or future.
* `human_readable_text` (string) The text to display to the user

## Payload, optional fields

* `ref` (string). Caller reference.

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "WPNHSTXSFRL5S7VJ7ZLBKQUS3EUFZLVI4UP3RHIZ7GK53FB5H4QQ",
  "nonce": "cpeq00yf9xs8/qo4d3kwpgtg/iae7lnmc6smc7btgye="
}

{
  "_type": "sign.request",
  "at": 1434806059000,
  "human_readable_text": "Deploy version 2015-08-01#23 to production?",
  "ref": "deploy:backend#6684"
}
{{< /highlight >}}
