# Integration tests

End-to-end tests that drive the `onlineconf-admin` docker-compose stack
(admin server + MySQL seeded from `admin/etc/*.sql`) over its HTTP API,
asserting behavior of the move / soft-delete / change-log / notification
paths. The only host requirements are Go and `docker compose`.

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

## Cases

| Test | What it checks |
|---|---|
| `TestSmoke` | Stack up; auth + CSRF guard; core read endpoints return 200. |
| `TestMoveHistoricalPath` | A move records each version's path; pre-move entries stay under the old path, the move entry under the new one. |
| `TestMoveOntoDeletedPath` | Moving onto a soft-deleted path renames the tombstone aside but `/log/<path>` still shows the former occupant's history alongside the new one. |
| `TestSubtreeMoveLogging` | A subtree move logs every live descendant at its new path, not just the root. |
| `TestNotifyOnCommit` | Legacy (notifyDB) notifications fire only on commit — a rejected write emits none; a subtree move notifies once, for the root only. |
| `TestBotapiSubtreeMoveSilent` | The botapi feed (onlineconf-bot delivery) serves one notification per subtree move: descendant log entries are marked `Silent` and skipped. |
| `TestDeleteReferencedBySymlink` | A parameter referenced by a live symlink cannot be deleted; it can once the symlink is gone. |
| `TestDeleteReferencedThroughChain` | A symlink resolved through another symlink protects both the intermediate hop and the terminal node. |
| `TestTemplateThroughSymlinkBlocksHop` | A template `${<symlink>/sub}` protects the symlink it resolves through and the terminal (tightening over the legacy check). |
| `TestCaseReferrerBlocks` | A symlink inside a case branch protects its target. |
| `TestNestedCaseReferrer` | A symlink buried in a nested case protects its target. |
| `TestHealNestedCaseTemplate` | A dangling template branch inside a nested case is healed when its path is created. |
| `TestMoveReferencedRequiresSymlink` | Moving a referenced parameter is refused without a symlink; with one it succeeds atomically and the referrer keeps resolving. |
| `TestMoveReferencedSubtreeRefused` | A referrer to a descendant blocks a no-symlink move of an ancestor subtree. |
| `TestDeepChainTrackedAtDefaultDepth` | Dependencies are tracked through a 6-hop directory-symlink chain (deeper than the old default of 5). |
| `TestDepthTunableViaParameter` | `/onlineconf/deleted-key-symlinks-check-depth` controls the depth live: raising it rebuilds the edge table, lowering it does not. |
| `TestHealAfterDisabledDeleteRecreate` | A subtree destroyed with the check disabled and recreated is protected again, including through intermediate symlinks (dangling points + dependent recompute). |
| `TestFailClosedBeforeBackfill` | While the dependency table has never been built, deletions are refused with a retriable 503 instead of passing unchecked. |
| `TestDepthTruncationLogged` | A dependency walk abandoned at the depth cap logs a warning naming the depth parameter. |
| `TestSelfReferencingSymlinks` | Symlinks to their own parent and to the root protect the target without deadlocking their own deletion. |
| `TestHealDanglingTemplate` | A template referencing a not-yet-existing path protects it as soon as it is created. |
| `TestBackfillProtectsExampleData` | The startup backfill populates dependency edges for a pre-existing database. |
| `TestDisableFlagSkipsCheck` | The `disable-deleted-key-symlinks-check` flag still bypasses the check. |
