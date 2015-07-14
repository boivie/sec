+++
date = "2015-06-20T11:56:35+02:00"
title = "Design"

+++

## Offer Identity

Format found at: [offer-identity]({{< relref "messages/offer-identity.md" >}})

<code class="sequence-diagram">
Note left of Client: Generates a offer-identity\nmessage and signs it using\nits RPC cert.
Client->WFE: Offer identity (signed payload)
Note over WFE: Validate the POST request
WFE->RA: Offer identity (signed payload)
Note over RA: Fetch "by" and validate cert\n chain and purpose. Validate "at".
Note over RA: Generate "registration payload"
RA->CA: Sign "registration payload"
CA->RA: Signature
Note over RA: Add message to topic
RA->WFE: Added
WFE->Client: Added
</code>


 * [claim-identity]({{< relref "messages/claim-identity.md" >}})
 * [offer-certificate]({{< relref "messages/offer-certificate.md" >}})
 * [accept-certificate]({{< relref "messages/accept-certificate.md" >}})

#### Revoke Certificate Flow

Used when a previous certificate is to be revoked

 * [revoke-certificate]({{< relref "messages/revoke-certificate.md" >}})

### Request Signing Flow
 * [request-signing]({{< relref "messages/request-signing.md" >}})
 * [signing-initiated]({{< relref "messages/signing-initiated.md" >}})
 * [signing-performed]({{< relref "messages/signing-performed.md" >}})
 * [signing-rejected]({{< relref "messages/signing-rejected.md" >}})
