+++
date = "2015-06-20T11:56:35+02:00"
title = "Workflows"

+++

## Introduction

As described in the terminology section, messages are signed JSON payloads
that are added to a topic. A message has a "type" which defines its structure
and use. The server will validate the message type when they are added to
a topic.

## Get Identity Flow

Used for distributing identities

 * [identity.offer]({{< relref "messages/identity.offer.md" >}})
 * [identity.claim]({{< relref "messages/identity.claim.md" >}})
 * [identity.offer-cert]({{< relref "messages/identity.offer-cert.md" >}})
 * [identity.accept-cert]({{< relref "messages/identity.accept-cert.md" >}})

## Revoke Identity Flow

Used when a previously issued identity is to be revoked

 * [identity.revoked]({{< relref "messages/identity.revoked.md" >}})

##  Request Signing Flow
 * [sign.request]({{< relref "messages/sign.request.md" >}})
 * [sign.initiated]({{< relref "messages/sign.initiated.md" >}})
 * [sign.performed]({{< relref "messages/sign.performed.md" >}})
 * [sign.rejected]({{< relref "messages/sign.rejected.md" >}})

