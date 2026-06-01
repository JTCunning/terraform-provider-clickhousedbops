#!/usr/bin/env bash
set -euo pipefail
exec > >(tee transcript.txt) 2>&1

CH=(curl -sS -u default:test 'http://127.0.0.1:8123/' --data-binary)

echo "===== Phase 1: table-level grants ====="
terraform init -input=false
terraform apply -auto-approve -input=false \
  -target=clickhousedbops_database.db \
  -target=clickhousedbops_user.grantee

for t in t1 t2 t3 t4 t5; do
  "${CH[@]}" "CREATE TABLE IF NOT EXISTS repro_db.${t} (id UInt64, ts DateTime) ENGINE=MergeTree ORDER BY id"
done

terraform apply -auto-approve -input=false -var='grant_mode=table'

echo "===== Phase 2: plan shape when widening scope (table -> database) ====="
terraform plan -no-color -input=false -var='grant_mode=database' | tee phase2.plan.txt
grep -E '^Plan: 2 to add, 0 to change, 10 to destroy' phase2.plan.txt \
  && echo "PASS: plan shape matches (2 add / 10 destroy)" \
  || { echo "FAIL: unexpected plan shape"; exit 1; }

echo "===== Phase 2: apply broad grants while table grants still exist ====="
terraform apply -auto-approve -input=false -var='grant_mode=both'

echo "===== Phase 2: revoke table grants against the broad grant ====="
terraform apply -refresh=false -auto-approve -input=false -var='grant_mode=database'

echo "===== system.grants for repro_grantee ====="
"${CH[@]}" "SELECT user_name, access_type, database, table, is_partial_revoke
            FROM system.grants
            WHERE user_name = 'repro_grantee'
            ORDER BY access_type, table FORMAT PrettyCompact" | tee grants.txt
partial_revoke_count=$("${CH[@]}" "SELECT count() FROM system.grants WHERE user_name = 'repro_grantee' AND is_partial_revoke = 1 FORMAT TabSeparatedRaw")
if [[ "${partial_revoke_count}" -ge 10 ]]; then
  echo "PASS: ${partial_revoke_count} partial-revoke rows present"
else
  echo "FAIL: expected >= 10 partial-revoke rows, got ${partial_revoke_count}"
  exit 1
fi

echo "===== INSERT as repro_grantee (expect 497) ====="
set +e
insert_out=$(curl -sS -u 'repro_grantee:repro-pw' \
  'http://127.0.0.1:8123/' --data-binary \
  "INSERT INTO repro_db.t1 (id, ts) VALUES (1, now())" 2>&1)
insert_rc=$?
set -e
echo "${insert_out}"
if echo "${insert_out}" | grep -q 'ACCESS_DENIED'; then
  echo "PASS: INSERT denied with ACCESS_DENIED"
else
  echo "FAIL: INSERT was not denied (exit ${insert_rc})"
  exit 1
fi

echo "===== ALL PASS — bug reproduced ====="
