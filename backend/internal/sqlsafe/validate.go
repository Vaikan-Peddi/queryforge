package sqlsafe

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	errEmpty          = errors.New("sql is required")
	errOnlySelect     = errors.New("only SELECT and read-only WITH queries are allowed")
	errSingleStmt     = errors.New("only one SQL statement is allowed")
	errComments       = errors.New("comments are not allowed in SQL")
	errForbiddenToken = errors.New("SQL contains a forbidden operation")

	forbiddenKeywordRE = regexp.MustCompile(`(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|ATTACH|DETACH|PRAGMA|VACUUM|REINDEX|REPLACE|MERGE|GRANT|REVOKE|CALL|EXEC|ANALYZE)\b`)
	limitRE            = regexp.MustCompile(`(?i)\bLIMIT\s+(\d+)\b`)
)

const (
	defaultLimit = 100
	maxLimit     = 500
)

func ValidateAndRewrite(sql string) (string, error) {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return "", errEmpty
	}
	if strings.Contains(trimmed, "--") || strings.Contains(trimmed, "/*") || strings.Contains(trimmed, "*/") {
		return "", errComments
	}
	trimmed = strings.TrimSuffix(trimmed, ";")
	if strings.Contains(trimmed, ";") {
		return "", errSingleStmt
	}

	lower := strings.ToLower(strings.TrimSpace(trimmed))
	if !(strings.HasPrefix(lower, "select") || strings.HasPrefix(lower, "with")) {
		return "", errOnlySelect
	}
	if forbiddenKeywordRE.MatchString(trimmed) {
		return "", errForbiddenToken
	}

	matches := limitRE.FindStringSubmatchIndex(trimmed)
	if matches == nil {
		return trimmed + fmt.Sprintf(" LIMIT %d", defaultLimit), nil
	}

	limitValue := trimmed[matches[2]:matches[3]]
	limit, err := strconv.Atoi(limitValue)
	if err != nil {
		return "", errors.New("invalid LIMIT value")
	}
	if limit <= maxLimit {
		return trimmed, nil
	}

	return trimmed[:matches[2]] + strconv.Itoa(maxLimit) + trimmed[matches[3]:], nil
}
