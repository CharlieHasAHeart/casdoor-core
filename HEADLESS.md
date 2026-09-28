# Headless Casdoor core

This repository is the YuanChuangLi-maintained Casdoor backend fork. It keeps
Casdoor's identity, credential, session, OIDC, OAuth, API, storage, and policy
logic, but deliberately removes the bundled `web/` and `web-old/` frontends.

## Repository boundary

`casdoor-core` is shared source code. It must not contain an application's
branding, login page, or business UI. Each product may maintain its own
authentication UI and connect to the same core through Casdoor's API and OIDC
contracts.

The canonical repository should live in the YuanChuangLi GitHub organization.
Local development checkouts belong at `~/workspace/casdoor-core`; product
repositories consume a pinned tag or commit instead of copying this source.
The local path is a workspace convenience, not the source of truth.

One source repository does not require one runtime instance. A project can run
its own Casdoor process and database from a pinned `casdoor-core` revision, or
several applications can share one issuer when shared identity and SSO are
intended.

## Headless runtime behavior

The default `conf/app.conf` enables `headless = true`. API routes, OIDC
discovery, token endpoints, JWKS, user persistence, and protocol handlers remain
available. Browser routes that would normally serve Casdoor's bundled UI return
`501 headless_ui_required` until an external authentication UI is configured.

This is intentional: a browser user still needs a Casdoor-owned login
interface for password, email verification, recovery, and other interactive
identity flows. That interface is maintained outside this core repository; it
must not turn Golden's business frontend into a password handler.

## Development

Install the Go toolchain required by `go.mod`, configure `conf/app.conf` to use
the project's Casdoor PostgreSQL database, then run:

```sh
go test ./...
go run main.go
```

The repository retains the official Casdoor source as the `upstream` Git remote.
Create and pin YuanChuangLi release tags before consuming this core from other
projects.

## Consuming the core from another project

The first integration should use a pinned Git submodule or a reproducible
bootstrap clone. Do not follow the default branch in application builds:

```sh
git submodule add https://github.com/<org>/casdoor-core.git third_party/casdoor-core
git -C third_party/casdoor-core checkout <yuanchuangli-tag-or-commit>
```

Each product keeps its own Casdoor configuration, database, OIDC client, and
login UI. Multiple products may run separate Casdoor instances from this same
source, or deliberately share one issuer when shared identity is required.
