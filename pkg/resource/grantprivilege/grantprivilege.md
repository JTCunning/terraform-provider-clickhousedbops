You can use the `clickhousedbops_grant_privilege` resource to grant privileges on databases and tables to either a `clickhousedbops_user` or a `clickhousedbops_role`.

Please note that in order to grant privileges to all database and/or all tables, the `database_name` and/or `table_name` fields must be set to null, and not to "*".

Some privileges target a user, role, definer, table engine, or named collection instead of a database or table. Use the matching optional attribute and leave `database_name`, `table_name`, and `column_name` null:

- `user_name` for privileges such as `CREATE USER`, `ALTER ROLE`, and `IMPERSONATE`
- `definer_name` for `SET DEFINER`
- `table_engine_name` for `TABLE ENGINE`
- `named_collection_name` for named-collection privileges such as `NAMED COLLECTION ADMIN`

Leave the matching attribute null to grant on all targets (`ON *`). Trailing `*` wildcards (for example `session-*`) are supported on these fields, same as on `database_name` and `table_name`.

Known limitations:

- Only a subset of privileges can be granted on ClickHouse cloud. For example the `ALL` privilege can't be granted. See https://clickhouse.com/docs/en/sql-reference/statements/grant#all
- It's not possible to grant privileges using their alias name. The canonical name must be used.
- It's not possible to grant group of privileges. Please grant each member of the group individually instead.
- It's not possible to grant the same `clickhousedbops_grant_privilege` to both a `clickhousedbops_user` and a `clickhousedbops_role` using a single `clickhousedbops_grant_privilege` stanza. You can do that using two different stanzas, one with `grantee_user_name` and the other with `grantee_role_name` fields set.
- It's not possible to grant the same privilege (example 'SELECT') to multiple entities (for example tables) with a single stanza. You can do that my creating one stanza for each entity you want to grant privileges on.
- Importing `clickhousedbops_grant_privilege` resources into terraform is not supported.
