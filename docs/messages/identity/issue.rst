identity.issue
==============

Issues an identity to the public key that was provided in the ``identity.claim``
message.

NOTE: This message can only be added to a topic containing a
      [identity.claim]({{< relref "messages/identity.claim.md" >}})
      message.

NOTE: Multiple ``identity.issue`` may exist in the same topic, but since only
      one ``identity.claim`` can exist in a topic, an identity will always have
      the same public key throughout its lifetime. If the public key needs to
      be replaced, a new identity has to be created and the old one should be
      revoked.

Payload, mandatory fields
-------------------------

* ``resource`` (string): Set to "identity.issue"
* ``title`` (string): Display name of the identity
* ``public_key`` (JSON object): Public key object.

Payload, optional fields
------------------------

* ``path`` (string): Set when issuing identities that can act as identity
  authority and itself issue identities. This sets the path, which must be a
  sub-path to the signing identity. If this field is not set, the identity MUST
  NOT be allowed to be used for issuing identities.
* ``not_before`` (timestamp): the time before which the identity MUST NOT be
  accepted for processing.  The processing of the ``not_before`` field requires
  that the current date/time MUST be after or equal to the not-before date/time
  listed in the ``not_before`` field.
* ``not_after`` (timestamp): the expiration time on or after which the identity
   MUST NOT be accepted for processing. The processing of the ``not_after`` field
   requires that the current date/time MUST be before the expiration date/time
   listed in the ``not_after`` field.

Example
-------

Showing the JWS header and payload.

.. code-block:: json

    {
      "alg": "RS256",
      "kid": "D9o6se6Z3MCmpXRiSLd5b4NTzComVZ62fVwPVXsY4jgW",
      "nonce": "qspepbtqom3/ebh1nohk4ozj1tl7bqeohhydthixfdo="
    }

    {
      "resource": "identity.issue",
      "topic": "484Losw8UdPDFT7oLhxmpApVZuYGtJryoFmupJLntm9J",
      "index": 2,
      "parent": "KBbtlNNcBZyIzuwyt74G3M1ktODJRLrZDm+oJ6ltNxA=",
      "at": 1434806059000,
      "title": "John Doe",
      "public_key": {
          "kty": "RSA",
          "n": "vrjOfz9Ccdgx5nQudyhdoR17V-IubWMeOZCwX_jj0hgAsz2J_pqYW08
                PLbK_PdiVGKPrqzmDIsLI7sA25VEnHU1uCLNwBuUiCO11_-7dYbsr4iJmG0Q
                u2j8DsVyT1azpJC_NG84Ty5KKthuCaPod7iI7w0LK9orSMhBEwwZDCxTWq4a
                YWAchc8t-emd9qOvWtVMDC2BXksRngh6X5bUYLy6AyHKvj-nUy1wgzjYQDwH
                MTplCoLtU-o-8SNnZ1tmRoGE9uJkBLdh5gFENabWnU5m1ZqZPdwS-qo-meMv
                VfJb6jJVWRpl2SUtCnYG2C32qvbWbjZ_jBPD5eunqsIo1vQ",
          "e": "AQAB"
        },
    }
