# Bookstore backend API

A simple bookstore API

API Documents on `swagger.yaml`, view in the browser using [Swagger Editor](https://editor.swagger.io)

## Features

- Api is guarded behind session tokens
- Only administrators can edit data, users are only allowed to list and search
- Cover image upload and display

## Layout

- cmd/bookstore_server: serves as the entrypoint that glues everything together
- auth: the package responsible for authentication
- cover/fs: is responsible for storing the cover files into filesystem
- db/psql: is the underlying db client
- http/rest: is the http REST handler

## bookstore_server

ENV required:

- DATABASE_URL: the connection string used to connect to a psql db
- URL: the canonical webroot(used for the cover service)
- LISTEN: the address to listen on

Args:

- `--routes`: make the app dump out automatically generated markdown API routes
- `--debug-routes`: makes the app mount an unprotected route to manage users on `/api/v1/debug/users` for debugging,
  allows you to give yourself admin without the DB
- `--debug-isbn`: makes the app ignore ISBN checksum

Environment:

Postgres Server, require extension `uuid-ossp` and `pg_trgm`