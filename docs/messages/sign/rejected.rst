sign.rejected
=============

Indicates that the user has rejected signing the message.

NOTE: This message will be signed by the account key.

Payload, mandatory fields
-------------------------

* ``resource`` (string): Set to "sign.rejected"

Example
-------

Showing the JWS header and payload.

.. code-block:: json

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
