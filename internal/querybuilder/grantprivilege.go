package querybuilder

import (
	"fmt"
	"strings"

	"github.com/pingcap/errors"
)

// GrantPrivilegeQueryBuilder is an interface to build GRANT SQL queries (already interpolated).
type GrantPrivilegeQueryBuilder interface {
	QueryBuilder
	WithDatabase(*string) GrantPrivilegeQueryBuilder
	WithTable(*string) GrantPrivilegeQueryBuilder
	WithColumn(*string) GrantPrivilegeQueryBuilder
	WithAccessObject(*string) GrantPrivilegeQueryBuilder
	WithGrantOption(bool) GrantPrivilegeQueryBuilder
	WithCluster(*string) GrantPrivilegeQueryBuilder
}

type grantPrivilegeQueryBuilder struct {
	accessType      string
	to              string
	database        *string
	table           *string
	column          *string
	accessObject    *string
	useAccessObject bool
	grantOption     bool
	clusterName     *string
}

func GrantPrivilege(accessType string, to string) GrantPrivilegeQueryBuilder {
	return &grantPrivilegeQueryBuilder{
		accessType: accessType,
		to:         to,
	}
}

func (q *grantPrivilegeQueryBuilder) WithDatabase(database *string) GrantPrivilegeQueryBuilder {
	q.database = database
	return q
}

func (q *grantPrivilegeQueryBuilder) WithTable(table *string) GrantPrivilegeQueryBuilder {
	q.table = table
	return q
}

func (q *grantPrivilegeQueryBuilder) WithColumn(column *string) GrantPrivilegeQueryBuilder {
	q.column = column
	return q
}

func (q *grantPrivilegeQueryBuilder) WithCluster(clusterName *string) GrantPrivilegeQueryBuilder {
	q.clusterName = clusterName
	return q
}

func (q *grantPrivilegeQueryBuilder) WithGrantOption(grantOption bool) GrantPrivilegeQueryBuilder {
	q.grantOption = grantOption
	return q
}

func (q *grantPrivilegeQueryBuilder) WithAccessObject(accessObject *string) GrantPrivilegeQueryBuilder {
	q.accessObject = accessObject
	q.useAccessObject = true
	return q
}

func (q *grantPrivilegeQueryBuilder) Build() (string, error) {
	if q.accessType == "" {
		return "", errors.New("AccessType cannot be empty")
	}
	if q.to == "" {
		return "", errors.New("To cannot be empty")
	}

	tokens := []string{
		"GRANT",
	}

	if q.clusterName != nil {
		tokens = append(tokens, "ON", "CLUSTER", quote(*q.clusterName))
	}

	// Privilege
	if q.column != nil && *q.column != "" {
		tokens = append(tokens, fmt.Sprintf("%s(%s)", q.accessType, backtick(*q.column)))
	} else {
		tokens = append(tokens, q.accessType)
	}

	// Target database/table or access object (user name, definer, etc.)
	{
		tokens = append(tokens, "ON")

		if q.useAccessObject {
			if q.database != nil || q.table != nil || (q.column != nil && *q.column != "") {
				return "", errors.New("database, table, and column must not be set when access object is used")
			}
			if q.accessObject == nil || *q.accessObject == "" || *q.accessObject == "*" {
				tokens = append(tokens, "*")
			} else {
				tokens = append(tokens, accessObjectToken(*q.accessObject))
			}
		} else if q.database != nil {
			if q.table != nil {
				tokens = append(tokens, fmt.Sprintf("%s.%s", identifierOrPattern(*q.database), identifierOrPattern(*q.table)))
			} else {
				tokens = append(tokens, fmt.Sprintf("%s.*", identifierOrPattern(*q.database)))
			}
		} else {
			tokens = append(tokens, "*.*")
		}
	}

	// Grantee
	{
		tokens = append(tokens, "TO")
		tokens = append(tokens, backtick(q.to))
	}

	if q.grantOption {
		tokens = append(tokens, "WITH GRANT OPTION")
	}

	return strings.Join(tokens, " ") + ";", nil
}
