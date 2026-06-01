terraform {
  required_providers {
    clickhousedbops = {
      source  = "ClickHouse/clickhousedbops"
      version = "1.10.0"
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

resource "clickhousedbops_database" "db" {
  name = "repro_db"
}

resource "clickhousedbops_user" "grantee" {
  name                            = "repro_grantee"
  password_sha256_hash_wo         = sha256("repro-pw")
  password_sha256_hash_wo_version = 1
}
