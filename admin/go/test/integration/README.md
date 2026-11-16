# Integration tests

End-to-end tests that drive the `onlineconf-admin` docker-compose stack
(admin server + MySQL seeded from `admin/etc/*.sql`) over its HTTP API,
asserting behavior of the move / soft-delete / change-log / notification
paths and of value encryption (the key file, reveals, sealing, restores,
rotation). The only host requirements are Go and `docker compose`.

The suite is separated from the unit tests with build tags: these files build
only with the `integration` tag, while unit tests carry `!integration` — so
each invocation runs exactly one suite:

```sh
cd admin/go
go test ./...                                            # unit tests only
go test -tags integration -count=1 -v ./test/integration # integration tests only
```

(`go test -tags integration ./...` also works: with the anti-tag the unit
tests are excluded from such a run.)

By default `TestMain` recreates the stack with a fresh database
(`docker compose down -v && up -d --build`), runs the tests, and tears the
stack down. Environment knobs:

| Variable | Effect |
|---|---|
| `ONLINECONF_NO_COMPOSE=1` | Don't manage the stack — assume it is already running and leave it up afterwards. |
| `ONLINECONF_URL` | Base URL of the admin server (default `http://localhost`). |

Auth uses the `admin`/`admin` demo account from `admin/etc/example-auth.sql`;
the botapi test provisions its own bot account through the API. Parameter
paths are suffixed with the test process pid, so re-runs against a persistent
stack do not collide. Database assertions go through
`docker compose exec onlineconf-database mysql` since the MySQL port is not
published on the host.

The encryption tests need a fresh database per run, also with
`ONLINECONF_NO_COMPOSE=1`: `TestKeyRotationSweep` re-keys the stack's data
and leaves a new key in the admin container's key file, and a container
re-created from the image reinstalls the demo key, which then fails closed on
that data — `docker compose down -v` before the next run.
`TestStartupWithoutKeyFails` starts a second admin from the serving
container's image with `docker run` on the stack's network.

## Cases

| Test | What it checks |
|---|---|
| `TestSmoke` | Stack up; auth + CSRF guard; core read endpoints return 200. |
| `TestMoveHistoricalPath` | A move records each version's path; pre-move entries stay under the old path, the move entry under the new one. |
| `TestMoveOntoDeletedPath` | Moving onto a soft-deleted path renames the tombstone aside but `/log/<path>` still shows the former occupant's history alongside the new one. |
| `TestSubtreeMoveLogging` | A subtree move logs every live descendant at its new path, not just the root. |
| `TestNotifyOnCommit` | Legacy (notifyDB) notifications fire only on commit — a rejected write emits none; a subtree move notifies once, for the root only. |
| `TestBotapiSubtreeMoveSilent` | The botapi feed (onlineconf-bot delivery) serves one notification per subtree move: descendant log entries are marked `Silent` and skipped. |
| `TestControlParameterSymlinks` | A setting under `/onlineconf` may be a symlink to a plain text; a symlink to an encrypted value, and a cycle, count as absent. |
| `TestAuditKinds` | The Go audit kinds and the `my_config_audit.Kind` enum stay in step. |
| `TestEncryptedCRUD` | Create, mask, reveal, set and declassify an encrypted parameter; the mandatory `encrypted` field; the audit rows. |
| `TestCaseBranchesNotEncrypted` | A case branch marked encrypted is refused; a symlink branch to an encrypted parameter is served plain, delivered with the plaintext, and protects its target from deletion. |
| `TestRevealPolicy` | The `rw` and `ro` reveal policies for a read-only user; `/ui-config` serves no policy. |
| `TestRevealFeed` | The botapi audit feed serves reveal events to a bot with the `audit` scope. |
| `TestEncryptedDelivery` | `/client/config` delivers the plaintext of encrypted values to updaters. |
| `TestSearchSkipsEncrypted` | Value search never matches an envelope. |
| `TestNotificationsRedacted` | Change notifications and the botapi feed show a lock instead of the secret. |
| `TestHistorySealedOnFlip` | Every earlier version is sealed when a parameter becomes encrypted; the history reveals on demand with one audit row. |
| `TestHistorySealedByType` | Sealing touches versions of eligible types only; sealed rows stay as they are. |
| `TestRestore` | A rollback by log entry id; a deletion entry, another path's entry and a stale version are refused. |
| `TestRestoreFormerOccupant` | A former occupant's entry restores under reveal rights and is refused to a user who may not reveal it. |
| `TestRestoreByType` | A null, symlink, case or template version restored onto an encrypted parameter comes back plain; a vanished symlink target is refused. |
| `TestRestoreDeletedParameter` | A deleted parameter comes back from its history without a version, encrypted as the version was, a new node at a moved-away path; gated by the switch and the parent's write access. |
| `TestEncryptionGated` | The feature switch gates encrypted writes, restores included; the `encrypted` field is `0` or `1`; `/ui-config` without access to the avatar. |
| `TestTemplatesNotEncrypted` | A template cannot be encrypted; a plaintext template referencing an encrypted parameter is delivered expanded and protects its target from deletion. |
| `TestStartupWithoutKeyFails` | An instance without a key file refuses to start against encrypted parameters. |
| `TestKeyRotationSweep` | A rotation through `/onlineconf/value-encryption/key-id` with a racing writer, back and forth, an unknown key refused, a restart with the new key alone. |
| `TestSchemaVersionGuard` | A schema behind the binary is fatal and names the update file; a newer one only warns. |
