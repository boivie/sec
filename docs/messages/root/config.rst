root.config
===========

.. note:: |initialmsg|

.. note:: The ``kid`` JWS header field will not be set. Instead, the ``jwk`` JWS
      header field will be set to provide the root public key.

Payload, mandatory fields
-------------------------

* ``resource`` (string): Set to "root.config"
* ``roots`` (JSON object)
** ``auditor_roots`` (JSON array of root keys): Valid keys to issue auditor identities.
** ``identity_roots`` (JSON array of root keys): Valid keys to issue identities.

Root key
--------

A trusted root of keys.

* ``not_before`` (timestamp): the time before which the identity MUST NOT be
  accepted for processing.  The processing of the ``not_before`` field requires
  that the current date/time MUST be after or equal to the not-before date/time
  listed in the ``not_before`` field.
* ``not_after`` (timestamp): the expiration time on or after which the identity
   MUST NOT be accepted for processing. The processing of the ``not_after`` field
   requires that the current date/time MUST be before the expiration date/time
   listed in the ``not_after`` field.
* ``public_key`` (JWK key) Public key


Example
-------

Showing the protected JWS header and payload.

.. code-block:: json

    {
      "alg": "RS256",
      "nonce": "Pc1ai0zYOkY23j5USPSzyfdZlCEvJBbgoNg1dzFtp3s=",
      "jwk": {
        "kty": "RSA",
        "n": "vrjOfz9Ccdgx5nQudyhdoR17V-IubWMeOZCwX_jj0hgAsz2J_pqYW08
              PLbK_PdiVGKPrqzmDIsLI7sA25VEnHU1uCLNwBuUiCO11_-7dYbsr4iJmG0Q
              u2j8DsVyT1azpJC_NG84Ty5KKthuCaPod7iI7w0LK9orSMhBEwwZDCxTWq4a
              YWAchc8t-emd9qOvWtVMDC2BXksRngh6X5bUYLy6AyHKvj-nUy1wgzjYQDwH
              MTplCoLtU-o-8SNnZ1tmRoGE9uJkBLdh5gFENabWnU5m1ZqZPdwS-qo-meMv
              VfJb6jJVWRpl2SUtCnYG2C32qvbWbjZ_jBPD5eunqsIo1vQ",
        "e": "AQAB"
      }
    }

    {
      "resource": "root.config",
      "at": 1434806059000,
      "roots": [
        {
          "auditor_roots": [
            {
              "public_key": {
                "kid": "root/1",
                "kty": "RSA",
                "n": "vrjOfz9Ccdgx5nQudyhdoR17V-IubWMeOZCwX_jj0hgAsz2J_pqYW08
                      PLbK_PdiVGKPrqzmDIsLI7sA25VEnHU1uCLNwBuUiCO11_-7dYbsr4iJmG0Q
                      u2j8DsVyT1azpJC_NG84Ty5KKthuCaPod7iI7w0LK9orSMhBEwwZDCxTWq4a
                      YWAchc8t-emd9qOvWtVMDC2BXksRngh6X5bUYLy6AyHKvj-nUy1wgzjYQDwH
                      MTplCoLtU-o-8SNnZ1tmRoGE9uJkBLdh5gFENabWnU5m1ZqZPdwS-qo-meMv
                      VfJb6jJVWRpl2SUtCnYG2C32qvbWbjZ_jBPD5eunqsIo1vQ",
                "e": "AQAB"
              }
            }
          ],
          "identity_roots": [
            {
              "public_key": {
                "kid": "Foobar",
                "kty": "RSA",
                "n": "sJrAe-FrJoph77EMN41YpREAQR4Lvzd5lHbAciGW4ZrM8aGwt0WvwtbS6F
                      qVYj9b5XnwXTeOFMCYOMlLnASVyKecUuwKAkA5ysJlpkY5IQxH9b7hdRDf
                      7EVK8JOBXd-7hzxo4teNXtltPct_oC-TcE2ocxx3OTCDhgQ1bg_Yshc0Qg
                      mWGIkCJ1fnZMfKw40mp6-Ui93S-H0ZQXgWLqNLljTcy24ICB2M7-3uhv8w
                      Bigwry77TbbJy4vPrVThJHKf_-FAochIxdvc7alXiVQ4Ec-OvW6gPg_esb
                      ejvI4sj-MrWFc_Wzlse5cw5_rI9s_JJi-gEX39mSIxVbsFX6yrpQ",
                "e":"AQAB"
              }
            }
          ]
        }
      ]
    }
