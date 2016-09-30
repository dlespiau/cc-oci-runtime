# `cc-proxy`

`cc-proxy` is a daemon offering access to the
[`hyperstart`](https://github.com/hyperhq/hyperstart) VM agent to multiple
clients on the host.

![High-level Architecture Diagram](../documentation/high-level-overview.png)

- The `hyperstart` interface consists of:
    - A control channel on which the [`hyperstart` API]
      (https://github.com/hyperhq/runv/tree/master/hyperstart/api/json) is
      delivered.
    - An I/O channel with the stdin/stout/stderr streams of the processes
      running inside the VM multiplexed onto.
- `cc-proxy`'s main role is to:
    - Arbitrate access to the `hyperstart` control channel between all the
      instances of the OCI runtimes and `cc-shim`.
    - Route the I/O streams between the various shim instances and `hyperstart`.
- There's only one instance of `cc-proxy` per host.
 

`cc-proxy` itself has an API to setup the route to the hypervisor/hyperstart
and to forward `hyperstart` commands. This API is done with a small JSON RPC
protocol on an `AF_UNIX` located at: `${localstatesdir}/run/cc-oci-runtime/proxy.sock`

## Protocol

The proxy protocol is composed of JSON messages: requests and responses. Each
of these message is on a separate line. The snippet below is showing two
requests (`hello` and `hyper` commands) and their corresponding responses:

```
{ "id": 0, "command": "hello", "data": { "containerId": "foo", "ctlSerial": "/tmp/sh.hyper.channel.0.sock", "ioSerial": "/tmp/sh.hyper.channel.1.sock"  } }
{"id":0,"success":true}
{ "id": 1, "command": "hyper", "data": { "name": "ping" }}
{"id":1,"success":true}
```

Requests have 3 fields: `id`, `command` and `data`.

```
type Request struct {
	Id      uint             `json:"id"`
	Command string           `json:"command"`
	Data    *json.RawMessage `json:"data"`
}
```

The `id` identifies the request and the corresponding response will have the
same `id`. Each request carries a `command` which can have `data` associated,
not unlike what a simple JSON RPC protocol would have.

Responses have 4 fields: `id`, `success`, `error` and `data`

```
type Response struct {
	Id      uint                   `json:"id"`
	Success bool                   `json:"success"`
	Error   string                 `json:"error,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}
```

Unsurprisingly, the response has the result of a command, with `success`
indicating if the request has succeeded for not. If `success` is `true`, the
response can carry additional return values in `data`. If success if `false`,
`error` may contain an error string.

### `hello`

```
type helloCmd struct {
        ContainerId string `json:containerId`
        CtlSerial   string `json:ctlSerial`
        IoSerial    string `json:ioSerial`
}
```

The `hello` command is issued first after connecting to the proxy socket.  It's
used to let the proxy know about a new container on the system along with the
paths go hyperstart's command and I/O channels (`AF_UNIX` sockets).

### `bye`

The `bye` commands does the opposite of what `hello` does, indicating to the
proxy it should release resources created by `hello`. This command has no
parameter.

### `hyper`

```
type hyperCmd struct {
        Name string          `json:name`
        Data json.RawMessage `json:data`
}
```

The `hyper` command will forward a command to hyperstart.

## `systemd` integration

When compiling in the presence of the systemd pkg-config file, two systemd unit
files are created and installed.

  - `cc-proxy.service`: the usual service unit file
  - `cc-proxy.socket`: the socket activation unit

The proxy doesn't have to run all the time, just when a Clear Container is
running. Socket activation can be used to start the proxy when a client
connects to the socket for the first time.

After having run `make install`, socket action is enabled with:

```
sudo systemctl enable cc-proxy.socket
```

The proxy can output log messages on stderr, which are automatically
handled by systemd and ca be viewed with:

```
journalctl -u cc-proxy -f
```
