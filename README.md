# webfs

`webfs` is a single binary nfs-over-websocket proxy. For when you need filesystem-like POSIX semantics over the wire, and can't use a direct netpath or object store. It's like WebDAV, but better... or worse? Either way, if you find yourself needing this, I apologize in advance.

## Building

You must have `golang` available. Then, simply run:

```bash
make
```

## Dependencies

Your system must have the required libraries to mount an nfs volume. On Ubuntu, this is:

```bash
sudo apt install nfs-common
```

## Usage

First, start the server, pointing to the directory you want to share:

```bash
$ webfs serve /path/to/dir
```

Then, start the client, pointing to the server and the directory you want to mount the shared directory to:

```bash
$ webfs mount ws://server:port /path/to/mount
```

There's some configuration options available, see `webfs serve -h` and `webfs mount -h` for more details. Generally, you should be able to use the defaults, unless you have loopback-specific port requirements.

## Authx

Since this relies on websockets, we shift the authx responsibility up to the edge of the network, and assume you will be running this behind some identity-aware gateway such as Istio. When a client mounts a remote volume, they can pass either a `-token` or `-token-cmd` option to the `mount` command. If `-token` is passed, it will be used as the bearer token for the websocket connection. If `-token-cmd` is passed, it will be executed and the output will be used as the bearer token. This is useful if you want to use a dynamic token, such as a JWT, and don't want to have to manage the token yourself.

In the same vein, the example above shows the client connecting directly to the plain-text TCP port of the proxy server (`ws://server:port`) - in practice, it's assumed this will be run behind a TLS-terminating and identity-aware gateway, so your client will actually be connecting to something like `wss://webfs.example.com/my-mount-point`. If you have a direct TCP netpath from the client to the server on arbitrary ports in a trusted network, just connect to the NFS directly.

## How it works

Under the hood, this bundles a websocket server, tcp proxy, and nfs server on the server side, and a websocket client, tcp proxy, and nfs client on the client side. This enables us to tunnel nfs over websocket, relying on the security layers implemented at the websocket layer and tunnel through conventional gateways and firewalls. If you have a direct netpath to your data source or can use an object store, use that. This is to support a very specific requirement and topology.