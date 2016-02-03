root.config
===========

.. note:: |initialmsg|

.. note:: The ``kid`` JWS header field will not be set. Instead, the ``jwk`` JWS
      header field will be set to provide the root public key.

Payload, mandatory fields
-------------------------

* ``resource`` (string): Set to "root.config"
* ``roots`` (JSON object)
  * ``auditor`` (JSON array of JWK keys): Valid keys

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
          "path": "/",
          "key": {
              "kid": "HhP2YvPiAAuDiWXFpmcD5dzwva9qCtd2Zwi2NcBMP9Lc",
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
        ]
      }
    }
