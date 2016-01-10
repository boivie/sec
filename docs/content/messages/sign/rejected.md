+++
date = "2015-06-20T11:56:35+02:00"
title = "Signing Rejected"
messagetype = "sign.rejected"
+++

## Introduction

Indicates that the user has rejected signing the message.

{{% started %}}[sign.request]({{< relref "messages/sign.request.md" >}}){{% /started %}}

{{% notin %}}[sign.performed]({{< relref "messages/sign.performed.md" >}}){{% /notin %}}

NOTE: This message will be signed by the account key.

## Payload, mandatory fields

* `resource` (string): Set to "sign.rejected"

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "YSVX7QBYJF4K32AXZJEGZTYXSCPUJGXD6PHQ4P5GF52PSXKQFEAQ",
  "nonce": "QgSovlxpXCce/OmEk8rifOm0ZNBSjFhWHPbZZQPwbzI="
}

{
  "resource": "sign.rejected",
  "topic": "ABRAKADABRA",
  "index": 1,
  "parent": "dCEDEH+C7Kw4g4JFhzw61PB7okSZxb2B0qtdwYpDBCI=",
  "at": 1434806059000
}
{{< /highlight >}}
