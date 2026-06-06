package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "C:/Users/hofma/AppData/Roaming/HuhnLite-Wails/HuhnLite.db")
	if err != nil {
		log.Fatalf("failed: %v", err)
	}
	defer db.Close()

	sqlStr := `SELECT
    V.BUCHUNGSDATUM,
    V.MENGESMALL,
    V.MENGEMEDIUM,
    V.MENGELARGE,
    V.MENGEXL,
    V.PREISSMALL,
    V.PREISMEDIUM,
    V.PREISLARGE,
    V.PREISXL,
    V.GESAMTPREIS,
    V.RABATTPROZENT
FROM VERKAUF V`

	res := translateSqlAliases(context.Background(), db, sqlStr, "en")
	fmt.Println("Resulting SQL:")
	fmt.Println(res)
}

func splitColumns(columnsPart string) []string {
	var cols []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	inBacktick := false
	parensDepth := 0
	
	for i := 0; i < len(columnsPart); i++ {
		ch := columnsPart[i]
		
		if ch == '\'' && !inDoubleQuote && !inBacktick {
			inSingleQuote = !inSingleQuote
		} else if ch == '"' && !inSingleQuote && !inBacktick {
			inDoubleQuote = !inDoubleQuote
		} else if ch == '`' && !inSingleQuote && !inDoubleQuote {
			inBacktick = !inBacktick
		}
		
		if !inSingleQuote && !inDoubleQuote && !inBacktick {
			if ch == '(' {
				parensDepth++
			} else if ch == ')' {
				parensDepth--
			}
		}
		
		if ch == ',' && parensDepth == 0 && !inSingleQuote && !inDoubleQuote && !inBacktick {
			cols = append(cols, current.String())
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		cols = append(cols, current.String())
	}
	return cols
}

func cleanQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || 
		   (s[0] == '\'' && s[len(s)-1] == '\'') || 
		   (s[0] == '`' && s[len(s)-1] == '`') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func isQuoted(s string) bool {
	if len(s) < 2 {
		return false
	}
	return (s[0] == '"' && s[len(s)-1] == '"') || 
		   (s[0] == '\'' && s[len(s)-1] == '\'') || 
		   (s[0] == '`' && s[len(s)-1] == '`')
}

func isValidSimpleIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

func parseColumnExpression(colExpr string) (string, string, bool) {
	trimmed := strings.TrimSpace(colExpr)
	inSingleQuote := false
	inDoubleQuote := false
	inBacktick := false
	parensDepth := 0
	asIndex := -1
	
	for i := len(trimmed) - 1; i >= 0; i-- {
		ch := trimmed[i]
		if ch == '\'' && !inDoubleQuote && !inBacktick {
			inSingleQuote = !inSingleQuote
		} else if ch == '"' && !inSingleQuote && !inBacktick {
			inDoubleQuote = !inDoubleQuote
		} else if ch == '`' && !inSingleQuote && !inDoubleQuote {
			inBacktick = !inBacktick
		}
		
		if !inSingleQuote && !inDoubleQuote && !inBacktick {
			if ch == ')' {
				parensDepth++
			} else if ch == '(' {
				parensDepth--
			}
			
			if parensDepth == 0 && i >= 2 {
				if (trimmed[i-1] == 'A' || trimmed[i-1] == 'a') && 
				   (trimmed[i] == 'S' || trimmed[i] == 's') {
					isWordBefore := i-2 >= 0 && (trimmed[i-2] == ' ' || trimmed[i-2] == '\t' || trimmed[i-2] == '\n' || trimmed[i-2] == '\r')
					isWordAfter := i+1 < len(trimmed) && (trimmed[i+1] == ' ' || trimmed[i+1] == '\t' || trimmed[i+1] == '\n' || trimmed[i+1] == '\r')
					if isWordBefore && isWordAfter {
						asIndex = i - 1
						break
					}
				}
			}
		}
	}
	
	if asIndex != -1 {
		baseExpr := strings.TrimSpace(trimmed[:asIndex])
		alias := strings.TrimSpace(trimmed[asIndex+2:])
		return baseExpr, cleanQuotes(alias), true
	}
	
	inSingleQuote = false
	inDoubleQuote = false
	inBacktick = false
	parensDepth = 0
	lastSpaceIndex := -1
	for i := len(trimmed) - 1; i >= 0; i-- {
		ch := trimmed[i]
		if ch == '\'' && !inDoubleQuote && !inBacktick {
			inSingleQuote = !inSingleQuote
		} else if ch == '"' && !inSingleQuote && !inBacktick {
			inDoubleQuote = !inDoubleQuote
		} else if ch == '`' && !inSingleQuote && !inDoubleQuote {
			inBacktick = !inBacktick
		}
		
		if !inSingleQuote && !inDoubleQuote && !inBacktick {
			if ch == ')' {
				parensDepth++
			} else if ch == '(' {
				parensDepth--
			}
			
			if parensDepth == 0 && (ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r') {
				lastSpaceIndex = i
				break
			}
		}
	}
	
	if lastSpaceIndex != -1 {
		baseExpr := strings.TrimSpace(trimmed[:lastSpaceIndex])
		alias := strings.TrimSpace(trimmed[lastSpaceIndex+1:])
		if isQuoted(alias) {
			return baseExpr, cleanQuotes(alias), true
		}
		if isValidSimpleIdentifier(alias) {
			return baseExpr, alias, true
		}
	}
	
	return trimmed, "", false
}

func translateSqlAliases(ctx context.Context, dbConn *sql.DB, sqlStr string, lang string) string {
	if sqlStr == "" || lang == "" {
		return sqlStr
	}

	rows, err := dbConn.QueryContext(ctx, `
		SELECT 
			fk.ID,
			fk.FELDNAME,
			t_de.BETREFF AS de_betreff,
			t_de.INHALT AS de_inhalt,
			t_curr.BETREFF AS curr_betreff
		FROM FELD_KATALOG fk
		JOIN TRANSLATEFELDNAMEN t_de ON fk.ID = t_de.ID_FELD_KATALOG AND t_de.SPRACHE_KZ = 'de'
		LEFT JOIN TRANSLATEFELDNAMEN t_curr ON fk.ID = t_curr.ID_FELD_KATALOG AND t_curr.SPRACHE_KZ = ?
	`, lang)
	if err != nil {
		log.Printf("[REPORTS] translateSqlAliases DB error: %v", err)
		return sqlStr
	}
	defer rows.Close()

	// germanLookup maps a German term (uppercased) to the ID_FELD_KATALOG (fk.ID)
	germanLookup := make(map[string]int64)
	// translationLookup maps ID_FELD_KATALOG to the translation in the target language (t_curr.BETREFF)
	translationLookup := make(map[int64]string)

	for rows.Next() {
		var id int64
		var fkFeldname, deBetreff, deInhalt, currBetreff sql.NullString
		if err := rows.Scan(&id, &fkFeldname, &deBetreff, &deInhalt, &currBetreff); err == nil {
			fName := strings.TrimSpace(fkFeldname.String)
			dBetreff := strings.TrimSpace(deBetreff.String)
			dInhalt := strings.TrimSpace(deInhalt.String)
			cBetreff := strings.TrimSpace(currBetreff.String)

			if fName != "" {
				germanLookup[strings.ToUpper(fName)] = id
			}
			if dBetreff != "" {
				germanLookup[strings.ToUpper(dBetreff)] = id
			}
			if dInhalt != "" {
				germanLookup[strings.ToUpper(dInhalt)] = id
			}

			if cBetreff != "" {
				translationLookup[id] = cBetreff
			}
		}
	}

	type SelectBlock struct {
		Start int
		End   int
	}

	var blocks []SelectBlock
	inSingleQuote := false
	inDoubleQuote := false
	inBacktick := false
	parensDepth := 0

	for i := 0; i < len(sqlStr); i++ {
		ch := sqlStr[i]
		if ch == '\'' && !inDoubleQuote && !inBacktick {
			inSingleQuote = !inSingleQuote
		} else if ch == '"' && !inSingleQuote && !inBacktick {
			inDoubleQuote = !inDoubleQuote
		} else if ch == '`' && !inSingleQuote && !inDoubleQuote {
			inBacktick = !inBacktick
		}

		if !inSingleQuote && !inDoubleQuote && !inBacktick {
			if ch == '(' {
				parensDepth++
			} else if ch == ')' {
				parensDepth--
			}

			if i >= 5 {
				if strings.EqualFold(sqlStr[i-5:i+1], "SELECT") {
					isBoundary := true
					if i-6 >= 0 {
						r := sqlStr[i-6]
						if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
							isBoundary = false
						}
					}
					if i+1 < len(sqlStr) {
						r := sqlStr[i+1]
						if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
							isBoundary = false
						}
					}
					if isBoundary {
						pDepth := 0
						inS := false
						inD := false
						inB := false
						fromIdx := -1

						for j := i + 1; j < len(sqlStr); j++ {
							ch2 := sqlStr[j]
							if ch2 == '\'' && !inD && !inB {
								inS = !inS
							} else if ch2 == '"' && !inS && !inD {
								inD = !inD
							} else if ch2 == '`' && !inS && !inD {
								inB = !inB
							}

							if !inS && !inD && !inB {
								if ch2 == '(' {
									pDepth++
								} else if ch2 == ')' {
									pDepth--
								}

								if pDepth == 0 && j >= i+4 {
									if strings.EqualFold(sqlStr[j-3:j+1], "FROM") {
										isBoundary2 := true
										if j-4 >= 0 {
											r := sqlStr[j-4]
											if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
												isBoundary2 = false
											}
										}
										if j+1 < len(sqlStr) {
											r := sqlStr[j+1]
											if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
												isBoundary2 = false
											}
										}
										if isBoundary2 {
											fromIdx = j - 3
											break
										}
									}
								}
							}
						}

						if fromIdx != -1 {
							blocks = append(blocks, SelectBlock{
								Start: i + 1,
								End:   fromIdx,
							})
						}
					}
				}
			}
		}
	}

	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Start > blocks[j].Start
	})

	extractWords := func(expr string) []string {
		cleaned := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(expr, " ")
		return strings.Fields(cleaned)
	}

	modifiedSQL := sqlStr
	for _, b := range blocks {
		columnsPart := modifiedSQL[b.Start:b.End]
		cols := splitColumns(columnsPart)
		translatedCols := make([]string, len(cols))

		for idx, col := range cols {
			baseExpr, alias, hasAlias := parseColumnExpression(col)
			var matchedID int64 = 0

			if hasAlias && alias != "" {
				aliasUpper := strings.ToUpper(strings.TrimSpace(alias))
				if id, found := germanLookup[aliasUpper]; found {
					matchedID = id
				}
			}

			if matchedID == 0 {
				baseUpper := strings.ToUpper(strings.TrimSpace(baseExpr))
				if id, found := germanLookup[baseUpper]; found {
					matchedID = id
				}
			}

			if matchedID == 0 {
				words := extractWords(baseExpr)
				for _, w := range words {
					wUpper := strings.ToUpper(w)
					if id, found := germanLookup[wUpper]; found {
						matchedID = id
						break
					}
				}
			}

			if matchedID == 0 {
				translatedCols[idx] = col
				continue
			}

			translation, found := translationLookup[matchedID]
			if !found || translation == "" {
				translatedCols[idx] = col
				continue
			}

			translatedCols[idx] = fmt.Sprintf(" %s AS %q ", strings.TrimSpace(baseExpr), translation)
		}

		newColumnsPart := strings.Join(translatedCols, ",")
		modifiedSQL = modifiedSQL[:b.Start] + newColumnsPart + modifiedSQL[b.End:]
	}

	return modifiedSQL
}
