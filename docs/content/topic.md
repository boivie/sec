+++
date = "2015-06-20T11:56:35+02:00"
title = "Topics"

+++

## Introduction

SEC maintains feeds of ordered messages in "topics".

A topic is identified by an id, which is 256 bits long and encoded in bitcoin
compatible base58(https://en.wikipedia.org/wiki/Base58).

The initial message SHA256 digest is used as the topic's ID.

### Topic types

#### Root Topic

The root topic's initial message contains the root public key. This cannot be
changed or updated once created. All messages in this topic must be signed by the root private key.

The allowed types of messages are:
 * [root.config]({{< relref "messages/root/config.md" >}})

#### Identities

Used for distributing identities.

The allowed types of messages are:
 * [identity.offer]({{< relref "messages/identity/offer.md" >}})
 * [identity.claim]({{< relref "messages/identity/claim.md" >}})
 * [identity.offer-cert]({{< relref "messages/identity/offer-cert.md" >}})
 * [identity.accept-cert]({{< relref "messages/identity/accept-cert.md" >}})
 * [identity.revoked]({{< relref "messages/identity/revoked.md" >}})

#### Signing

Used for signing a payload.

The allowed types of messages are:
* [sign.request]({{< relref "messages/sign/request.md" >}})
* [sign.initiated]({{< relref "messages/sign/initiated.md" >}})
* [sign.performed]({{< relref "messages/sign/performed.md" >}})
* [sign.rejected]({{< relref "messages/sign/rejected.md" >}})
