Messages
========

Introduction
------------

As described in the terminology section, messages are JWS signed JSON payloads
that are added to a topic. A message has a "type" which defines its structure
and use. The server will validate the message type when they are added to
a topic.

JWS requirements
~~~~~~~~~~~~~~~~

The server will reject messages that do not fulfill the following requirements.

1. The JWS MUST use the Flattened JSON Serialization
2. The JWS MUST be encoded using UTF-8
3. The unprotected JWS header field ``alg`` (algorithm) MUST be set to ``RS256``.
   No other fields are set in the unprotected JWS header.
4. The protected JWS header field ``alg`` (algorithm) MUST be set to ``RS256``.
5. The ``nonce`` field in the JWS protected header MUST be present and MUST
   contain at least 256 bits of entropy.
6. The ``kid`` field in the JWS header MUST be present, except for the following
   resources where the ``jwk`` field must be present in both the unprotected and
   protected JWS header.
    * [account.create]({{< relref "messages/account/create.md" >}})
    * [root.config]({{< relref "messages/root/config.md" >}})

Standard fields
~~~~~~~~~~~~~~~

The following fields are standard fields and must appear in all message payloads.

* ``resource`` (string). Defines the type of the message (e.g. ``identity.claim``)
* ``topic`` (string). The topic to which this message is posted. This field MUST
  NOT be present in the initial message for a given topic.
* ``index`` (number). Index of the message within the topic (zero-based). The
  server will validate that the indexes are monotonically increasing. This field
  SHOULD NOT be present in the initial message for a given topic.
* ``parent`` (base64). The base64-encoded SHA256 of the previous message's
  signature. This field MUST NOT be present in the initial message for a given
  topic.
* ``at`` (timestamp): The timestamp when this message was created, specified
  as milliseconds since 1970-01-01 00:00:00 UTC. Note that servers may reject
  a message with a timestamp that is too far in the past or future.
* ``ref`` (string). Optional caller reference. The length of the string MUST be
  less than or equal to 256 characters.
