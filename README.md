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

## Configuring Redundancy

RESI supports three transportation methods:

1. ESI over TCP;
2. ESI over NKN;
3. ESI over a special RESI transport protocol

To configure the RESI gateway, pass in a `-config` flag which points to a
TOML flavoured configuration file:

```toml
# At least one of [tcp] or [nkn] must be present for RESI to function correctly.

# If the [tcp] section is not configued, RESI will not attempt to use TCP.
[Tcp]
Address = "127.0.0.1:9000" # IPv4 or IPv6
Listen = "localhost:4001"

# If the [nkn] section is not configued, RESI will not attempt to use NKN.
[Nkn]
Address = "919c54b38f907e82f030068c2c4a06239a2941f712306d9409474ebade479208"
Seed = "039e481266e5a05168c1d834a94db512dbc235877f150c5a3cc1e3903662d673"
Subclients = 4 # See NKN documentation for more details.

# RESI specific configuration options
[Resi]
# RESI can operate in either 'strict' or 'permissive' mode. In the default
# strict mode, identical data must be observed on all transports before it
# is passed through the gateway. This is considered the most reliable mode,
# although is not always beneficial and necessarily implies that the transport
# will be as slow as the slowest underlying protocol.
#
# In permissive mode, the RESI gateway will wait for the data to arrive on one
# of the transports and disregard any duplicate data that arrives in the
# future.
Mode = "strict"
```

[1]: https://datatracker.ietf.org/doc/html/rfc4960
[2]: https://nkn.org

