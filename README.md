# Auth Manager

A generic account and auth management backend

Design patterns heavily influenced by Alex Edwards [books](https://lets-go-further.alexedwards.net/).

## Features

* Teams
* Users
* Roles
* Tokens

## Components

* [Gin](https://github.com/gin-gonic/gin) is used as the router
* [Casbin](https://casbin.org/) as the auth backend 

## Run the server

Locally

```bash
make run-server
```

## Tests

Run the integration tests via the makefile

```bash
make ENV=int test-int-local
```
