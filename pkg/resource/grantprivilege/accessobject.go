package grantprivilege

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ClickHouse/terraform-provider-clickhousedbops/internal/dbops"
)

var scopeAccessObjectAttributes = map[string]string{
	"USER_NAME":         "user_name",
	"DEFINER":           "definer_name",
	"TABLE_ENGINE":      "table_engine_name",
	"NAMED_COLLECTION":  "named_collection_name",
}

func isGlobalWithParameterScope(scope string) bool {
	_, ok := scopeAccessObjectAttributes[scope]
	return ok
}

func accessObjectAttributeForScope(scope string) (string, bool) {
	attr, ok := scopeAccessObjectAttributes[scope]
	return attr, ok
}

func grantPrivilegeFromModel(model GrantPrivilege, groups map[string][]string) dbops.GrantPrivilege {
	privilegeName := model.Privilege.ValueString()
	scope := parsedGrants().Scopes[privilegeName]

	grant := dbops.GrantPrivilege{
		AccessType:          privilegeName,
		ExpandedAccessTypes: AllDescendants(groups, privilegeName),
		DatabaseName:        model.Database.ValueStringPointer(),
		TableName:           model.Table.ValueStringPointer(),
		ColumnName:          model.Column.ValueStringPointer(),
		GranteeUserName:     model.GranteeUserName.ValueStringPointer(),
		GranteeRoleName:     model.GranteeRoleName.ValueStringPointer(),
		GrantOption:         model.GrantOption.ValueBool(),
	}

	if isGlobalWithParameterScope(scope) {
		grant.UsesAccessObject = true
		if v := model.UserName.ValueString(); !model.UserName.IsNull() {
			grant.AccessObject = v
		} else if v := model.DefinerName.ValueString(); !model.DefinerName.IsNull() {
			grant.AccessObject = v
		} else if v := model.TableEngineName.ValueString(); !model.TableEngineName.IsNull() {
			grant.AccessObject = v
		} else if v := model.NamedCollectionName.ValueString(); !model.NamedCollectionName.IsNull() {
			grant.AccessObject = v
		}
	}

	return grant
}

func syncAccessObjectFields(state *GrantPrivilege, accessObject string, scope string) {
	state.UserName = types.StringNull()
	state.DefinerName = types.StringNull()
	state.TableEngineName = types.StringNull()
	state.NamedCollectionName = types.StringNull()

	if accessObject == "" {
		return
	}

	switch scope {
	case "USER_NAME":
		state.UserName = types.StringValue(accessObject)
	case "DEFINER":
		state.DefinerName = types.StringValue(accessObject)
	case "TABLE_ENGINE":
		state.TableEngineName = types.StringValue(accessObject)
	case "NAMED_COLLECTION":
		state.NamedCollectionName = types.StringValue(accessObject)
	}
}

func validateGlobalWithParameterPlan(plan GrantPrivilege, scope string, resp *resource.ModifyPlanResponse) {
	expectedAttr, ok := accessObjectAttributeForScope(scope)
	if !ok {
		return
	}

	if !plan.Database.IsNull() || !plan.Table.IsNull() || !plan.Column.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("privilege_name"),
			"Invalid Grant Privilege",
			fmt.Sprintf("'database_name', 'table_name', and 'column_name' must be null when privilege_name is %q", plan.Privilege.ValueString()),
		)
		return
	}

	type fieldCheck struct {
		attr  string
		value types.String
	}
	checks := []fieldCheck{
		{"user_name", plan.UserName},
		{"definer_name", plan.DefinerName},
		{"table_engine_name", plan.TableEngineName},
		{"named_collection_name", plan.NamedCollectionName},
	}

	setCount := 0
	for _, check := range checks {
		if check.value.IsNull() {
			continue
		}
		setCount++
		if check.attr != expectedAttr {
			resp.Diagnostics.AddAttributeError(
				path.Root(check.attr),
				"Invalid Grant Privilege",
				fmt.Sprintf("'%s' cannot be set when privilege_name is %q (scope %q)", check.attr, plan.Privilege.ValueString(), scope),
			)
		}
	}

	if setCount > 1 {
		resp.Diagnostics.AddError(
			"Invalid Grant Privilege",
			"Only one of user_name, definer_name, table_engine_name, and named_collection_name may be set",
		)
	}
}

func accessObjectFromPlan(plan GrantPrivilege) string {
	if !plan.UserName.IsNull() {
		return plan.UserName.ValueString()
	}
	if !plan.DefinerName.IsNull() {
		return plan.DefinerName.ValueString()
	}
	if !plan.TableEngineName.IsNull() {
		return plan.TableEngineName.ValueString()
	}
	if !plan.NamedCollectionName.IsNull() {
		return plan.NamedCollectionName.ValueString()
	}
	return ""
}

func accessObjectPointer(grant dbops.GrantPrivilege) *string {
	if !grant.UsesAccessObject || grant.AccessObject == "" {
		return nil
	}
	return &grant.AccessObject
}
