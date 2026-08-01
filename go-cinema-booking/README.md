# Cinema Booking

An experimental Go application for exploring concurrent seat reservations. The
project pairs a browser-based seat picker with a small booking domain that is
intended to guarantee that only one user can reserve a seat, even when many
requests arrive at the same time.

> [!NOTE]
> This project is a work in progress. The frontend describes the intended user
> experience, but the HTTP server, persistence layer, and booking operation are
> not implemented yet. See [Current status](#current-status) before trying to
> run it.

## Booking flow

The UI is designed around a short-lived hold rather than immediately confirming
a seat:

1. The browser loads the available movies and seats.
2. A user places a temporary hold on one available seat.
3. The UI displays the hold's expiry time and polls for seat changes every two
   seconds.
4. The user confirms the hold or releases it.
5. An unconfirmed hold becomes available again after it expires.

The important consistency rule is that a `(movie_id, seat_id)` pair can have at
most one active owner. The concurrency test models 100,000 users competing for
the same seat and expects exactly one booking attempt to succeed.

## Current status

| Area | State |
| --- | --- |
| Booking model and store interface | Present in `internal/booking/domain.go` |
| Booking service constructor | Present in `internal/booking/service.go` |
| Concurrent-booking test | Present, but does not compile until `Service.Book` and its dependency are added |
| Seat-picker frontend | Present in `static/index.html` |
| In-memory or database store | Not implemented |
| HTTP routes and static-file serving | Not implemented |
| Application entry point | Placeholder |

The frontend is therefore a contract/prototype for the backend, not currently a
standalone working application.

## Project layout

```text
.
├── cmd/main.go                      # Future application entry point
├── internal/booking/
│   ├── domain.go                    # Booking entity and store abstraction
│   ├── service.go                   # Booking service construction
│   └── service_test.go              # Concurrent reservation invariant
├── internal/utils/utils.go          # JSON response helper
├── static/index.html                # Seat-selection UI and API client
└── go.mod
```

Keeping the booking package under `internal/` prevents other modules from
depending directly on implementation details while still allowing the command
package to use them.

## Intended HTTP contract

The existing frontend expects the following endpoints:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/movies` | List movies and their seating dimensions |
| `GET` | `/movies/{movieID}/seats` | Get the current state of every occupied seat |
| `POST` | `/movies/{movieID}/seats/{seatID}/hold` | Create a temporary hold for a user |
| `PUT` | `/sessions/{sessionID}/confirm` | Confirm a user's active hold |
| `DELETE` | `/sessions/{sessionID}` | Release a user's active hold |

A hold request sends:

```json
{
  "user_id": "3c3c60ef715a"
}
```

The corresponding successful response must include the values used by the
countdown and checkout panel:

```json
{
  "session_id": "73553c96-5953-49c9-b6f1-559bb341d85d",
  "movie_id": "screen-1",
  "seat_id": "A1",
  "expires_at": "2026-07-31T18:05:00Z"
}
```

Errors should use a non-2xx status and a JSON body containing an `error` field:

```json
{
  "error": "seat is already held"
}
```

## Development

The module currently declares Go 1.25. Install a compatible Go toolchain, then
run commands from this directory:

```bash
go test ./...
```

At the current stage, that command is expected to fail because the test calls a
`Service.Book` method that has not been implemented and imports
`github.com/google/uuid`, which is not yet declared in `go.mod`. These failures
are tracked here so they are not mistaken for a setup problem.

To make the first backend slice executable:

1. Define the behavior and errors for `Service.Book`.
2. Implement a concurrency-safe `BookingStore` (an in-memory store is enough to
   start).
3. Replace the `nil` store in the concurrency test with that implementation.
4. Add or replace the UUID dependency used by the test.
5. Implement the HTTP routes above and serve `static/index.html` from
   `cmd/main.go`.

When the backend is wired up, run the race detector as part of verification:

```bash
go test -race ./...
```

The race-enabled test matters here: producing one logical winner is not enough
if the store reaches that result through unsynchronized map access or another
data race.
