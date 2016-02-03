sign.performed
==============

Indicates that the user has signed the message.

{{% started %}}[sign.request]({{< relref "messages/sign.request.md" >}}){{% /started %}}

{{% notin %}}[sign.rejected]({{< relref "messages/sign.rejected.md" >}}){{% /notin %}}

NOTE: This message will be signed by the identity key.

Payload, mandatory fields
-------------------------

* ``resource`` (string): Set to "sign.performed"

Example
-------

Showing the JWS header and payload.

.. code-block:: json

    {
      "alg": "RS256",
      "kid": "7PJT4UGLGLZ5MDJRM7HO6TSQD52AVQXQ5UWSBNHCRNBCQJDANXCA",
      "nonce": "QgSovlxpXCce/OmEk8rifOm0ZNBSjFhWHPbZZQPwbzI="
    }

    {
      "resource": "sign.performed",
      "topic": "ABRAKADABRA",
      "index": 2,
      "parent": "WvnPibXJIEsfPMbxNKIdQ7hiS3wUqiJdqHpLQM5VaHo=",
      "at": 1434806059000
    }
