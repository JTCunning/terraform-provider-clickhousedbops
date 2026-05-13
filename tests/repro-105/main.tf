terraform {
  required_providers {
    clickhousedbops = {
      version = "1.10.0"
      source  = "ClickHouse/clickhousedbops"
    }
  }
}

provider "clickhousedbops" {
  host     = "127.0.0.1"
  protocol = "http"
  port     = 8123

  auth_config = {
    strategy = "basicauth"
    username = "default"
    password = "test"
  }
}

resource "clickhousedbops_database" "example" {
  name = "example105"
}

resource "clickhousedbops_role" "admin" {
  name = "admin"
}

locals {
  privileges = toset([
    "SELECT",
    "INSERT",
    "ALTER",
    "OPTIMIZE",
    "CREATE",
    "DROP",
    "SHOW",
    "TRUNCATE",
    "KILL QUERY",
    "SYSTEM",
  ])
}

resource "clickhousedbops_grant_privilege" "role_grants" {
  for_each          = local.privileges
  privilege_name    = each.value
  database_name     = clickhousedbops_database.example.name
  grantee_role_name = clickhousedbops_role.admin.name
  grant_option      = false
}
