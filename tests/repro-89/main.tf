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
  name = "example1"
}

resource "clickhousedbops_user" "example1" {
  name                            = "example1"
  password_sha256_hash_wo         = sha256("password")
  password_sha256_hash_wo_version = 1
}

resource "clickhousedbops_grant_privilege" "example" {
  privilege_name    = "ACCESS MANAGEMENT"
  database_name     = clickhousedbops_database.example.name
  grantee_user_name = clickhousedbops_user.example1.name
  grant_option      = true
}

resource "clickhousedbops_grant_privilege" "example_all" {
  privilege_name    = "ALL"
  database_name     = clickhousedbops_database.example.name
  grantee_user_name = clickhousedbops_user.example1.name
  grant_option      = true
}
