package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func main() {
	content, err := ioutil.ReadFile("backend/db/repo/querier.go")
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(content), "\n")
	
	var methods []string
	methodRegex := regexp.MustCompile(`^\s+([a-zA-Z0-9]+)\(ctx context\.Context,?(.*?)\) \((.*?), error\)`)
	methodRegexNoRet := regexp.MustCompile(`^\s+([a-zA-Z0-9]+)\(ctx context\.Context,?(.*?)\) error`)

	output := []string{
		"package db",
		"",
		"import (",
		"	\"context\"",
		"	\"database/sql\"",
		"	\"huhnlite-wails/backend/db/repo\"",
		"	\"huhnlite-wails/backend/db/repo_mysql\"",
		")",
		"",
		"type MariaDBQuerier struct {",
		"	sqliteQueries *repo.Queries",
		"	mysqlQueries  *repo_mysql.Queries",
		"	db            *sql.DB",
		"}",
		"",
		"func NewMariaDBQuerier(db *sql.DB) *MariaDBQuerier {",
		"	return &MariaDBQuerier{",
		"		sqliteQueries: repo.New(db),",
		"		mysqlQueries:  repo_mysql.New(db),",
		"		db:            db,",
		"	}",
		"}",
		"",
	}

	for _, line := range lines {
		m := methodRegex.FindStringSubmatch(line)
		if m != nil {
			name := m[1]
			args := m[2]
			retType := m[3]
			
			// Fix up retType to use repo. prefix if it's a model
			if !strings.Contains(retType, "[]") && !strings.Contains(retType, "interface") && !strings.Contains(retType, "Row") {
				retType = "repo." + retType
			} else if strings.Contains(retType, "[]") {
				retType = "[]repo." + strings.TrimPrefix(retType, "[]")
			} else if strings.Contains(retType, "Row") {
				retType = "repo." + retType
			}

			output = append(output, fmt.Sprintf("func (m *MariaDBQuerier) %s(ctx context.Context, %s) (%s, error) {", name, args, retType))
			
			// Logic for MySQL
			// If it's a Create/Update (originally returning model), we need to fetch it
			if strings.HasPrefix(name, "Create") || strings.HasPrefix(name, "Update") || strings.HasPrefix(name, "Add") || strings.HasPrefix(name, "Upsert") {
				// Most MySQL queries for these now return (sql.Result, error) because they are :execresult
				output = append(output, fmt.Sprintf("	res, err := m.mysqlQueries.%s(ctx, %s)", name, stripTypes(args)))
				output = append(output, "	if err != nil { return "+zeroValue(retType)+", err }")
				
				// Try to get the ID and fetch the record
				// This is a simplification, but works for most Create methods
				if strings.HasPrefix(name, "Create") {
					output = append(output, "	id, _ := res.LastInsertId()")
					// This assumes a Get[Model] method exists
					getModel := strings.TrimPrefix(name, "Create")
					getModel = strings.TrimSuffix(getModel, "Checked")
					output = append(output, fmt.Sprintf("	return m.Get%s(ctx, id)", getModel))
				} else {
					// For Update, we might not have the ID handy easily in this generic way
					output = append(output, "	return "+zeroValue(retType)+", nil // Update wrapper simplified")
				}
			} else {
				// For List/Get, we still have type mismatch
				output = append(output, "	// TODO: Map types from repo_mysql to repo")
				output = append(output, "	return m.sqliteQueries."+name+"(ctx, "+stripTypes(args)+")")
			}
			output = append(output, "}")
			output = append(output, "")
			continue
		}
		
		m2 := methodRegexNoRet.FindStringSubmatch(line)
		if m2 != nil {
			name := m2[1]
			args := m2[2]
			output = append(output, fmt.Sprintf("func (m *MariaDBQuerier) %s(ctx context.Context, %s) error {", name, args))
			output = append(output, "	return m.mysqlQueries."+name+"(ctx, "+stripTypes(args)+")")
			output = append(output, "}")
			output = append(output, "")
		}
	}

	ioutil.WriteFile("backend/db/mariadb_querier.go", []byte(strings.Join(output, "\n")), 0644)
}

func stripTypes(args string) string {
	parts := strings.Split(args, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" { continue }
		ps := strings.Split(p, " ")
		result = append(result, ps[0])
	}
	return strings.Join(result, ", ")
}

func zeroValue(t string) string {
	if strings.Contains(t, "[]") { return "nil" }
	return t + "{}"
}
