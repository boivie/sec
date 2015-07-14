+++
date = "2015-06-20T11:56:35+02:00"
title = "Signing Initiated"
messagetype = "sign.initiated"

+++

## Introduction

Indicates that the user has been prompted to sign this message. This message
can be used to know the progress of the signing request and provide step-by-step
instructions to the user.

{{% msgin %}}[sign.request]({{< relref "messages/sign.request.md" >}}){{% /msgin %}}

NOTE: This message is optional.

## Payload, mandatory fields

* `_type` (string): Set to "sign.initiated"
* `_parent` (base64). SHA256 of the parent message.
* `at` (timestamp): The timestamp when this message was created, specified
   as milliseconds since 1970-01-01 00:00:00 UTC. Note that servers may reject
   a message with a timestamp that is too far in the past or future.

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "JYRZLPYGENJEAR6EM5U7PUKZZ557F243JYXSXQR5Z657JZ43LMGQ",
  "nonce": "G98nOayupTE7arCXZ8hK3P1VGa+PNA2us/RVCigLLew="
}

{
  "_type": "sign.initiated",
  "_parent": "POjPY10O/PKNGI/8LqrRG9BZHq+Iamv46JzFfaWdRG8=",
  "at": 1434806059000
}
{{< /highlight >}}
