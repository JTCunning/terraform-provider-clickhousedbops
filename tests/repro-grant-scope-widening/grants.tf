variable "grant_mode" {
  type        = string
  description = "table = per-table grants; both = table + database-wide grants; database = database-wide grants only"
  default     = "table"

  validation {
    condition     = contains(["table", "both", "database"], var.grant_mode)
    error_message = "grant_mode must be one of: table, both, database."
  }
}

locals {
  tables = toset(["t1", "t2", "t3", "t4", "t5"])
  privs  = toset(["INSERT", "SELECT"])

  table_grants = {
    for p in setproduct(tolist(local.privs), tolist(local.tables)) :
    "${p[0]}_${p[1]}" => {
      privilege_name = p[0]
      table_name     = p[1]
    }
  }

  database_grants = {
    for priv in local.privs :
    "${priv}_ALL" => {
      privilege_name = priv
      table_name     = null
    }
  }

  grants = (
    var.grant_mode == "table" ? local.table_grants :
    var.grant_mode == "both" ? merge(local.table_grants, local.database_grants) :
    local.database_grants
  )
}

resource "clickhousedbops_grant_privilege" "this" {
  for_each          = local.grants
  privilege_name    = each.value.privilege_name
  database_name     = clickhousedbops_database.db.name
  table_name        = each.value.table_name
  grantee_user_name = clickhousedbops_user.grantee.name
}
