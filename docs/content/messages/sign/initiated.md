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

NOTE: This message will be signed by the account key.

## Payload, mandatory fields

* `resource` (string): Set to "sign.initiated"

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "JYRZLPYGENJEAR6EM5U7PUKZZ557F243JYXSXQR5Z657JZ43LMGQ",
  "nonce": "G98nOayupTE7arCXZ8hK3P1VGa+PNA2us/RVCigLLew="
}

{
  "resource": "sign.initiated",
  "topic": "ABRAKADABRA",
  "index": 1,
  "parent": "SeRLB9mmltxcgL6JJ2adlXnkwUHfPrp3l6Zae2+6xl0=",
  "at": 1434806059000
}
{{< /highlight >}}
