# RESI

RESI is the Redundant Energy Services Interface project. The RESI project is
an experimental repository has two contributions:

1. Gives a path toward a protocol specification for an ESI which allows
extensions

2. A demonstration of an extension to the ESI which allows for redundant
transports.

## Extendable ESI

This repository proposes to design the ESI protocol such that it allows for
future extension. The RESI uses SCTP[1] as inspiration. To better allow future
extensions to the ESI, RESI proposes that a special RPC is included in the ESI,
called `Modify`:

```
service Esi {
  // ..
  rpc Modify (ModifyRequest) returns (ModifyResponse) {}
  // ..
}
message Extension {
  uint32 id = 1;
  bool pass = 2;
  bytes data = 3;
}
message ModifyRequest {
  Extension ext = 1;
}
message ModifyResponse {
  Extension ext = 1;
}
```

The `Extension` message can contain any bytes encoded message which may be
implemented by an extending protocol. The `id` value is used to identify which
extension the message belongs to, while the `pass` value indicates whether 
unrecognized extension messages should be passed to other ESI servers.

## Redundant ESI

![Connect sequence diagram](diagrams/connect.png?raw=true "Connect sequence diagram")

RESI allows for redundant transport connections between ESI gateways. The ideal
method to achieve redundancy using RESI is to deploy two RESI gateways: one
in front of end nodes, and another in front of users of the ESI.

At the moment, two transport protocols are aimed to be supported by the RESI
extension: TCP and NKN[2].

[1]: https://datatracker.ietf.org/doc/html/rfc4960
[2]: https://nkn.org

