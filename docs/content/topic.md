+++
date = "2015-06-20T11:56:35+02:00"
title = "Topics"

+++

## Introduction

SEC maintains feeds of ordered messages in "topics".

A topic is identified by an id, which is 384 bits long and encoded in bitcoin
compatible base58(https://en.wikipedia.org/wiki/Base58). The topic ID consists
of two parts:
 * The first 256 bits are the resource identifier, which is the SHA256 digest
   of the initial encrypted message.
 * The remaining 128 bits are the encryption key used to decrypt the resource.

When communicating a topic identifier to a party that should be able to read and
understand the data in it, the full 384 bit topic id must be specified. This
id is known as `topic_id_and_key`

When communicating to a party that should only be able to read the data,
the first 256 bits should be used as topic identifier. This is typically the
case when communicating with the message registry. This topic is known as
`topic_id`

### Topic types

#### Root Topic

The root topic's initial message contains the root public key. This cannot be
changed or updated once created. All messages in this topic must be signed by the root private key.

The allowed types of messages are:
 * [root.config]({{< relref "messages/root/config.md" >}})

#### Account

The allowed types of messages are:
 * [account.create]({{< relref "messages/account/create.md" >}})

#### Identities

Used for distributing identities.

The allowed types of messages are:
 * [identity.offer]({{< relref "messages/identity/offer.md" >}})
 * [identity.claim]({{< relref "messages/identity/claim.md" >}})
 * [identity.issue]({{< relref "messages/identity/issue.md" >}})
 * [identity.activate]({{< relref "messages/identity/activate.md" >}})
 * [identity.deactivate]({{< relref "messages/identity/deactivate.md" >}})
 * [identity.revoked]({{< relref "messages/identity/revoked.md" >}})

#### Signing

Used for signing a payload.

The allowed types of messages are:
* [sign.request]({{< relref "messages/sign/request.md" >}})
* [sign.initiated]({{< relref "messages/sign/initiated.md" >}})
* [sign.performed]({{< relref "messages/sign/performed.md" >}})
* [sign.rejected]({{< relref "messages/sign/rejected.md" >}})
