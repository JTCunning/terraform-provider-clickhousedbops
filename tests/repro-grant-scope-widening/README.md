# Grant scope widening ordering bug — local reproducer

## Summary

Changing a `clickhousedbops_grant_privilege` set from per-table grants to a database-wide grant for the same user produces a Terraform plan that adds the broader grants and destroys the narrower ones. When the broader `GRANT … ON db.*` is applied before the per-table revokes run, ClickHouse records those revokes as **partial revokes** against the broad grant. The user ends up with `ON db.* EXCEPT t1, t2, …` and loses effective access on the originally listed tables (`Code: 497 ACCESS_DENIED`).

This fixture reproduces that failure mode deterministically against the in-repo docker-compose ClickHouse using provider `v1.10.0`.

## Affected versions

- **Provider:** `ClickHouse/clickhousedbops` `v1.10.0`
- **ClickHouse server:** `26.3.3.20` (from `SELECT version()` against the local docker-compose stack)
- **Terraform:** `>= 1.11` (required for write-only password attributes on `clickhousedbops_user`)

## Prerequisites

- Docker and Docker Compose
- Terraform `>= 1.11`
- This repository checked out

## Steps to reproduce

From the repository root:

```bash
cd tests
REPLICAS=1 CONFIGFILE=config-single.xml docker compose up -d clickhouse proxy_http
curl -u default:test 'http://127.0.0.1:8123/?query=SELECT%20version()'

cd repro-grant-scope-widening
./run.sh
```

The script:

1. Applies per-table `INSERT`/`SELECT` grants on `repro_db.t1`…`t5` for user `repro_grantee`.
2. Plans the scope-widening change (`grant_mode=database`) and checks for `Plan: 2 to add, 0 to change, 10 to destroy`.
3. Applies database-wide grants while table grants still exist (`grant_mode=both`).
4. Revokes the table grants against the broad grant (`grant_mode=database`, with `-refresh=false` so Terraform still issues the `REVOKE` statements).
5. Verifies `system.grants` and a denied `INSERT`.

## Expected output

Salient lines from a successful run (full transcript in `transcript.txt` after `./run.sh`):

```
Plan: 2 to add, 0 to change, 10 to destroy.
PASS: plan shape matches (2 add / 10 destroy)
PASS: 10 partial-revoke rows present
Code: 497. DB::Exception: repro_grantee: Not enough privileges. ... (ACCESS_DENIED)
PASS: INSERT denied with ACCESS_DENIED
===== ALL PASS — bug reproduced =====
```

After phase 2, `system.grants` for `repro_grantee` contains ten rows with `is_partial_revoke = 1` (one per table × privilege) plus two database-wide rows with `is_partial_revoke = 0`. See [`expected-output.txt`](expected-output.txt) for the canonical excerpt.

## What this proves

- **Plan shape:** widening scope from per-table to database-wide yields `2 to add, 10 to destroy` for this fixture.
- **Partial revokes:** after broad grants exist, per-table revokes leave `is_partial_revoke = 1` rows in `system.grants`.
- **Access loss:** the grantee cannot `INSERT` into a table that was originally granted explicitly.

## Notes on apply ordering

A single `terraform apply` when switching `for_each` keys may interleave creates and destroys under default parallelism; the exact SQL order is not guaranteed. The reproducer uses a three-step apply (`table` → `both` → `database` with `-refresh=false`) to deterministically exercise the failure mode: broad grant first, then table revokes against it.

## Cleanup

```bash
terraform destroy -auto-approve
cd ..
docker compose down -v
```
