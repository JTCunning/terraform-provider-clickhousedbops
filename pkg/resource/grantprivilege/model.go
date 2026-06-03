package grantprivilege

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GrantPrivilege struct {
	ClusterName         types.String `tfsdk:"cluster_name"`
	Privilege           types.String `tfsdk:"privilege_name"`
	Database            types.String `tfsdk:"database_name"`
	Table               types.String `tfsdk:"table_name"`
	Column              types.String `tfsdk:"column_name"`
	UserName            types.String `tfsdk:"user_name"`
	DefinerName         types.String `tfsdk:"definer_name"`
	TableEngineName     types.String `tfsdk:"table_engine_name"`
	NamedCollectionName types.String `tfsdk:"named_collection_name"`
	GranteeUserName     types.String `tfsdk:"grantee_user_name"`
	GranteeRoleName     types.String `tfsdk:"grantee_role_name"`
	GrantOption         types.Bool   `tfsdk:"grant_option"`
}
