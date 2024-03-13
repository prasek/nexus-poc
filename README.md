# Nexus PoC

This PoC is meant to demonstrate the simple "happy path" end to end experience for cross namespace calls with Nexus.
It does not represent the final APIs at this point.

TODO: include info on Nexus.

## Environment setup

- Update the Nexus Go SDK, Temporal Go SDK, and Temporal API Go submodules.

```
git submodule init
git submodule update
```

### Local

- Checkout the temporal OSS repo from https://github.com/bergundy/temporal/tree/nexus-poc.
- Run the local server (VS code has a launch config for an in-memory SQLite setup).


Run callee worker in namespace 1:
```
cd callee/worker
go run .
```

Run caller worker in namespace 2:
```
cd caller/worker
go run .
```

Start caller workflow from CLI
```
cd start
go run .
```

### Cloud

- Create a couple of namespaces, one for the caller side, and one for the handler side.
  (Ensure that the namespaces are placed in the PoC cluster).
- Obtain intra cluster certs.
- Port forward 7233 from a frontend host to your local machine.

#### Run

Obtain the nexus-poc-client.key and pem from 1password and run:

```
go run . -cloud -certs-dir <path-to-certs-dir> -skip-env-setup # optional
```
