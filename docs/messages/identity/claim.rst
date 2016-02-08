identity.claim
==============

..note:: Only one message of the type ``identity.claim`` can exist in a given topic.

..todo:: Should this be signed by the account key instead?

Payload, mandatory fields
-------------------------

* ``resource`` (string): Set to "identity.claim"
* ``oob_hash`` (base64): Optional hashed out-of-bounds data, described below.

Notes about oob_proof
---------------------

When offering identities, the issuer should provide some data to the individual
indenting to claim this identity in order to prove that the individual is the
actual indented recipient. This data, called ``oob_data`` will never be provided
to the system.

The ``oob_hash`` field should be the base64-encoded SHA256 hash of the message:
``topic`` + ``oob_data`` + ``account_id``, where + is the concatentation of the
strings.

Example
-------

Showing the protected JWS header and payload.

.. code-block:: json

    {
      "alg": "RS256",
      "jwk": {
        "kty": "RSA",
        "n": "vrjOfz9Ccdgx5nQudyhdoR17V-IubWMeOZCwX_jj0hgAsz2J_pqYW08
              PLbK_PdiVGKPrqzmDIsLI7sA25VEnHU1uCLNwBuUiCO11_-7dYbsr4iJmG0Q
              u2j8DsVyT1azpJC_NG84Ty5KKthuCaPod7iI7w0LK9orSMhBEwwZDCxTWq4a
              YWAchc8t-emd9qOvWtVMDC2BXksRngh6X5bUYLy6AyHKvj-nUy1wgzjYQDwH
              MTplCoLtU-o-8SNnZ1tmRoGE9uJkBLdh5gFENabWnU5m1ZqZPdwS-qo-meMv
              VfJb6jJVWRpl2SUtCnYG2C32qvbWbjZ_jBPD5eunqsIo1vQ",
        "e": "AQAB"
      },
      "nonce": "nliwch3bv2pallp95vrukktzjcd+dz7tpdybya0ijmc="
    }

    {
      "resource": "identity.claim",
      "topic": "FLprEtiKrK6ht5b3kziCACzzhX9cR2me99vUaysexb4d",
      "index": 1,
      "parent": "0eInsyhvWgJpM+2i0gw2AkrML8HdKxS+5Y4h4nTdo8c=",
      "at": 1434806059000,
      "oob_hash": "KMjiwHJayytCOAxaFjR3g+N2Kwq18pZYUdpjMAZkqaE="
    }
