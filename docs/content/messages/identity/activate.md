+++
date = "2015-06-20T11:56:35+02:00"
title = "Activate"

+++

## Introduction

Indicates that the client has acknowledged the issued identity and is ready
to use it. This message is signed by the account key.

NOTE: This message can only be added to a topic containing a
      [identity.issue]({{< relref "messages/identity.issue.md" >}})
      message.

## Payload, mandatory fields

* `resource` (string): Set to "identity.activate"

## Example

Showing the JWS header and payload.

{{< highlight json >}}
{
  "alg": "RS256",
  "kid": "Ge5fzLDoopZ5HWndHQTCGPUyBrn4q5MD7BapH6RpGGxh",
  "nonce": "mnjjr2j+4dyimxw1hm7boxm+jknfbzjihi4/oxftv6y="
}

{
  "resource": "identity.activate",
  "topic": "49iRWcz4mwz7HtgcCJCYcGXPb1eeDh6WghsxB4RZ8Yau",
  "index": 3,
  "parent": "qBkoixFkkvidmNMQVWQ7cCk0s9kQvxecZpKnoMxNGCo=",  
  "at": 1434806059000
}
{{< /highlight >}}
