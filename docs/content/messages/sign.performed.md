+++
date = "2015-06-20T11:56:35+02:00"
title = "Signing Performed"
messagetype = "sign.performed"
+++

## Introduction

Indicates that the user has signed the message.

{{% started %}}[sign.request]({{< relref "messages/sign.request.md" >}}){{% /started %}}

{{% notin %}}[sign.rejected]({{< relref "messages/sign.rejected.md" >}}){{% /notin %}}


## Payload, mandatory fields

* `_type` (string): Set to "sign.performed"
* `_parent` (base64). SHA256 of the parent message.
* `at` (timestamp): The timestamp when this message was created, specified
   as milliseconds since 1970-01-01 00:00:00 UTC. Note that servers may reject
   a message with a timestamp that is too far in the past or future.

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "7PJT4UGLGLZ5MDJRM7HO6TSQD52AVQXQ5UWSBNHCRNBCQJDANXCA",
  "nonce": "QgSovlxpXCce/OmEk8rifOm0ZNBSjFhWHPbZZQPwbzI="
}

{
  "_type": "sign.performed",
  "_parent": "XDBLO//e4Ti76DK87sydkB+bQXe4Vf7MUYNTJJjCF/c=",
  "at": 1434806059000
}
{{< /highlight >}}
