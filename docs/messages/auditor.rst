Auditor Signature
-----------------

The auditor signature is a JWS signature with the following:

Unprotected JWS Header fields
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

* ``alg`` will be ``RS256``

Protected JWS Header fields
~~~~~~~~~~~~~~~~~~~~~~~~~~~

* ``alg`` will be ``RS256``
* ``kid`` will be set and valid.
* ``nonce`` will be set and valid.

Payload fields
~~~~~~~~~~~~~~

 * ``topic`` (string) The topic ID
 * ``index`` (number) The message index within the topic that this signature covers.
 * ``at``: (timestamp) The timestamp when this message was created, specified
   as milliseconds since 1970-01-01 00:00:00 UTC.
 * ``signature_hash``: (base64) The SHA256 of the JWS signature of the message signed.

Payload fields, optional
~~~~~~~~~~~~~~~~~~~~~~~~
 * ``references``: (array of objects) Direct references used (for validating certificates etc).
   * ``topic``: (ref) topic id
   * ``index`` (number) index used when validating.
   * ``hash`` (base64) The SHA256 of the auditor signature of that record.

example
~~~~~~~

.. code-block:: json

    {
      "alg": "RS256",
      "kid": "Br4R8rMhAK6Yj4nJxyLhp51pQsMurYypBhoBAcYejk6S",
      "nonce": "yyq1pv0IsqDWjTYlHHzHBGqcnLJth2LwKXfbMJzKt+U="
    }

    {
      "topic": "HVv3RdBv8ocYpCGV2oR2PQjDWUnoJ7bYrb7x3Cpzpxby",
      "index": 4,
      "parent": "w0rr0mamPbdHB2Momx9bN2J2a1FJk9tRcmvSOfYnmeU=",
      "at": 1436820304865,
      "signature_hash": "OBXzBo4Rks4kIvcrdLm9XfHZ0A1K+UlPS++zymJ76Uk=",
      "references": [
        {
          "topic": "G4FJy9dVX4H27Yvvmwi8BZyhSwd2wmWAwSkfHEjYVr2m",
          "index": 5,
          "hash": "0RN2SfhVxSScamiATvFi+wsIKtO6U4XgE24qi5yRY2s="
        }
      ]
    }
