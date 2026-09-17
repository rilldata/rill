# pgwire

`pgwire` implements the server side of PostgreSQL protocol version 3 on top of
`github.com/jackc/pgx/v5/pgproto3`.

The package owns transport and connection-local protocol state: startup and TLS
negotiation, password authentication, cancellation, prepared statements,
portals, and simple and extended query flows. Product-specific query execution
is provided by a `Session` created for each connection.

Prepared statements and portals intentionally live on a connection, not on the
server. This matches PostgreSQL semantics and prevents state or credentials from
being shared between clients.

Result messages are flushed in batches of approximately 64 KiB, plus at most
one row, so unlimited Execute and simple queries do not accumulate the entire
encoded result in memory. Error recovery shares the normal Sync handler.
