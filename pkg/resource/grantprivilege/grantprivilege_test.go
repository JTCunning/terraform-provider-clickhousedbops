package grantprivilege_test

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/ClickHouse/terraform-provider-clickhousedbops/internal/dbops"
	"github.com/ClickHouse/terraform-provider-clickhousedbops/internal/testutils/nilcompare"
	"github.com/ClickHouse/terraform-provider-clickhousedbops/internal/testutils/resourcebuilder"
	"github.com/ClickHouse/terraform-provider-clickhousedbops/internal/testutils/runner"
	"github.com/ClickHouse/terraform-provider-clickhousedbops/pkg/resource/grantprivilege"
)

//go:embed grants.tsv
var testGrantsTSV string

const (
	resourceType = "clickhousedbops_grant_privilege"
	resourceName = "foo"

	granteeRoleName = "grantee"
	granteeUserName = "user1"
)

func TestGrantprivilege_acceptance(t *testing.T) {
	clusterName := "cluster1"

	grantsGroups := grantprivilege.ParseGrantsTSV(testGrantsTSV).Groups
	grantsScopes := grantprivilege.ParseGrantsTSV(testGrantsTSV).Scopes

	buildGrantPrivilege := func(accessType string, attrs map[string]string, grantOption bool) dbops.GrantPrivilege {
		var database, table, column *string
		if v := attrs["database_name"]; v != "" {
			database = &v
		}
		if v := attrs["table_name"]; v != "" {
			table = &v
		}
		if v := attrs["column_name"]; v != "" {
			column = &v
		}

		var granteeUserName, granteeRoleName *string
		if v := attrs["grantee_user_name"]; v != "" {
			granteeUserName = &v
		}
		if v := attrs["grantee_role_name"]; v != "" {
			granteeRoleName = &v
		}

		grant := dbops.GrantPrivilege{
			AccessType:          accessType,
			ExpandedAccessTypes: grantprivilege.AllDescendants(grantsGroups, accessType),
			DatabaseName:        database,
			TableName:           table,
			ColumnName:          column,
			GranteeUserName:     granteeUserName,
			GranteeRoleName:     granteeRoleName,
			GrantOption:         grantOption,
		}

		scope := grantsScopes[accessType]
		switch scope {
		case "USER_NAME", "DEFINER", "TABLE_ENGINE", "NAMED_COLLECTION":
			grant.UsesAccessObject = true
			for _, key := range []string{"user_name", "definer_name", "table_engine_name", "named_collection_name"} {
				if v := attrs[key]; v != "" {
					grant.AccessObject = v
					break
				}
			}
		}

		return grant
	}

	granteeRoleResource := resourcebuilder.
		New("clickhousedbops_role", granteeRoleName).
		WithStringAttribute("name", granteeRoleName)
	granteeUserResource := resourcebuilder.
		New("clickhousedbops_user", granteeUserName).
		WithStringAttribute("name", granteeUserName).
		WithFunction("password_sha256_hash_wo", "sha256", "test").
		WithIntAttribute("password_sha256_hash_wo_version", 1)

	checkNotExistsFunc := func(ctx context.Context, dbopsClient dbops.Client, clusterName *string, attrs map[string]string) (bool, error) {
		accessType := attrs["privilege_name"]
		if accessType == "" {
			return false, fmt.Errorf("privilege_name attribute was not set")
		}

		granteeUser := attrs["grantee_user_name"]
		granteeRole := attrs["grantee_role_name"]

		if granteeUser == "" && granteeRole == "" {
			return false, fmt.Errorf("both grantee_user_name and grantee_role_name attribute were not set")
		}

		grantPrivilege := buildGrantPrivilege(accessType, attrs, false)

		grantprivilege, err := dbopsClient.GetGrantPrivilege(ctx, &grantPrivilege, clusterName)
		return grantprivilege != nil, err
	}

	checkAttributesFunc := func(ctx context.Context, dbopsClient dbops.Client, clusterName *string, attrs map[string]any) error {
		accessType := attrs["privilege_name"].(string)
		if accessType == "" {
			return fmt.Errorf("privilege_name attribute was not set")
		}

		if stringAttr(attrs, "grantee_user_name") == "" && stringAttr(attrs, "grantee_role_name") == "" {
			return fmt.Errorf("both grantee_user_name and grantee_role_name attribute were not set")
		}

		grantOption := false
		if attrs["grant_option"] != nil {
			grantOption = attrs["grant_option"].(bool)
		}

		stringAttrs := map[string]string{
			"privilege_name":        accessType,
			"database_name":         stringAttr(attrs, "database_name"),
			"table_name":            stringAttr(attrs, "table_name"),
			"column_name":           stringAttr(attrs, "column_name"),
			"user_name":             stringAttr(attrs, "user_name"),
			"definer_name":          stringAttr(attrs, "definer_name"),
			"table_engine_name":     stringAttr(attrs, "table_engine_name"),
			"named_collection_name": stringAttr(attrs, "named_collection_name"),
			"grantee_user_name":     stringAttr(attrs, "grantee_user_name"),
			"grantee_role_name":     stringAttr(attrs, "grantee_role_name"),
		}

		grantPrivilege := buildGrantPrivilege(accessType, stringAttrs, grantOption)

		grantprivilege, err := dbopsClient.GetGrantPrivilege(ctx, &grantPrivilege, clusterName)
		if err != nil {
			return err
		}

		if grantprivilege == nil {
			return fmt.Errorf("grantprivilege was not found")
		}

		if attrs["privilege_name"].(string) != grantprivilege.AccessType {
			return fmt.Errorf("expected privilege_name to be %q, was %q", grantprivilege.AccessType, attrs["privilege_name"].(string))
		}

		if !nilcompare.NilCompare(grantprivilege.DatabaseName, attrs["database_name"]) {
			return fmt.Errorf("wrong value for database attribute")
		}

		if !nilcompare.NilCompare(grantprivilege.TableName, attrs["table_name"]) {
			return fmt.Errorf("wrong value for table attribute")
		}

		if !nilcompare.NilCompare(grantprivilege.ColumnName, attrs["column_name"]) {
			return fmt.Errorf("wrong value for column attribute")
		}

		if !nilcompare.NilCompare(clusterName, attrs["cluster_name"]) {
			return fmt.Errorf("wrong value for cluster_name attribute")
		}

		if !nilcompare.NilCompare(grantprivilege.GranteeUserName, attrs["grantee_user_name"]) {
			return fmt.Errorf("wrong value for grantee_user_name attribute")
		}

		if !nilcompare.NilCompare(grantprivilege.GranteeRoleName, attrs["grantee_role_name"]) {
			return fmt.Errorf("wrong value for grantee_role_name attribute")
		}

		if grantprivilege.GrantOption != grantOption {
			return fmt.Errorf("wrong value for grant_option attribute")
		}

		return nil
	}

	tests := []runner.TestCase{
		// Single replica, Native
		{
			Name:     "Grant global privilege to role using Native protocol on a single replica",
			ChEnv:    map[string]string{"CONFIGFILE": "config-single.xml"},
			Protocol: "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "SHOW USERS").
				WithResourceFieldReference("grantee_role_name", "clickhousedbops_role", granteeRoleName, "name").
				AddDependency(granteeRoleResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant privilege to user on a database using Native protocol on a single replica",
			ChEnv:    map[string]string{"CONFIGFILE": "config-single.xml"},
			Protocol: "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "CREATE TABLE").
				WithStringAttribute("database_name", "default").
				WithResourceFieldReference("grantee_user_name", "clickhousedbops_user", granteeUserName, "name").
				AddDependency(granteeUserResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant privilege to user on a table with grant option using Native protocol on a single replica",
			ChEnv:    map[string]string{"CONFIGFILE": "config-single.xml"},
			Protocol: "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "SELECT").
				WithStringAttribute("database_name", "system").
				WithStringAttribute("table_name", "databases").
				WithResourceFieldReference("grantee_user_name", "clickhousedbops_user", granteeUserName, "name").
				WithBoolAttribute("grant_option", true).
				AddDependency(granteeUserResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant global parent privilege ACCESS MANAGEMENT to role using Native protocol on a single replica",
			ChEnv:    map[string]string{"CONFIGFILE": "config-single.xml"},
			Protocol: "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "ACCESS MANAGEMENT").
				WithResourceFieldReference("grantee_role_name", "clickhousedbops_role", granteeRoleName, "name").
				AddDependency(granteeRoleResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant parent privilege CREATE on a database to role using Native protocol on a single replica",
			ChEnv:    map[string]string{"CONFIGFILE": "config-single.xml"},
			Protocol: "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "CREATE").
				WithStringAttribute("database_name", "default").
				WithResourceFieldReference("grantee_role_name", "clickhousedbops_role", granteeRoleName, "name").
				AddDependency(granteeRoleResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant CREATE USER on all users to role using Native protocol on a single replica",
			ChEnv:    map[string]string{"CONFIGFILE": "config-single.xml"},
			Protocol: "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "CREATE USER").
				WithResourceFieldReference("grantee_role_name", "clickhousedbops_role", granteeRoleName, "name").
				AddDependency(granteeRoleResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant CREATE USER on user name pattern to role using Native protocol on a single replica",
			ChEnv:    map[string]string{"CONFIGFILE": "config-single.xml"},
			Protocol: "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "CREATE USER").
				WithStringAttribute("user_name", "session-*").
				WithResourceFieldReference("grantee_role_name", "clickhousedbops_role", granteeRoleName, "name").
				AddDependency(granteeRoleResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant ALTER ROLE on role name to role using Native protocol on a single replica",
			ChEnv:    map[string]string{"CONFIGFILE": "config-single.xml"},
			Protocol: "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "ALTER ROLE").
				WithStringAttribute("user_name", "app_role").
				WithResourceFieldReference("grantee_role_name", "clickhousedbops_role", granteeRoleName, "name").
				AddDependency(granteeRoleResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		// Single replica, Native, wildcard database
		{
			Name:     "Grant privilege on wildcard database to role using Native protocol on a single replica",
			ChEnv:    map[string]string{"CONFIGFILE": "config-single.xml"},
			Protocol: "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "SELECT").
				WithStringAttribute("database_name", "test_prefix_*").
				WithResourceFieldReference("grantee_role_name", "clickhousedbops_role", granteeRoleName, "name").
				AddDependency(granteeRoleResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		// Single replica, HTTP
		{
			Name:     "Grant privilege on single column to role using HTTP protocol on a single replica",
			ChEnv:    map[string]string{"CONFIGFILE": "config-single.xml"},
			Protocol: "http",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "SELECT").
				WithStringAttribute("database_name", "system").
				WithStringAttribute("table_name", "databases").
				WithStringAttribute("column_name", "name").
				WithResourceFieldReference("grantee_role_name", "clickhousedbops_role", granteeRoleName, "name").
				AddDependency(granteeRoleResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant global privilege to user using HTTP protocol on a single replica",
			ChEnv:    map[string]string{"CONFIGFILE": "config-single.xml"},
			Protocol: "http",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "SHOW TABLES").
				WithResourceFieldReference("grantee_user_name", "clickhousedbops_user", granteeUserName, "name").
				AddDependency(granteeUserResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant privilege on database to user with grant option using HTTP protocol on a single replica",
			ChEnv:    map[string]string{"CONFIGFILE": "config-single.xml"},
			Protocol: "http",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "CREATE TABLE").
				WithStringAttribute("database_name", "default").
				WithResourceFieldReference("grantee_user_name", "clickhousedbops_user", granteeUserName, "name").
				WithBoolAttribute("grant_option", true).
				AddDependency(granteeUserResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		// Replicated storage, native
		{
			Name:     "Grant privilege on table to role using Native protocol on a cluster using replicated storage",
			ChEnv:    map[string]string{"CONFIGFILE": "config-replicated.xml"},
			Protocol: "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "SELECT").
				WithStringAttribute("database_name", "system").
				WithStringAttribute("table_name", "databases").
				WithResourceFieldReference("grantee_role_name", "clickhousedbops_role", granteeRoleName, "name").
				AddDependency(granteeRoleResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant privilege on column to user using Native protocol on a cluster using replicated storage",
			ChEnv:    map[string]string{"CONFIGFILE": "config-replicated.xml"},
			Protocol: "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "SELECT").
				WithStringAttribute("database_name", "system").
				WithStringAttribute("table_name", "databases").
				WithStringAttribute("column_name", "name").
				WithResourceFieldReference("grantee_user_name", "clickhousedbops_user", granteeUserName, "name").
				AddDependency(granteeUserResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant global privilege to user with grant option using Native protocol on a cluster using replicated storage",
			ChEnv:    map[string]string{"CONFIGFILE": "config-replicated.xml"},
			Protocol: "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "SHOW QUOTAS").
				WithResourceFieldReference("grantee_user_name", "clickhousedbops_user", granteeUserName, "name").
				WithBoolAttribute("grant_option", true).
				AddDependency(granteeUserResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		// Replicated storage, http
		{
			Name:     "Grant privilege on database to role using HTTP protocol on a cluster using replicated storage",
			ChEnv:    map[string]string{"CONFIGFILE": "config-replicated.xml"},
			Protocol: "http",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "CREATE TABLE").
				WithStringAttribute("database_name", "default").
				WithResourceFieldReference("grantee_role_name", "clickhousedbops_role", granteeRoleName, "name").
				AddDependency(granteeRoleResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant privilege on table to user using HTTP protocol on a cluster using replicated storage",
			ChEnv:    map[string]string{"CONFIGFILE": "config-replicated.xml"},
			Protocol: "http",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "SELECT").
				WithStringAttribute("database_name", "system").
				WithStringAttribute("table_name", "databases").
				WithResourceFieldReference("grantee_user_name", "clickhousedbops_user", granteeUserName, "name").
				AddDependency(granteeUserResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:     "Grant privilege on column to user with grant option using HTTP protocol on a cluster using replicated storage",
			ChEnv:    map[string]string{"CONFIGFILE": "config-replicated.xml"},
			Protocol: "http",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("privilege_name", "SELECT").
				WithStringAttribute("database_name", "system").
				WithStringAttribute("table_name", "databases").
				WithStringAttribute("column_name", "name").
				WithResourceFieldReference("grantee_user_name", "clickhousedbops_user", granteeUserName, "name").
				WithBoolAttribute("grant_option", true).
				AddDependency(granteeUserResource.Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		// Localfile storage, native
		{
			Name:        "Grant global privilege to role using Native protocol on a cluster using localfile storage",
			ChEnv:       map[string]string{"CONFIGFILE": "config-localfile.xml"},
			ClusterName: &clusterName,
			Protocol:    "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("cluster_name", clusterName).
				WithStringAttribute("privilege_name", "SHOW ACCESS").
				WithResourceFieldReference("grantee_role_name", "clickhousedbops_role", granteeRoleName, "name").
				AddDependency(granteeRoleResource.WithStringAttribute("cluster_name", clusterName).Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:        "Grant privilege on database to user using Native protocol on a cluster using localfile storage",
			ChEnv:       map[string]string{"CONFIGFILE": "config-localfile.xml"},
			ClusterName: &clusterName,
			Protocol:    "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("cluster_name", clusterName).
				WithStringAttribute("privilege_name", "DROP TABLE").
				WithStringAttribute("database_name", "default").
				WithResourceFieldReference("grantee_user_name", "clickhousedbops_user", granteeUserName, "name").
				AddDependency(granteeUserResource.WithStringAttribute("cluster_name", clusterName).Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:        "Grant privilege on table to user with grant option using Native protocol on a cluster using localfile storage",
			ChEnv:       map[string]string{"CONFIGFILE": "config-localfile.xml"},
			ClusterName: &clusterName,
			Protocol:    "native",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("cluster_name", clusterName).
				WithStringAttribute("privilege_name", "SELECT").
				WithStringAttribute("database_name", "system").
				WithStringAttribute("table_name", "databases").
				WithResourceFieldReference("grantee_user_name", "clickhousedbops_user", granteeUserName, "name").
				WithBoolAttribute("grant_option", true).
				AddDependency(granteeUserResource.WithStringAttribute("cluster_name", clusterName).Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		// Localfile storage, http
		{
			Name:        "Grant privilege on column to role using HTTP protocol on a cluster using localfile storage",
			ChEnv:       map[string]string{"CONFIGFILE": "config-localfile.xml"},
			ClusterName: &clusterName,
			Protocol:    "http",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("cluster_name", clusterName).
				WithStringAttribute("privilege_name", "SELECT").
				WithStringAttribute("database_name", "system").
				WithStringAttribute("table_name", "tables").
				WithStringAttribute("column_name", "name").
				WithResourceFieldReference("grantee_role_name", "clickhousedbops_role", granteeRoleName, "name").
				AddDependency(granteeRoleResource.WithStringAttribute("cluster_name", clusterName).Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:        "Grant global privilege to user using HTTP protocol on a cluster using localfile storage",
			ChEnv:       map[string]string{"CONFIGFILE": "config-localfile.xml"},
			ClusterName: &clusterName,
			Protocol:    "http",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("cluster_name", clusterName).
				WithStringAttribute("privilege_name", "SYSTEM RELOAD CONFIG").
				WithResourceFieldReference("grantee_user_name", "clickhousedbops_user", granteeUserName, "name").
				AddDependency(granteeUserResource.WithStringAttribute("cluster_name", clusterName).Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
		{
			Name:        "Grant privilege on database to user with grant option using HTTP protocol on a cluster using localfile storage",
			ChEnv:       map[string]string{"CONFIGFILE": "config-localfile.xml"},
			ClusterName: &clusterName,
			Protocol:    "http",
			Resource: resourcebuilder.New(resourceType, resourceName).
				WithStringAttribute("cluster_name", clusterName).
				WithStringAttribute("privilege_name", "DROP VIEW").
				WithStringAttribute("database_name", "default").
				WithResourceFieldReference("grantee_user_name", "clickhousedbops_user", granteeUserName, "name").
				WithBoolAttribute("grant_option", true).
				AddDependency(granteeUserResource.WithStringAttribute("cluster_name", clusterName).Build()).
				Build(),
			ResourceName:        resourceName,
			ResourceAddress:     fmt.Sprintf("%s.%s", resourceType, resourceName),
			CheckNotExistsFunc:  checkNotExistsFunc,
			CheckAttributesFunc: checkAttributesFunc,
		},
	}

	runner.RunTests(t, tests)
}

func stringAttr(attrs map[string]any, key string) string {
	if attrs[key] == nil {
		return ""
	}
	return attrs[key].(string)
}
