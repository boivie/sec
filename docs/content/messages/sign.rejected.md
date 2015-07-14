+++
date = "2015-06-20T11:56:35+02:00"
title = "Signing Rejected"
messagetype = "sign.rejected"
+++

## Introduction

Indicates that the user has signed the message.

{{% started %}}[sign.request]({{< relref "messages/sign.request.md" >}}){{% /started %}}

{{% notin %}}[sign.performed]({{< relref "messages/sign.performed.md" >}}){{% /notin %}}


## Payload, mandatory fields

* `_type` (string): Set to "sign.rejected"
* `_parent` (base64). SHA256 of the parent message.
* `at` (timestamp): The timestamp when this message was created, specified
   as milliseconds since 1970-01-01 00:00:00 UTC. Note that servers may reject
   a message with a timestamp that is too far in the past or future.

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "YSVX7QBYJF4K32AXZJEGZTYXSCPUJGXD6PHQ4P5GF52PSXKQFEAQ",
  "nonce": "QgSovlxpXCce/OmEk8rifOm0ZNBSjFhWHPbZZQPwbzI="
}

{
  "_type": "sign.rejected",
  "_parent": "AUi+xBxbv1FLadu0iFWeagsOEqlZJ/XAKi0GWG92RPs=",
  "at": 1434806059000
}
{{< /highlight >}}
