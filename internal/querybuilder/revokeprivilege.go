package querybuilder

import (
	"fmt"
	"strings"

	"github.com/pingcap/errors"
)

// RevokePrivilegeQueryBuilder is an interface to build REVOKE SQL queries (already interpolated).
type RevokePrivilegeQueryBuilder interface {
	QueryBuilder
	WithDatabase(*string) RevokePrivilegeQueryBuilder
	WithTable(*string) RevokePrivilegeQueryBuilder
	WithColumn(*string) RevokePrivilegeQueryBuilder
	WithAccessObject(*string) RevokePrivilegeQueryBuilder
	WithCluster(*string) RevokePrivilegeQueryBuilder
}

type revokePrivilegeQueryBuilder struct {
	accessType      string
	from            string
	database        *string
	table           *string
	column          *string
	accessObject    *string
	useAccessObject bool
	clusterName     *string
}

func RevokePrivilege(accessType string, from string) RevokePrivilegeQueryBuilder {
	return &revokePrivilegeQueryBuilder{
		accessType: accessType,
		from:       from,
	}
}

func (q *revokePrivilegeQueryBuilder) WithDatabase(database *string) RevokePrivilegeQueryBuilder {
	q.database = database
	return q
}

func (q *revokePrivilegeQueryBuilder) WithTable(table *string) RevokePrivilegeQueryBuilder {
	q.table = table
	return q
}

func (q *revokePrivilegeQueryBuilder) WithColumn(column *string) RevokePrivilegeQueryBuilder {
	q.column = column
	return q
}

func (q *revokePrivilegeQueryBuilder) WithCluster(clusterName *string) RevokePrivilegeQueryBuilder {
	q.clusterName = clusterName
	return q
}

func (q *revokePrivilegeQueryBuilder) WithAccessObject(accessObject *string) RevokePrivilegeQueryBuilder {
	q.accessObject = accessObject
	q.useAccessObject = true
	return q
}

func (q *revokePrivilegeQueryBuilder) Build() (string, error) {
	if q.accessType == "" {
		return "", errors.New("AccessType cannot be empty")
	}
	if q.from == "" {
		return "", errors.New("From cannot be empty")
	}

	tokens := []string{
		"REVOKE",
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
		tokens = append(tokens, "FROM")
		tokens = append(tokens, backtick(q.from))
	}

	return strings.Join(tokens, " ") + ";", nil
}
