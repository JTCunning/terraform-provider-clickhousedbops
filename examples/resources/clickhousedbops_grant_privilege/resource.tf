# Global privilege (ON *.*) granted to a role.
resource "clickhousedbops_grant_privilege" "show_users_to_role" {
  privilege_name    = "SHOW USERS"
  grantee_role_name = "my_role"
}

# Database-level privilege (ON db.*) granted to a user, with grant option.
resource "clickhousedbops_grant_privilege" "create_table_on_db" {
  privilege_name    = "CREATE TABLE"
  database_name     = "default"
  grantee_user_name = "my_user"
  grant_option      = true
}

# Wildcard database (ON prefix_*.*) granted to a role.
resource "clickhousedbops_grant_privilege" "select_on_prefix_dbs" {
  privilege_name    = "SELECT"
  database_name     = "analytics_*"
  grantee_role_name = "analyst"
}

# Table-level privilege.
resource "clickhousedbops_grant_privilege" "select_on_table" {
  privilege_name    = "SELECT"
  database_name     = "default"
  table_name        = "events"
  grantee_user_name = "my_user"
}

# Column-level privilege.
resource "clickhousedbops_grant_privilege" "select_on_column" {
  privilege_name    = "SELECT"
  database_name     = "default"
  table_name        = "tbl1"
  column_name       = "count"
  grantee_user_name = "my_user_name"
  grant_option      = true
}

# Privilege scoped to a specific cluster (self-hosted, non-replicated user_directory).
resource "clickhousedbops_grant_privilege" "select_on_cluster" {
  cluster_name      = "my_cluster"
  privilege_name    = "SELECT"
  database_name     = "default"
  table_name        = "events"
  grantee_user_name = "my_user"
}

# USER_NAME-scope grant: any user or role (ON *).
resource "clickhousedbops_grant_privilege" "create_user_any" {
  privilege_name    = "CREATE USER"
  grantee_role_name = "user_admin"
}

# USER_NAME-scope grant restricted by name pattern.
resource "clickhousedbops_grant_privilege" "create_user_session" {
  privilege_name    = "CREATE USER"
  user_name         = "session-*"
  grantee_role_name = "session_admin"
}

# USER_NAME-scope grant on a specific role name.
resource "clickhousedbops_grant_privilege" "alter_role_app" {
  privilege_name    = "ALTER ROLE"
  user_name         = "app_role"
  grantee_role_name = "app_admin"
}

# DEFINER-scope grant.
resource "clickhousedbops_grant_privilege" "set_definer" {
  privilege_name    = "SET DEFINER"
  definer_name      = "etl_user"
  grantee_role_name = "etl_admin"
}

# TABLE_ENGINE-scope grant.
resource "clickhousedbops_grant_privilege" "table_engine_mergetree" {
  privilege_name    = "TABLE ENGINE"
  table_engine_name = "MergeTree"
  grantee_user_name = "my_user"
}

# NAMED_COLLECTION-scope grant.
resource "clickhousedbops_grant_privilege" "named_collection_admin" {
  privilege_name        = "NAMED COLLECTION ADMIN"
  named_collection_name = "s3_credentials"
  grantee_role_name     = "s3_admin"
}
