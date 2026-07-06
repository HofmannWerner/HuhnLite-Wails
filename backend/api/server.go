package api

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"

	wailsdb "huhnlite-wails/backend/db"
	db "huhnlite-wails/backend/db/repo"
	appconfig "huhnlite-wails/backend/config"
)

func toInt64(i interface{}) int64 {
	if i == nil {
		return 0
	}
	switch v := i.(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case sql.NullInt64:
		if v.Valid {
			return v.Int64
		}
		return 0
	case sql.NullFloat64:
		if v.Valid {
			return int64(v.Float64)
		}
		return 0
	case []byte:
		parsed, _ := strconv.ParseInt(string(v), 10, 64)
		return parsed
	case string:
		parsed, _ := strconv.ParseInt(v, 10, 64)
		return parsed
	default:
		return 0
	}
}

func extractFloat64(i interface{}) float64 {
	val, _ := toFloat64(i)
	return val
}

func toString(i interface{}) string {
	if i == nil {
		return ""
	}
	switch v := i.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case sql.NullString:
		if v.Valid {
			return v.String
		}
		return ""
	case sql.NullInt64:
		if v.Valid {
			return fmt.Sprintf("%v", v.Int64)
		}
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

type dbExecutor interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func getSiloNrForHerd(ctx context.Context, dbExecutor dbExecutor, herdID int64) int64 {
	var idSilo int64
	err := dbExecutor.QueryRowContext(ctx, "SELECT ID_SILO FROM HERDEN WHERE ID = ?", herdID).Scan(&idSilo)
	if err != nil || idSilo <= 0 {
		return 0
	}
	var silonr int64
	err = dbExecutor.QueryRowContext(ctx, "SELECT SILONUMMER FROM SILO WHERE ID = ?", idSilo).Scan(&silonr)
	if err != nil {
		return 0
	}
	return silonr
}

func toNullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}

func toNullString(v string) sql.NullString {
	return sql.NullString{String: v, Valid: v != ""}
}

func toNullFloat64(v float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: v, Valid: true}
}

func dInt(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func dFloat(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}

func isTrue(i interface{}) bool {
	if i == nil {
		return false
	}
	switch v := i.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	case string:
		un := strings.ToLower(v)
		return un == "true" || un == "1"
	}
	return false
}

var (
	// Optimierung: Globale RegEx zum Verhindern massiver GC Overhead in Loops
	rePara              = regexp.MustCompile(`(['"]?):?%([^;%]+);(.+?)%['"]?`)
	reBacktick          = regexp.MustCompile("`([^`]+)`")
	reSection           = regexp.MustCompile(`§([^§]+)§`)
	reDetail            = regexp.MustCompile(`(?s)<!-- BEGIN Detail -->(.*?)<!-- END Detail -->`)
	reDetailPlaceholder = regexp.MustCompile(`<Detail\.([a-zA-Z0-9_]+)>`)

	// Globaler DB-Lock als Workaround für SQLite CGO Segfaults auf ARM64
	dbMutex sync.Mutex
)

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
	log.Printf("[REPORTS] translateSqlAliases called - Lang: %q, SQL: %q", lang, sqlStr)
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

			// 1. Mit den Feldnamen im SQL Statement und dem SPRACHE_KZ = 'de' die ID_FELDKATALOG ermitteln.
			
			// Try alias if present
			if hasAlias && alias != "" {
				aliasUpper := strings.ToUpper(strings.TrimSpace(alias))
				if id, found := germanLookup[aliasUpper]; found {
					matchedID = id
				}
			}

			// Try the whole baseExpr as a single string (e.g. if it matches a field name directly or a translated name)
			if matchedID == 0 {
				baseUpper := strings.ToUpper(strings.TrimSpace(baseExpr))
				if id, found := germanLookup[baseUpper]; found {
					matchedID = id
				}
			}

			// Try parts of baseExpr (words)
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

			// Wenn nicht erfolgreich - keine weiter aktion
			if matchedID == 0 {
				translatedCols[idx] = col
				continue
			}

			// 2. Wenn erfolgreich mit ID_FELDKATALOG und SPRACHE_KZ die Übersetzung suchen.
			// Wenn nicht erfolgreich - Keine Aktion
			translation, found := translationLookup[matchedID]
			if !found || translation == "" {
				translatedCols[idx] = col
				continue
			}

			// 3. Wenn erfolgreich dann
			// wenn im SQL statement ein Alias vorhanden ist, diesen mit dem Betreff austauchen.
			// Ist kein Alias vorhanden, dannden BETREFF als Alias für das Feld im Sqlstatement anhängen
			translatedCols[idx] = fmt.Sprintf(" %s AS %q ", strings.TrimSpace(baseExpr), translation)
		}

		newColumnsPart := strings.Join(translatedCols, ",")
		modifiedSQL = modifiedSQL[:b.Start] + newColumnsPart + modifiedSQL[b.End:]
	}

	log.Printf("[REPORTS] translateSqlAliases - Original SQL: %s", sqlStr)
	log.Printf("[REPORTS] translateSqlAliases - Result SQL: %s", modifiedSQL)
	return modifiedSQL
}

func populateOriginalKeys(rows []map[string]interface{}, lang string, dbConn *sql.DB) {
	if len(rows) == 0 {
		return
	}

	dbRows, err := dbConn.Query(`
		SELECT 
			UPPER(TRIM(fk.FELDNAME)) AS original_key,
			t_curr.BETREFF AS translated_key
		FROM FELD_KATALOG fk
		JOIN TRANSLATEFELDNAMEN t_curr ON fk.ID = t_curr.ID_FELD_KATALOG
		WHERE t_curr.SPRACHE_KZ = ?
	`, lang)
	if err != nil {
		log.Printf("[REPORTS] populateOriginalKeys DB error: %v", err)
		return
	}
	defer dbRows.Close()

	reverseMap := make(map[string]string)
	for dbRows.Next() {
		var orig, trans string
		if err := dbRows.Scan(&orig, &trans); err == nil && trans != "" {
			reverseMap[strings.ToUpper(strings.TrimSpace(trans))] = orig
		}
	}

	for _, row := range rows {
		updates := make(map[string]interface{})
		for k, v := range row {
			kUpper := strings.ToUpper(strings.TrimSpace(k))
			if origKey, found := reverseMap[kUpper]; found {
				updates[origKey] = v
				updates[strings.ToLower(origKey)] = v
			}
		}
		for k, v := range updates {
			row[k] = v
		}
	}
}

func translateResponseKeys(columns []string, rows []map[string]interface{}, lang string, dbConn *sql.DB) ([]string, []map[string]interface{}) {
	if len(rows) == 0 || lang == "" || strings.EqualFold(lang, "de") {
		return columns, rows
	}

	dbRows, err := dbConn.Query(`
		SELECT 
			UPPER(TRIM(fk.FELDNAME)) AS original_key,
			t_curr.BETREFF AS translated_key
		FROM FELD_KATALOG fk
		JOIN TRANSLATEFELDNAMEN t_curr ON fk.ID = t_curr.ID_FELD_KATALOG
		WHERE t_curr.SPRACHE_KZ = ?
	`, lang)
	if err != nil {
		log.Printf("[REPORTS] translateResponseKeys DB error: %v", err)
		return columns, rows
	}
	defer dbRows.Close()

	translationMap := make(map[string]string)
	for dbRows.Next() {
		var orig, trans string
		if err := dbRows.Scan(&orig, &trans); err == nil && trans != "" {
			translationMap[orig] = trans
		}
	}

	// Translate columns array
	translatedCols := make([]string, len(columns))
	for i, col := range columns {
		colUpper := strings.ToUpper(strings.TrimSpace(col))
		if trans, found := translationMap[colUpper]; found && trans != "" {
			translatedCols[i] = trans
		} else {
			translatedCols[i] = col
		}
	}

	// Translate row keys
	translatedRows := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		newRow := make(map[string]interface{})
		for k, v := range row {
			kUpper := strings.ToUpper(strings.TrimSpace(k))
			if trans, found := translationMap[kUpper]; found && trans != "" {
				newRow[trans] = v
			} else {
				newRow[k] = v
			}
			// Keep original key as uppercase for templating/fallbacks
			newRow[kUpper] = v
		}
		translatedRows[i] = newRow
	}

	return translatedCols, translatedRows
}


func generateCharge(conn *sql.DB, herdeID int64, datum string, params db.Firmenparameter, eilagerID int64) string {
	sep := "-"
	if s := toString(params.Chargetrennung); s != "" {
		sep = s
	}

	var parts []string

	// 1. Prefix
	if s := toString(params.Chargeprefixfirma); s != "" {
		parts = append(parts, s)
	}

	// 2. Herdennummer
	if params.Chargeprefixherdennummer == 1 {
		var herdenNummer int64
		_ = conn.QueryRow("SELECT HERDENNUMMER FROM HERDEN WHERE ID = ?", herdeID).Scan(&herdenNummer)
		if herdenNummer > 0 {
			parts = append(parts, fmt.Sprintf("%d", herdenNummer))
		} else {
			parts = append(parts, fmt.Sprintf("%d", herdeID))
		}
	}

	// 3. Datum
	if params.Chargedatum == 1 {
		if len(datum) >= 10 {
			// YYYY-MM-DD -> DDMMYY
			d := datum[8:10] + datum[5:7] + datum[2:4]
			parts = append(parts, d)
		}
	}

	// 4. Lagernummer
	if params.Chargelagernummer == 1 {
		var lagerNummer int64
		_ = conn.QueryRow("SELECT LAGERNUMMER FROM EILAGER WHERE ID = ?", eilagerID).Scan(&lagerNummer)
		if lagerNummer > 0 {
			parts = append(parts, fmt.Sprintf("%d", lagerNummer))
		}
	}

	if len(parts) == 0 {
		return fmt.Sprintf("%s-%d", datum, herdeID)
	}
	return strings.Join(parts, sep)
}

func syncShowTV(ctx context.Context, queries *db.Queries, conn *sql.DB) {
	rows, err := conn.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%' AND name != 'SHOWTV'")
	if err != nil {
		log.Printf("Error fetching tables for SHOWTV sync: %v", err)
		return
	}
	defer rows.Close()

	var allNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			allNames = append(allNames, name)
		}
	}

	existing, _ := queries.ListShowTV(ctx)
	existingNames := make(map[string]bool)
	for _, e := range existing {
		if s, ok := e.Tvname.(string); ok {
			existingNames[s] = true
		}
	}

	for _, name := range allNames {
		if !existingNames[name] {
			log.Printf("SHOWTV-Sync: Registriere neue Tabelle/View: %s", name)
			_, _ = queries.CreateShowTV(ctx, db.CreateShowTVParams{
				Tvname: name,
				Showit: 0,
			})
		}
	}
}

func doAutomaticEilagerBuchung(ctx context.Context, conn *sql.DB, queries db.Querier, buchungID int64) error {
	log.Printf("[DEBUG] doAutomaticEilagerBuchung started for ID=%d", buchungID)
	b, err := queries.GetBuchung(ctx, buchungID)
	if err != nil {
		log.Printf("[DEBUG] error loading buchung: %v", err)
		return err
	}

	// Nur für Normal-, Source- oder Vermittelt-Sätze
	vKz := toString(b.Vermittelt)
	log.Printf("[DEBUG] Vermittelt-Kennzeichen: %s", vKz)
	if vKz != "N" && vKz != "S" && vKz != "V" {
		log.Printf("[DEBUG] skipping automatic booking for Vermittelt=%s", vKz)
		return nil
	}

	params, err := queries.GetFirmenparameterByHerde(ctx, b.IDHerden)
	if err != nil {
		log.Printf("[DEBUG] error loading params: %v", err)
		return err
	}

	log.Printf("[DEBUG] LAGERBUCHUNGBEIBUCHUNG value: %d", params.Lagerbuchungbeibuchung)
	if params.Lagerbuchungbeibuchung == 0 {
		return nil
	}

	h, err := queries.GetHerde(ctx, b.IDHerden)
	if err != nil {
		log.Printf("[DEBUG] error loading herde: %v", err)
		return err
	}
	if h.IDEilager <= 0 {
		log.Printf("[DEBUG] no ID_EILAGER assigned to herde")
		return nil
	}

	// Charge generieren
	charge := generateCharge(conn, b.IDHerden, b.Buchungsdatum, params, h.IDEilager)

	// Lagerplatz (Fremdeslager) ermitteln (erster Treffer)
	var idFremdesLager int64
	_ = conn.QueryRowContext(ctx, "SELECT ID FROM LAGERPLATZ WHERE ID_EILAGER = ? LIMIT 1", h.IDEilager).Scan(&idFremdesLager)

	// Bestehende automatische Buchungen für diese Buchung-ID löschen
	_, _ = conn.ExecContext(ctx, "DELETE FROM EILAGERBUCHUNG WHERE ID_BUCHUNG = ?", buchungID)

	// Sicherheits-Bereinigung: Verwaiste Sätze für dieses Lager am selben Tag löschen
	// (Verhindert doppelte Bestände, wenn alte Buchungen unsauber gelöscht wurden)
	_, _ = conn.ExecContext(ctx, `
		DELETE FROM EILAGERBUCHUNG 
		WHERE ID_EILAGER = ? AND BUCHUNGSDATUM = ? 
		AND (ID_BUCHUNG = 0 OR ID_BUCHUNG NOT IN (SELECT ID FROM BUCHUNG))`,
		h.IDEilager, b.Buchungsdatum)

	log.Printf("[DEBUG] Creating automatic Eilagerbuchung for Datum=%s, Medium=%d, ID_BUCHUNG=%d",
		b.Buchungsdatum, b.Medium, buchungID)

	// Neu anlegen
	res, err := queries.AddEilagerBuchung(ctx, db.AddEilagerBuchungParams{
		IDBuchung:      buchungID,
		IDEilager:      h.IDEilager,
		IDFremdeslager: idFremdesLager,
		Buchungsdatum:  b.Buchungsdatum,
		Jumbos:         b.Kl6,
		Xl:             b.Xl,
		Large:          b.Large,
		Medium:         b.Medium,
		Small:          b.Small,
		Volleikg:       b.Vollei,
		Schmutz:        b.Schmutz,
		Knickeier:      b.Knickeier,
		Brucheier:      b.Brucheier,
		Charge:         charge,
		Buchungstyp:    "E",
		KzLager:        "E",
	})
	if err != nil {
		log.Printf("[DEBUG] Error creating automatic Eilagerbuchung: %v", err)
	} else {
		log.Printf("[DEBUG] Created automatic Eilagerbuchung ID=%d with Medium=%d", res.ID, res.Medium)
	}

	return err
}

func getHelpDir(database *wailsdb.DB) string {
	langs := []string{"de", "en", "it"}
	var pathsToCheck []string

	for _, lang := range langs {
		fileName := fmt.Sprintf("HuhnLite-%s.html", lang)

		// 1. If SQLite, check its directory
		if database != nil && database.Engine == "sqlite" && database.Config.DBConnectionString != "" {
			dbDir := filepath.Dir(database.Config.DBConnectionString)
			pathsToCheck = append(pathsToCheck, filepath.Join(dbDir, fileName))
		}

		// 2. Check executable directory
		if execPath, err := os.Executable(); err == nil {
			bundleDir := filepath.Dir(execPath)
			if filepath.Base(bundleDir) == "MacOS" && filepath.Base(filepath.Dir(bundleDir)) == "Contents" {
				// Check Contents/Resources inside the .app bundle
				resourcesDir := filepath.Join(filepath.Dir(bundleDir), "Resources")
				pathsToCheck = append(pathsToCheck, filepath.Join(resourcesDir, fileName))

				bundleDir = filepath.Dir(filepath.Dir(filepath.Dir(bundleDir)))
			}
			pathsToCheck = append(pathsToCheck, filepath.Join(bundleDir, fileName))
		}

		// 3. Check CWD
		if cwd, err := os.Getwd(); err == nil {
			pathsToCheck = append(pathsToCheck, filepath.Join(cwd, fileName))
		}

		// 4. Check AppDataDir
		if configDir, err := os.UserConfigDir(); err == nil {
			pathsToCheck = append(pathsToCheck, filepath.Join(configDir, "HuhnLite-Wails", fileName))
			pathsToCheck = append(pathsToCheck, filepath.Join(configDir, fileName))
		}

		// 5. Check LOCALAPPDATA
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			pathsToCheck = append(pathsToCheck, filepath.Join(localAppData, "HuhnLite-Wails", fileName))
			pathsToCheck = append(pathsToCheck, filepath.Join(localAppData, fileName))
		}

		// 6. Check APPDATA
		if appData := os.Getenv("APPDATA"); appData != "" {
			pathsToCheck = append(pathsToCheck, filepath.Join(appData, "HuhnLite-Wails", fileName))
			pathsToCheck = append(pathsToCheck, filepath.Join(appData, fileName))
		}
	}

	for _, p := range pathsToCheck {
		if _, err := os.Stat(p); err == nil {
			return filepath.Dir(p)
		}
	}
	return ""
}

func StartServer(database *wailsdb.DB) *gin.Engine {
	log.Printf("[DEBUG] StartServer called with engine: %s", database.Engine)
	conn := database.SQL
	queries := database.Repo
	currentDBPath := database.Config.DBConnectionString
	backupDir := filepath.Join(filepath.Dir(currentDBPath), "backups")
	_ = os.MkdirAll(backupDir, 0755)
	migrateDB(database)

	// Gin Engine initialisieren
	r := gin.New()
	r.Use(gin.Recovery())
	// Default to release mode unless configured otherwise
	gin.SetMode(gin.ReleaseMode)
	if os.Getenv("APP_ENV") == "dev" {
		r.Use(gin.Logger())
	}

	// Serve help directory statically if found
	if helpDir := getHelpDir(database); helpDir != "" {
		log.Printf("[API] Serving help files from: %s", helpDir)
		r.Static("/help", helpDir)
	} else {
		log.Println("[API] Help directory not found, static /help route will not be served")
	}

	// Verhindert SQLite CGO Segfaults bei abgebrochenen HTTP-Requests (Axios unmount cancellation)
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(context.WithoutCancel(c.Request.Context()))
		c.Next()
	})

	// Robustere Recovery (fängt Panics ab, bevor der Server stirbt)
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("[PANIC-RECOVERY] Recovered from: %v", recovered)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "Interner Server-Fehler (Panic)",
			"details": fmt.Sprintf("%v", recovered),
		})
	}))

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	r.Use(cors.New(config))

	// Disable caching to prevent frontend stale data issues
	r.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	})

	// System-Konfiguration
	r.GET("/api/config", func(c *gin.Context) {
		// Default from .env
		authEnabled := os.Getenv("APP_AUTH_ENABLED") == "true"

		// Database override
		var dbVal string
		err := conn.QueryRowContext(c, "SELECT VALUE FROM SYSTEMSETTINGS WHERE NAME = 'auth_required'").Scan(&dbVal)
		if err == nil {
			authEnabled = (dbVal == "true" || dbVal == "1")
			log.Printf("[API] /api/config - auth_required from DB: %s -> authEnabled: %v", dbVal, authEnabled)
		} else {
			log.Printf("[API] /api/config - auth_required not found in DB, using default: %v", authEnabled)
		}

		cfg := appconfig.LoadConfig()
		database.Config = cfg // Update in-memory config

		c.JSON(http.StatusOK, gin.H{
			"auth_enabled": authEnabled,
			"system_edit_enabled": cfg.System == 1,
		})
	})

	r.POST("/api/system/shutdown", func(c *gin.Context) {
		log.Println("[API] Shutdown-Anfrage erhalten")
		c.JSON(http.StatusOK, gin.H{"message": "Server-Shutdown wird vorbereitet..."})
	})

	// System-Settings APIs
	r.GET("/api/system-settings/:name", func(c *gin.Context) {
		name := c.Param("name")
		var value string
		err := conn.QueryRowContext(c, "SELECT VALUE FROM SYSTEMSETTINGS WHERE NAME = ?", name).Scan(&value)
		if err != nil {
			log.Printf("[API] GET /api/system-settings/%s - Error: %v", name, err)
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var jsonVal interface{} = value
		if value == "true" || value == "1" {
			jsonVal = true
		} else if value == "false" || value == "0" {
			jsonVal = false
		}

		c.JSON(http.StatusOK, gin.H{"name": name, "value": jsonVal})
	})

	r.POST("/api/system-settings", func(c *gin.Context) {
		var req struct {
			Name  string `json:"name" binding:"required"`
			Value string `json:"value" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		upsertQuery := "INSERT INTO SYSTEMSETTINGS (NAME, VALUE) VALUES (?, ?) ON CONFLICT(NAME) DO UPDATE SET VALUE = excluded.VALUE"
		if database.Engine == "mysql" {
			upsertQuery = "INSERT INTO SYSTEMSETTINGS (NAME, VALUE) VALUES (?, ?) ON DUPLICATE KEY UPDATE VALUE = VALUES(VALUE)"
		}
		_, err := conn.ExecContext(c, upsertQuery, req.Name, req.Value)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Firmenparameter APIs

	r.GET("/api/firmenparameter/:typ/:id", func(c *gin.Context) {
		kz := c.Param("typ")
		idStr := c.Param("id")
		idInt, _ := strconv.ParseInt(idStr, 10, 64)

		if kz == "F" {
			idInt = -1
		}

		res, err := queries.GetFirmenparameterByHerde(c, idInt)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusOK, gin.H{})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/firmenparameter-herden-ids", func(c *gin.Context) {
		rows, err := conn.QueryContext(c, "SELECT DISTINCT ID_HERDEN FROM FIRMENPARAMETER WHERE ID_HERDEN > 0")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var ids []int64 = []int64{}
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err == nil {
				ids = append(ids, id)
			}
		}
		c.JSON(http.StatusOK, ids)
	})

	r.GET("/api/firmenparameter/get-or-create/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		herdeID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		// 1. Check if parameters exist for this Herde
		currentParam, err := queries.GetFirmenparameterByHerde(c, herdeID)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"data": currentParam, "is_new": false})
			return
		} else if err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 2. If it does not exist, get Firma parameter (ID_HERDEN = -1)
		firmaParam, err := queries.GetFirmenparameterByHerde(c, -1)
		if err != nil && err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err == sql.ErrNoRows {
			// Fallback: Default values if even -1 is missing
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"id_herden": herdeID,
					"kz":        "H",
				},
				"is_new": true,
			})
			return
		}

		// Return Firma parameters as a template for the new herd parameters
		// BUT set id_herden to the requested herd ID so the frontend knows what to save to
		firmaParam.IDHerden = herdeID
		firmaParam.Kz = "H"

		c.JSON(http.StatusOK, gin.H{"data": firmaParam, "is_new": true})
	})

	r.GET("/api/firmenparameter/fieldcontrol/:id_herden", func(c *gin.Context) {
		idStr := c.Param("id_herden")
		herdeID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		// 1. Try herd specific
		res, err := queries.GetFirmenparameterByHerde(c, herdeID)
		if err == nil {
			c.JSON(http.StatusOK, res)
			return
		}

		// 2. Fallback to global (-1)
		res, err = queries.GetFirmenparameterByHerde(c, -1)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "default parameters not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	type ParamInput struct {
		Kz                        string   `json:"KZ"`
		Jumbos                    int64    `json:"JUMBOS"`
		Klassenerfassen           int64    `json:"KLASSENERFASSEN"`
		Klasseaerfassen           int64    `json:"KLASSEAERFASSEN"`
		Klasseaerrechnen          int64    `json:"KLASSEAERRECHNEN"`
		Klasseavermitteln         int64    `json:"KLASSEAVERMITTELN"`
		Erfasseschmutzei          int64    `json:"ERFASSESCHMUTZEI"`
		Erfasseknickei            int64    `json:"ERFASSEKNICKEI"`
		Erfassebruchei            int64    `json:"ERFASSEBRUCHEI"`
		Erfassevollei             int64    `json:"ERFASSEVOLLEI"`
		Massvollei                *int64   `json:"MASSVOLLEI"`
		Aufteilunggewicht         int64    `json:"AUFTEILUNGGEWICHT"`
		Kontrollwiegung           int64    `json:"KONTROLLWIEGUNG"`
		Anzahlkontrollw           *int64   `json:"ANZAHLKONTROLLW"`
		Verpackungkg              *float64 `json:"VERPACKUNGKG"`
		Aufteilungalter           int64    `json:"AUFTEILUNGALTER"`
		Erfassevolleikg           int64    `json:"ERFASSEVOLLEIKG"`
		Laufzeitwochen            *int64   `json:"LAUFZEITWOCHEN"`
		Zeitstempel               string   `json:"ZEITSTEMPEL"`
		Schlachterloeshenne       *float64 `json:"SCHLACHTERLOESHENNE"`
		Produktionsdauer          *int64   `json:"PRODUKTIONSDAUER"`
		IDTabellegewicht          *int64   `json:"ID_TABELLEGEWICHT"`
		IDTabellealter            *int64   `json:"ID_TABELLEALTER"`
		LegebeginnLw              *int64   `json:"LEGEBEGINN_LW"`
		Verlustebeibuchung        int64    `json:"VERLUSTEBEIBUCHUNG"`
		Lagerbuchungbeibuchung    int64    `json:"LAGERBUCHUNGBEIBUCHUNG"`
		Maxtagevermitteln         *int64   `json:"MAXTAGEVERMITTELN"`
		Chargejumbos              int64    `json:"CHARGEJUMBOS"`
		Chargexl                  int64    `json:"CHARGEXL"`
		Chargemedium              int64    `json:"CHARGEMEDIUM"`
		ChargeSmall               int64    `json:"CHARGESMALL"`
		Chargelarge               int64    `json:"CHARGELARGE"`
		Chargevollei              int64    `json:"CHARGEVOLLEI"`
		Chargeprefixfirma         string   `json:"CHARGEPREFIXFIRMA"`
		Chargeprefixherdennummer  int64    `json:"CHARGEPREFIXHERDENNUMMER"`
		Chargedatum               int64    `json:"CHARGEDATUM"`
		Chargelagernummer         int64    `json:"CHARGELAGERNUMMER"`
		Chargetrennung            string   `json:"CHARGETRENNUNG"`
		Beivermittelndatumaktuell int64    `json:"BEIVERMITTELNDATUMAKTUELL"`
		Pseudolager               int64    `json:"PSEUDOLAGER"`
		Bio                       int64    `json:"BIO"`
		Haltungstyp               string   `json:"HALTUNGSTYP"`
		Bioaufschlag              float64  `json:"BIOAUFSCHLAG"`
		Futterinventur            int64    `json:"FUTTERINVENTUR"`
	}

	r.POST("/api/firmenparameter", func(c *gin.Context) {
		var req struct {
			IDHerden int64 `json:"ID_HERDEN"`
			ParamInput
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[API] POST /api/firmenparameter - Bind Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf("[API] POST /api/firmenparameter - Data: %+v", req)

		params := db.CreateFirmenparameterParams{
			IDHerden:                  req.IDHerden,
			Kz:                        req.Kz,
			Jumbos:                    req.Jumbos,
			Klassenerfassen:           req.Klassenerfassen,
			Klasseaerfassen:           req.Klasseaerfassen,
			Klasseaerrechnen:          req.Klasseaerrechnen,
			Klasseavermitteln:         req.Klasseavermitteln,
			Erfasseschmutzei:          req.Erfasseschmutzei,
			Erfasseknickei:            req.Erfasseknickei,
			Erfassebruchei:            req.Erfassebruchei,
			Erfassevollei:             req.Erfassevollei,
			Massvollei:                dInt(req.Massvollei),
			Aufteilunggewicht:         req.Aufteilunggewicht,
			Kontrollwiegung:           req.Kontrollwiegung,
			Anzahlkontrollw:           dInt(req.Anzahlkontrollw),
			Verpackungkg:              dFloat(req.Verpackungkg),
			Aufteilungalter:           req.Aufteilungalter,
			Erfassevolleikg:           req.Erfassevolleikg,
			Laufzeitwochen:            dInt(req.Laufzeitwochen),
			Zeitstempel:               req.Zeitstempel,
			Schlachterloeshenne:       dFloat(req.Schlachterloeshenne),
			Produktionsdauer:          dInt(req.Produktionsdauer),
			IDTabellegewicht:          dInt(req.IDTabellegewicht),
			IDTabellealter:            dInt(req.IDTabellealter),
			LegebeginnLw:              dInt(req.LegebeginnLw),
			Verlustebeibuchung:        req.Verlustebeibuchung,
			Lagerbuchungbeibuchung:    req.Lagerbuchungbeibuchung,
			Maxtagevermitteln:         dInt(req.Maxtagevermitteln),
			Chargejumbos:              req.Chargejumbos,
			Chargexl:                  req.Chargexl,
			Chargemedium:              req.Chargemedium,
			Chargesmall:               req.ChargeSmall,
			Chargelarge:               req.Chargelarge,
			Chargevollei:              req.Chargevollei,
			Chargeprefixfirma:         req.Chargeprefixfirma,
			Chargeprefixherdennummer:  req.Chargeprefixherdennummer,
			Chargedatum:               req.Chargedatum,
			Chargelagernummer:         req.Chargelagernummer,
			Chargetrennung:            req.Chargetrennung,
			Beivermittelndatumaktuell: req.Beivermittelndatumaktuell,
			Pseudolager:               req.Pseudolager,
			Bio:                       req.Bio,
			Haltungstyp:               req.Haltungstyp,
			Bioaufschlag:              req.Bioaufschlag,
			Futterinventur:            req.Futterinventur,
		}

		// FOOLPROOF STRATEGY: Delete any existing record for this herd, then insert the new one.
		_ = queries.DeleteFirmenparameter(c, req.IDHerden)

		res, err := queries.CreateFirmenparameter(c, params)
		if err != nil {
			log.Printf("[API] POST Firmenparameter Save Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/firmenparameter/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idHerden, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req ParamInput
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[API] PUT /api/firmenparameter/%d - Bind Error: %v", idHerden, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf("[API] PUT /api/firmenparameter/%d - Data: %+v", idHerden, req)

		// FOOLPROOF STRATEGY: Delete any existing record for this herd, then insert the new one.
		_ = queries.DeleteFirmenparameter(c, idHerden)

		params := db.CreateFirmenparameterParams{
			IDHerden:                  idHerden,
			Kz:                        req.Kz,
			Jumbos:                    req.Jumbos,
			Klassenerfassen:           req.Klassenerfassen,
			Klasseaerfassen:           req.Klasseaerfassen,
			Klasseaerrechnen:          req.Klasseaerrechnen,
			Klasseavermitteln:         req.Klasseavermitteln,
			Erfasseschmutzei:          req.Erfasseschmutzei,
			Erfasseknickei:            req.Erfasseknickei,
			Erfassebruchei:            req.Erfassebruchei,
			Erfassevollei:             req.Erfassevollei,
			Massvollei:                dInt(req.Massvollei),
			Aufteilunggewicht:         req.Aufteilunggewicht,
			Kontrollwiegung:           req.Kontrollwiegung,
			Anzahlkontrollw:           dInt(req.Anzahlkontrollw),
			Verpackungkg:              dFloat(req.Verpackungkg),
			Aufteilungalter:           req.Aufteilungalter,
			Erfassevolleikg:           req.Erfassevolleikg,
			Laufzeitwochen:            dInt(req.Laufzeitwochen),
			Zeitstempel:               req.Zeitstempel,
			Schlachterloeshenne:       dFloat(req.Schlachterloeshenne),
			Produktionsdauer:          dInt(req.Produktionsdauer),
			IDTabellegewicht:          dInt(req.IDTabellegewicht),
			IDTabellealter:            dInt(req.IDTabellealter),
			LegebeginnLw:              dInt(req.LegebeginnLw),
			Verlustebeibuchung:        req.Verlustebeibuchung,
			Lagerbuchungbeibuchung:    req.Lagerbuchungbeibuchung,
			Maxtagevermitteln:         dInt(req.Maxtagevermitteln),
			Chargejumbos:              req.Chargejumbos,
			Chargexl:                  req.Chargexl,
			Chargemedium:              req.Chargemedium,
			Chargesmall:               req.ChargeSmall,
			Chargelarge:               req.Chargelarge,
			Chargevollei:              req.Chargevollei,
			Chargeprefixfirma:         req.Chargeprefixfirma,
			Chargeprefixherdennummer:  req.Chargeprefixherdennummer,
			Chargedatum:               req.Chargedatum,
			Chargelagernummer:         req.Chargelagernummer,
			Chargetrennung:            req.Chargetrennung,
			Beivermittelndatumaktuell: req.Beivermittelndatumaktuell,
			Pseudolager:               req.Pseudolager,
			Bio:                       req.Bio,
			Haltungstyp:               req.Haltungstyp,
			Bioaufschlag:              req.Bioaufschlag,
			Futterinventur:            req.Futterinventur,
		}

		_, dbErr := queries.CreateFirmenparameter(c, params)
		if dbErr != nil {
			log.Printf("[API] Firmenparameter Save Error: %v", dbErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": dbErr.Error()})
			return
		}

		log.Printf("DEBUG: Result for ID_HERDEN: %d, Pseudolager: %d (1=ON, 0=OFF)\n", idHerden, req.Pseudolager)

		// Now fetch the record to return it - this is MUCH SAFER than RETURNING *
		res, err := queries.GetFirmenparameterByHerde(c, idHerden)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch record after save: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/firmenparameter/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idHerden, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		log.Printf("DEBUG: Request to DELETE FIRMENPARAMETER for ID_HERDEN: %d\n", idHerden)

		// Get the underlying db.Exec result to check RowsAffected
		arg := idHerden
		res, err := conn.ExecContext(c, "DELETE FROM FIRMENPARAMETER WHERE ID_HERDEN = ?", arg)
		if err != nil {
			log.Printf("DEBUG: DELETE failed for ID %d: %v\n", idHerden, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		rows, _ := res.RowsAffected()
		log.Printf("DEBUG: DELETE successful for ID_HERDEN: %d, Rows affected: %d\n", idHerden, rows)

		c.JSON(http.StatusOK, gin.H{"status": "deleted", "rows_affected": rows})
	})

	// SHOWTV APIs
	r.GET("/api/showtv", func(c *gin.Context) {
		res, err := queries.ListShowTV(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/showtv/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req struct {
			Tvname string `json:"TVNAME"`
			Showit int64  `json:"SHOWIT"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateShowTVParams{
			ID:     idInt,
			Tvname: req.Tvname,
			Showit: req.Showit,
		}
		res, err := queries.UpdateShowTV(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	// Rasse APIs

	r.GET("/api/rasse", func(c *gin.Context) {
		res, err := queries.ListRassen(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/rasse", func(c *gin.Context) {
		var req struct {
			Rasse string `json:"rasse" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := queries.CreateRasse(c, req.Rasse)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/rasse/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req struct {
			Rasse string `json:"rasse" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateRasseParams{
			ID:    idInt,
			Rasse: req.Rasse,
		}
		res, err := queries.UpdateRasse(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/rasse/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		// Check if any herds are using this rasse
		var count int64
		err = conn.QueryRowContext(c, "SELECT COUNT(*) FROM HERDEN WHERE ID_RASSE = ?", idInt).Scan(&count)
		if err == nil && count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Rasse kann nicht gelöscht werden, da sie noch Herden zugeordnet ist."})
			return
		}

		if err := queries.DeleteRasse(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Mwst APIs
	r.GET("/api/mwst", func(c *gin.Context) {
		res, err := queries.ListMwst(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/mwst", func(c *gin.Context) {
		var req struct {
			Mwstkz  string  `json:"mwstkz" binding:"required"`
			Prozent float64 `json:"PROZENT"`
			Konto   string  `json:"KONTO"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.CreateMwstParams{
			Mwstkz:  req.Mwstkz,
			Prozent: req.Prozent,
			Konto:   req.Konto,
		}
		res, err := queries.CreateMwst(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/mwst/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req struct {
			Mwstkz  string  `json:"mwstkz" binding:"required"`
			Prozent float64 `json:"PROZENT"`
			Konto   string  `json:"KONTO"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateMwstParams{
			ID:      idInt,
			Mwstkz:  req.Mwstkz,
			Prozent: req.Prozent,
			Konto:   req.Konto,
		}
		res, err := queries.UpdateMwst(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/mwst/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		if err := queries.DeleteMwst(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Futtersorten APIs
	r.GET("/api/eierpreise", func(c *gin.Context) {
		res, err := queries.ListEierpreise(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/futtersorten", func(c *gin.Context) {
		res, err := queries.ListFuttersorten(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/futtersorten", func(c *gin.Context) {
		var req struct {
			Bezeichnung string `json:"bezeichnung" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := queries.CreateFuttersorte(c, req.Bezeichnung)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/futtersorten/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req struct {
			Bezeichnung string `json:"bezeichnung" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateFuttersorteParams{
			ID:          idInt,
			Bezeichnung: req.Bezeichnung,
		}
		res, err := queries.UpdateFuttersorte(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/futtersorten/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		if err := queries.DeleteFuttersorte(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Silo APIs
	r.GET("/api/silo", func(c *gin.Context) {
		res, err := queries.ListSilos(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/silo", func(c *gin.Context) {
		var req struct {
			Silonummer         int64  `json:"silonummer" binding:"required"`
			Bezeichnung        string `json:"bezeichnung" binding:"required"`
			Inventurdatumalt   string `json:"INVENTURDATUMALT"`
			Inventurdatumneu   string `json:"INVENTURDATUMNEU"`
			Maxfuellmenge      int64  `json:"MAXFUELLMENGE"`
			Minfuellmenge      int64  `json:"MINFUELLMENGE"`
			Inventurfuellmenge int64  `json:"INVENTURFUELLMENGE"`
			IDLieferant        int64  `json:"ID_LIEFERANT"`
			Personennummer     int64  `json:"PERSONENNUMMER"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.CreateSiloParams{
			Silonummer:         req.Silonummer,
			Personennummer:     req.Personennummer,
			Bezeichnung:        req.Bezeichnung,
			Inventurdatumalt:   req.Inventurdatumalt,
			Inventurdatumneu:   req.Inventurdatumneu,
			Maxfuellmenge:      req.Maxfuellmenge,
			Minfuellmenge:      req.Minfuellmenge,
			Inventurfuellmenge: req.Inventurfuellmenge,
			IDLieferant:        req.IDLieferant,
		}
		res, err := queries.CreateSilo(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/silo/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}

		var req struct {
			Silonummer         int64  `json:"silonummer" binding:"required"`
			Bezeichnung        string `json:"bezeichnung" binding:"required"`
			Inventurdatumalt   string `json:"INVENTURDATUMALT"`
			Inventurdatumneu   string `json:"INVENTURDATUMNEU"`
			Maxfuellmenge      int64  `json:"MAXFUELLMENGE"`
			Minfuellmenge      int64  `json:"MINFUELLMENGE"`
			Inventurfuellmenge int64  `json:"INVENTURFUELLMENGE"`
			IDLieferant        int64  `json:"ID_LIEFERANT"`
			Personennummer     int64  `json:"PERSONENNUMMER"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateSiloParams{
			ID:                 idInt,
			Silonummer:         req.Silonummer,
			Personennummer:     req.Personennummer,
			Bezeichnung:        req.Bezeichnung,
			Inventurdatumalt:   req.Inventurdatumalt,
			Inventurdatumneu:   req.Inventurdatumneu,
			Maxfuellmenge:      req.Maxfuellmenge,
			Minfuellmenge:      req.Minfuellmenge,
			Inventurfuellmenge: req.Inventurfuellmenge,
			IDLieferant:        req.IDLieferant,
		}
		res, err := queries.UpdateSilo(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/silo/:id/inventur", func(c *gin.Context) {
		idStr := c.Param("id")
		siloID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("[INVENTUR] Invalid Silo ID: %s, error: %v", idStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiges Silo-ID Format"})
			return
		}

		var req struct {
			Inventurdatum      string  `json:"INVENTURDATUMNEU" binding:"required"`
			Inventurfuellmenge float64 `json:"INVENTURFUELLMENGE" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[INVENTUR] JSON bind error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("[INVENTUR] Start processing SiloID: %d, Date: %s, Qty: %.2f", siloID, req.Inventurdatum, req.Inventurfuellmenge)

		// 1. Transaction starten
		tx, err := conn.BeginTx(c, nil)
		if err != nil {
			log.Printf("[INVENTUR] Failed to begin transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Starten der Transaktion: " + err.Error()})
			return
		}
		defer tx.Rollback()

		// 2. Silo laden
		var silonummer int64
		var bezeichnung string
		var inventurDatumAlt string
		var inventurDatumNeu string
		err = tx.QueryRowContext(c, "SELECT SILONUMMER, BEZEICHNUNG, INVENTURDATUMALT, INVENTURDATUMNEU FROM SILO WHERE ID = ?", siloID).Scan(&silonummer, &bezeichnung, &inventurDatumAlt, &inventurDatumNeu)
		if err != nil {
			log.Printf("[INVENTUR] Failed to load Silo ID %d: %v", siloID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Silo nicht gefunden: " + err.Error()})
			return
		}
		log.Printf("[INVENTUR] Loaded Silo Nr: %d, Bez: %s, AltDate: %s, NeuDate: %s", silonummer, bezeichnung, inventurDatumAlt, inventurDatumNeu)

		// Fallback für leeres Alt-Datum
		if strings.TrimSpace(inventurDatumAlt) == "" {
			inventurDatumAlt = "0001-01-01"
		}

		// 3. Check 1: Max(Buchungsdatum) der angeschlossenen Herden > Inventurdatum (nur aktive Herden berücksichtigen)
		var maxBuchungsdatum sql.NullString
		err = tx.QueryRowContext(c, `
			SELECT MAX(B.BUCHUNGSDATUM) 
			FROM BUCHUNG B
			JOIN HERDEN H ON B.ID_HERDEN = H.ID
			WHERE H.ID_SILO = ? AND H.AKTIV = 1`, siloID).Scan(&maxBuchungsdatum)
		if err != nil {
			log.Printf("[INVENTUR] Failed to fetch max booking date for SiloNr %d: %v", silonummer, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abfragen des maximalen Buchungsdatums: " + err.Error()})
			return
		}

		maxDateStr := "0001-01-01"
		if maxBuchungsdatum.Valid {
			maxDateStr = maxBuchungsdatum.String
		}
		log.Printf("[INVENTUR] Max booking date for SiloNr %d is %s", silonummer, maxDateStr)

		if maxDateStr <= req.Inventurdatum {
			log.Printf("[INVENTUR] Validation failed: Max booking date %s <= entered date %s", maxDateStr, req.Inventurdatum)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Das maximale Buchungsdatum der angeschlossenen Herden (%s) muss nach dem Inventurdatum (%s) liegen.", maxDateStr, req.Inventurdatum)})
			return
		}

		// 4. Letzte Futterlieferung holen
		var lastF struct {
			ID_SILO         int64
			SILONUMMER      int64
			HERDENR         int64
			ID_PERSON       int64
			LIEFERDATUM     string
			LIEFERMENGE     float64
			PREISDT         float64
			RABATTPROZ      float64
			NETTO           float64
			BRUTTO          float64
			MWSTPROZ        float64
			MWSTKZ          interface{}
			DATUM           string
			ZEITSTEMPEL     string
			ID_FUTTERSORTEN int64
			AW              int64
		}
		err = tx.QueryRowContext(c, `
			SELECT ID_SILO, SILONUMMER, HERDENR, ID_PERSON, LIEFERDATUM, LIEFERMENGE, PREISDT, RABATTPROZ, NETTO, BRUTTO, MWSTPROZ, MWSTKZ, DATUM, ZEITSTEMPEL, ID_FUTTERSORTEN, AW
			FROM FUTTER WHERE ID_SILO = ? ORDER BY LIEFERDATUM DESC, ID DESC LIMIT 1`, siloID).
			Scan(&lastF.ID_SILO, &lastF.SILONUMMER, &lastF.HERDENR, &lastF.ID_PERSON, &lastF.LIEFERDATUM, &lastF.LIEFERMENGE, &lastF.PREISDT, &lastF.RABATTPROZ, &lastF.NETTO, &lastF.BRUTTO, &lastF.MWSTPROZ, &lastF.MWSTKZ, &lastF.DATUM, &lastF.ZEITSTEMPEL, &lastF.ID_FUTTERSORTEN, &lastF.AW)
		
		if err != nil {
			log.Printf("[INVENTUR] Failed to load last feed delivery: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Keine vorherige Futterlieferung für dieses Silo gefunden. Kopieren nicht möglich."})
			return
		}
		log.Printf("[INVENTUR] Copying last feed delivery: ID_PERSON=%d, SortenID=%d, PreisDt=%.2f, MwStProz=%.2f, MwStKz=%v",
			lastF.ID_PERSON, lastF.ID_FUTTERSORTEN, lastF.PREISDT, lastF.MWSTPROZ, lastF.MWSTKZ)

		// 5. Zwei Buchungssätze erstellen
		// a) Liefermenge auf Inventurmenge, Lieferdatum auf Inventurdatum
		inventurMenge := req.Inventurfuellmenge
		inventurDatum := req.Inventurdatum
		mwstKzStr := toString(lastF.MWSTKZ)

		// Netto und Brutto berechnen
		nettoA := inventurMenge * (lastF.PREISDT / 100.0) * (1.0 - lastF.RABATTPROZ / 100.0)
		bruttoA := nettoA * (1.0 + lastF.MWSTPROZ / 100.0)
		log.Printf("[INVENTUR] Calculated Booking A: Net=%.2f, Gross=%.2f", nettoA, bruttoA)

		insertQuery := `
			INSERT INTO FUTTER (ID_SILO, SILONUMMER, HERDENR, ID_PERSON, LIEFERDATUM, LIEFERMENGE, PREISDT, RABATTPROZ, NETTO, BRUTTO, MWSTPROZ, MWSTKZ, DATUM, ZEITSTEMPEL, AW, ID_FUTTERSORTEN)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		resA, err := tx.ExecContext(c, insertQuery,
			lastF.ID_SILO, lastF.SILONUMMER, lastF.HERDENR, lastF.ID_PERSON,
			inventurDatum, inventurMenge, lastF.PREISDT, lastF.RABATTPROZ,
			nettoA, bruttoA, lastF.MWSTPROZ, mwstKzStr,
			inventurDatum, inventurDatum + " 12:00:00Z", lastF.AW, lastF.ID_FUTTERSORTEN,
		)
		if err != nil {
			log.Printf("[INVENTUR] Failed to insert Booking A: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Speichern der ersten Futterbuchung (a): " + err.Error()})
			return
		}
		idA, _ := resA.LastInsertId()
		log.Printf("[INVENTUR] Booking A inserted successfully. ID=%d", idA)

		// b) Lieferdatum auf Inventurdatum - 1, Liefermenge auf minus (negative)
		tParsed, err := time.Parse("2006-01-02", inventurDatum)
		if err != nil {
			log.Printf("[INVENTUR] Date parse error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiges Datumsformat: " + err.Error()})
			return
		}
		prevDatum := tParsed.AddDate(0, 0, -1).Format("2006-01-02")
		mengeB := -inventurMenge
		nettoB := mengeB * (lastF.PREISDT / 100.0) * (1.0 - lastF.RABATTPROZ / 100.0)
		bruttoB := nettoB * (1.0 + lastF.MWSTPROZ / 100.0)
		log.Printf("[INVENTUR] Calculated Booking B: Date=%s, Qty=%.2f, Net=%.2f, Gross=%.2f", prevDatum, mengeB, nettoB, bruttoB)

		resB, err := tx.ExecContext(c, insertQuery,
			lastF.ID_SILO, lastF.SILONUMMER, lastF.HERDENR, lastF.ID_PERSON,
			prevDatum, mengeB, lastF.PREISDT, lastF.RABATTPROZ,
			nettoB, bruttoB, lastF.MWSTPROZ, mwstKzStr,
			prevDatum, prevDatum + " 12:00:00Z", lastF.AW, lastF.ID_FUTTERSORTEN,
		)
		if err != nil {
			log.Printf("[INVENTUR] Failed to insert Booking B: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Speichern der zweiten Futterbuchung (b): " + err.Error()})
			return
		}
		idB, _ := resB.LastInsertId()
		log.Printf("[INVENTUR] Booking B inserted successfully. ID=%d", idB)

		// 6. Ab diesem Datum (INVENTURDATUMALT) bis Inventurdatum - 1
		// Futterlieferungen abfragen
		var gesamtLiefermenge, gesamtNetto, gesamtBrutto float64
		err = tx.QueryRowContext(c, `
			SELECT COALESCE(SUM(LIEFERMENGE), 0.0), COALESCE(SUM(NETTO), 0.0), COALESCE(SUM(BRUTTO), 0.0)
			FROM FUTTER WHERE ID_SILO = ? AND LIEFERDATUM >= ? AND LIEFERDATUM <= ?`,
			siloID, inventurDatumAlt, prevDatum).Scan(&gesamtLiefermenge, &gesamtNetto, &gesamtBrutto)

		if err != nil {
			log.Printf("[INVENTUR] Failed to sum deliveries in period [%s, %s]: %v", inventurDatumAlt, prevDatum, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Berechnen der Gesamtlieferungen: " + err.Error()})
			return
		}
		log.Printf("[INVENTUR] Summed deliveries: Qty=%.2f, Net=%.2f, Gross=%.2f", gesamtLiefermenge, gesamtNetto, gesamtBrutto)

		// 7. Alle Buchungssätze für diesen Zeitraum und diese Silonummer (alle Herden des Silos) lesen
		// Errechne die FuttertageGesamt (= Tierbestand der Buchungen)
		var futtertageGesamt int64
		err = tx.QueryRowContext(c, `
			SELECT COALESCE(SUM(B.TIERBESTAND), 0) 
			FROM BUCHUNG B
			JOIN HERDEN H ON B.ID_HERDEN = H.ID
			WHERE H.ID_SILO = ? 
			  AND B.BUCHUNGSDATUM >= ? AND B.BUCHUNGSDATUM <= ?`,
			siloID, inventurDatumAlt, prevDatum).Scan(&futtertageGesamt)

		if err != nil {
			log.Printf("[INVENTUR] Failed to sum bird days: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Berechnen der Futtertage: " + err.Error()})
			return
		}
		log.Printf("[INVENTUR] Summed bird days: %d", futtertageGesamt)

		if futtertageGesamt <= 0 {
			log.Printf("[INVENTUR] Error: Bird days is <= 0")
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Keine Tierbestände in den Herdenbuchungen für den Zeitraum %s bis %s gefunden. Berechnung abgebrochen.", inventurDatumAlt, prevDatum)})
			return
		}

		// Futterverbrauch Tier (in Gramm) und Futterkosten Tier (in Euro) Errechnen
		futterverbrauchTierGrams := int64(math.Round((gesamtLiefermenge * 1000.0) / float64(futtertageGesamt)))
		futterkostenTier := gesamtNetto / float64(futtertageGesamt)
		log.Printf("[INVENTUR] Consumption: %d g/bird/day, Costs: %.6f EUR/bird/day", futterverbrauchTierGrams, futterkostenTier)

		// 8. Hole alle Buchungen für den Zeitraum (alle Herden des Silos) und speichere
		rows, err := tx.QueryContext(c, `
			SELECT B.ID, B.TIERBESTAND 
			FROM BUCHUNG B
			JOIN HERDEN H ON B.ID_HERDEN = H.ID
			WHERE H.ID_SILO = ? 
			  AND B.BUCHUNGSDATUM >= ? AND B.BUCHUNGSDATUM <= ?`,
			siloID, inventurDatumAlt, prevDatum)
		if err != nil {
			log.Printf("[INVENTUR] Failed to query bookings for update: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Lesen der Buchungssätze: " + err.Error()})
			return
		}
		defer rows.Close()

		type buchungUpdate struct {
			id          int64
			tierbestand int64
		}
		var updates []buchungUpdate
		for rows.Next() {
			var buID, tb int64
			if err := rows.Scan(&buID, &tb); err == nil {
				updates = append(updates, buchungUpdate{id: buID, tierbestand: tb})
			}
		}
		rows.Close()
		log.Printf("[INVENTUR] Found %d booking rows to update", len(updates))

		for _, up := range updates {
			futterKtag := int64(math.Round(futterkostenTier * float64(up.tierbestand)))
			_, err = tx.ExecContext(c, `
				UPDATE BUCHUNG
				SET FUTTERVERBRAUCHTIER = ?, FUTTERKTAG = ?, SILONR = ?
				WHERE ID = ?`,
				futterverbrauchTierGrams, futterKtag, silonummer, up.id)
			if err != nil {
				log.Printf("[INVENTUR] Failed to update booking ID %d: %v", up.id, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Fehler beim Aktualisieren der Buchung ID %d: %s", up.id, err.Error())})
				return
			}
		}
		log.Printf("[INVENTUR] Successfully updated all %d daily bookings", len(updates))

		// 9. Speichere das Inventurdatum (inventurDatum) in den Silosatz in das Feld Inventurdatum Alt,
		// und setze Inventurdatum Neu auf das neue Datum, und die Inventurfüllmenge.
		_, err = tx.ExecContext(c, `
			UPDATE SILO
			SET INVENTURDATUMALT = ?, INVENTURDATUMNEU = ?, INVENTURFUELLMENGE = ?
			WHERE ID = ?`,
			inventurDatum, inventurDatum, inventurMenge, siloID)
		if err != nil {
			log.Printf("[INVENTUR] Failed to update Silo: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren des Silosatzes: " + err.Error()})
			return
		}
		log.Printf("[INVENTUR] Successfully updated SILO table details")

		// 10. Commit transaction
		err = tx.Commit()
		if err != nil {
			log.Printf("[INVENTUR] Commit failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Commit der Transaktion: " + err.Error()})
			return
		}
		log.Printf("[INVENTUR] Transaction committed successfully!")

		c.JSON(http.StatusOK, gin.H{
			"message": "Futterinventur erfolgreich durchgeführt",
			"gesamt_liefermenge": gesamtLiefermenge,
			"gesamt_netto": gesamtNetto,
			"futtertage_gesamt": futtertageGesamt,
			"futterverbrauch_tier_g": futterverbrauchTierGrams,
			"futterkosten_tier_eur": futterkostenTier,
		})
	})

	r.DELETE("/api/silo/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		// Check if any herds or fodder bookings are using this silo
		var countH, countF int64
		_ = conn.QueryRowContext(c, "SELECT COUNT(*) FROM HERDEN WHERE ID_SILO = ?", idInt).Scan(&countH)
		_ = conn.QueryRowContext(c, "SELECT COUNT(*) FROM FUTTER WHERE ID_SILO = ?", idInt).Scan(&countF)

		if countH > 0 || countF > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Silo kann nicht gelöscht werden, da es noch Herden oder Futterbuchungen zugeordnet ist."})
			return
		}

		if err := queries.DeleteSilo(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Eilager & EilagerBuchung APIs
	r.GET("/api/eilager", func(c *gin.Context) {
		res, err := queries.ListEilager(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/eilagerbuchungen/lager/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		res, err := queries.ListEilagerBuchungenByLager(c, idInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/eilagerbuchungen/kz/:kz", func(c *gin.Context) {
		kz := c.Param("kz")
		res, err := queries.ListEilagerBuchungenByKZ(c, kz)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/eilagerbuchungen/sum-by-source/:buchungId/:lagerId", func(c *gin.Context) {
		bId, _ := strconv.ParseInt(c.Param("buchungId"), 10, 64)
		lId, _ := strconv.ParseInt(c.Param("lagerId"), 10, 64)
		res, err := queries.GetEilagerSumBySource(c, db.GetEilagerSumBySourceParams{
			IDBuchung:      bId,
			IDFremdeslager: lId,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/eilagerbuchungen/sum-by-buchung/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		res, err := queries.GetEilagerSumByBuchungID(c, idInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/eilagerbuchungen/group-balance/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}

		// 1. Get the booking
		b, err := queries.GetBuchung(c, idInt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Performance record not found"})
			return
		}

		kz := toString(b.Vermittelt)
		vam := toString(b.Vermitteltam)

		// 2. If it's a split record, get group totals
		if (kz == "V" || kz == "S") && vam != "" {
			perf, err := queries.GetBuchungGroupTotals(c, db.GetBuchungGroupTotalsParams{
				IDHerden:     b.IDHerden,
				Vermitteltam: b.Vermitteltam,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get group performance: " + err.Error()})
				return
			}
			used, err := queries.GetEilagerGroupTotals(c, db.GetEilagerGroupTotalsParams{
				IDHerden:     b.IDHerden,
				Vermitteltam: b.Vermitteltam,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get group storage sum: " + err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"is_group": true,
				"perf":     perf,
				"used":     used,
			})
			return
		}

		// 3. Normal booking: return single record totals
		perf := gin.H{
			"JUMBOS":   b.Kl6,
			"XL":       b.Xl,
			"LARGE":    b.Large,
			"MEDIUM":   b.Medium,
			"SMALL":    b.Small,
			"VOLLEIKG": b.Vollei,
		}
		used, _ := queries.GetEilagerSumByBuchungID(c, idInt)
		c.JSON(http.StatusOK, gin.H{
			"is_group": false,
			"perf":     perf,
			"used":     used,
		})
	})

	r.POST("/api/eilagerbuchungen", func(c *gin.Context) {
		var req struct {
			IDFremdeslager  *int64   `json:"ID_FREMDESLAGER"`
			IDBuchung       *int64   `json:"ID_BUCHUNG"`
			IDEilager       *int64   `json:"ID_EILAGER"`
			Buchungsdatum   *string  `json:"BUCHUNGSDATUM"`
			Jumbos          *int64   `json:"JUMBOS"`
			Xl              *int64   `json:"XL"`
			Large           *int64   `json:"LARGE"`
			Medium          *int64   `json:"MEDIUM"`
			Small           *int64   `json:"SMALL"`
			Volleikg        *float64 `json:"VOLLEIKG"`
			Schmutz         *int64   `json:"SCHMUTZ"`
			Knickeier       *int64   `json:"KNICKEIER"`
			Brucheier       *int64   `json:"BRUCHEIER"`
			Buchungstyp     *string  `json:"BUCHUNGSTYP"`
			Charge          *string  `json:"CHARGE"`
			KzLager         *string  `json:"KZ_LAGER"`
			IDFremdebuchung *int64   `json:"ID_FREMDEBUCHUNG"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		params := db.AddEilagerBuchungParams{}
		if req.IDFremdeslager != nil {
			params.IDFremdeslager = *req.IDFremdeslager
		}
		if req.IDBuchung != nil {
			params.IDBuchung = *req.IDBuchung
		}
		if req.IDEilager != nil {
			params.IDEilager = *req.IDEilager
		}
		if req.Buchungsdatum != nil {
			params.Buchungsdatum = *req.Buchungsdatum
		}
		if req.Jumbos != nil {
			params.Jumbos = *req.Jumbos
		}
		if req.Xl != nil {
			params.Xl = *req.Xl
		}
		if req.Large != nil {
			params.Large = *req.Large
		}
		if req.Medium != nil {
			params.Medium = *req.Medium
		}
		if req.Small != nil {
			params.Small = *req.Small
		}
		if req.Volleikg != nil {
			params.Volleikg = *req.Volleikg
		}
		if req.Schmutz != nil {
			params.Schmutz = *req.Schmutz
		}
		if req.Knickeier != nil {
			params.Knickeier = *req.Knickeier
		}
		if req.Brucheier != nil {
			params.Brucheier = *req.Brucheier
		}
		if req.Buchungstyp != nil {
			params.Buchungstyp = *req.Buchungstyp
		}
		if req.Charge != nil {
			params.Charge = *req.Charge
		}
		if req.KzLager != nil {
			params.KzLager = *req.KzLager
		}
		if req.IDFremdebuchung != nil {
			params.IDFremdebuchung = *req.IDFremdebuchung
		}

		if params.KzLager == "" || params.KzLager == "x" {
			if params.IDEilager != 0 {
				if l, err := queries.GetEilager(c, params.IDEilager); err == nil {
					if s, ok := l.Kz.(string); ok {
						params.KzLager = s
					}
				}
			}
		}

		// LOGIK FÜR UMBUCHUNGEN IN VERKAUFSLAGER (KZ 'V')
		isVerkaufLager := params.KzLager == "V"
		if !isVerkaufLager && params.IDEilager != 0 {
			if l, err := queries.GetEilager(c, params.IDEilager); err == nil && l.Kz == "V" {
				isVerkaufLager = true
			}
		}
		if !isVerkaufLager && params.IDFremdeslager != 0 {
			if l, err := queries.GetEilager(c, params.IDFremdeslager); err == nil && l.Kz == "V" {
				isVerkaufLager = true
			}
		}
		if isVerkaufLager {
			params.Verkauf = 1
		}

		// 1. Mengen von Quellbuchung abziehen
		if params.IDFremdebuchung > 0 {
			ebSource, err := queries.GetEilagerBuchung(c, params.IDFremdebuchung)
			if err == nil {
				updateParams := db.UpdateEilagerBuchungParams{
					ID:              ebSource.ID,
					IDFremdeslager:  ebSource.IDFremdeslager,
					IDBuchung:       ebSource.IDBuchung,
					IDEilager:       ebSource.IDEilager,
					Buchungsdatum:   ebSource.Buchungsdatum,
					Jumbos:          ebSource.Jumbos - params.Jumbos,
					Xl:              ebSource.Xl - params.Xl,
					Large:           ebSource.Large - params.Large,
					Medium:          ebSource.Medium - params.Medium,
					Small:           ebSource.Small - params.Small,
					Volleikg:        ebSource.Volleikg - params.Volleikg,
					Schmutz:         ebSource.Schmutz - params.Schmutz,
					Knickeier:       ebSource.Knickeier - params.Knickeier,
					Brucheier:       ebSource.Brucheier - params.Brucheier,
					Buchungstyp:     ebSource.Buchungstyp,
					Charge:          ebSource.Charge,
					KzLager:         ebSource.KzLager,
					IDFremdebuchung: ebSource.IDFremdebuchung,
					Verkauf:         ebSource.Verkauf,
				}
				_, _ = queries.UpdateEilagerBuchung(c, updateParams)
			}
		}

		res, err := queries.AddEilagerBuchung(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if isVerkaufLager {
			// 2. Verkauf-Satz erstellen
			prices, _ := queries.ListEierpreise(c)
			var pS, pM, pL, pXL float64
			for _, p := range prices {
				switch p.Eierklasse {
				case "S":
					pS = extractFloat64(p.PreisVon)
				case "M":
					pM = extractFloat64(p.PreisVon)
				case "L":
					pL = extractFloat64(p.PreisVon)
				case "XL":
					pXL = extractFloat64(p.PreisVon)
				}
			}

			gesamt := ((float64(params.Small) * pS) +
				(float64(params.Medium) * pM) +
				(float64(params.Large) * pL) +
				(float64(params.Xl) * pXL)) / 100.0

			verkaufParams := db.CreateVerkaufParams{
				IDEilagerbuchung: res.ID,
				IDBuchung:        params.IDBuchung,
				Buchungsdatum:    params.Buchungsdatum,
				Mengesmall:       params.Small,
				Mengemedium:      params.Medium,
				Mengelarge:       params.Large,
				Mengexl:          params.Xl,
				Preissmall:       pS / 100.0,
				Preismedium:      pM / 100.0,
				Preislarge:       pL / 100.0,
				Preisxl:          pXL / 100.0,
				Gesamtpreis:      gesamt,
				Bio:              false,
				Verbucht:         false,
				Charge:           params.Charge,
				Rabattprozent:    0,
			}
			_, _ = queries.CreateVerkauf(c, verkaufParams)
		}

		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/eilagerbuchungen/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		var req struct {
			IDFremdeslager  *int64   `json:"ID_FREMDESLAGER"`
			IDBuchung       *int64   `json:"ID_BUCHUNG"`
			IDEilager       *int64   `json:"ID_EILAGER"`
			Buchungsdatum   *string  `json:"BUCHUNGSDATUM"`
			Jumbos          *int64   `json:"JUMBOS"`
			Xl              *int64   `json:"XL"`
			Large           *int64   `json:"LARGE"`
			Medium          *int64   `json:"MEDIUM"`
			Small           *int64   `json:"SMALL"`
			Volleikg        *float64 `json:"VOLLEIKG"`
			Schmutz         *int64   `json:"SCHMUTZ"`
			Knickeier       *int64   `json:"KNICKEIER"`
			Brucheier       *int64   `json:"BRUCHEIER"`
			Buchungstyp     *string  `json:"BUCHUNGSTYP"`
			Charge          *string  `json:"CHARGE"`
			KzLager         *string  `json:"KZ_LAGER"`
			IDFremdebuchung *int64   `json:"ID_FREMDEBUCHUNG"`
			Verkauf         *bool    `json:"VERKAUF"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		params := db.UpdateEilagerBuchungParams{ID: idInt}
		if req.IDFremdeslager != nil {
			params.IDFremdeslager = *req.IDFremdeslager
		}
		if req.IDBuchung != nil {
			params.IDBuchung = *req.IDBuchung
		}
		if req.IDEilager != nil {
			params.IDEilager = *req.IDEilager
		}
		if req.Buchungsdatum != nil {
			params.Buchungsdatum = *req.Buchungsdatum
		}
		if req.Jumbos != nil {
			params.Jumbos = *req.Jumbos
		}
		if req.Xl != nil {
			params.Xl = *req.Xl
		}
		if req.Large != nil {
			params.Large = *req.Large
		}
		if req.Medium != nil {
			params.Medium = *req.Medium
		}
		if req.Small != nil {
			params.Small = *req.Small
		}
		if req.Volleikg != nil {
			params.Volleikg = *req.Volleikg
		}
		if req.Schmutz != nil {
			params.Schmutz = *req.Schmutz
		}
		if req.Knickeier != nil {
			params.Knickeier = *req.Knickeier
		}
		if req.Brucheier != nil {
			params.Brucheier = *req.Brucheier
		}
		if req.Buchungstyp != nil {
			params.Buchungstyp = *req.Buchungstyp
		}
		if req.Charge != nil {
			params.Charge = *req.Charge
		}
		if req.KzLager != nil {
			params.KzLager = *req.KzLager
		}
		if req.IDFremdebuchung != nil {
			params.IDFremdebuchung = *req.IDFremdebuchung
		}
		if req.Verkauf != nil {
			if *req.Verkauf {
				params.Verkauf = 1
			} else {
				params.Verkauf = 0
			}
		}

		// Altdaten holen
		old, err := queries.GetEilagerBuchung(c, idInt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Buchung nicht gefunden"})
			return
		}

		// 1. Mengen-Differenz auf Quellbuchung anpassen
		if params.IDFremdebuchung > 0 && params.IDFremdebuchung == old.IDFremdebuchung {
			ebSource, err := queries.GetEilagerBuchung(c, params.IDFremdebuchung)
			if err == nil {
				updateParams := db.UpdateEilagerBuchungParams{
					ID:              ebSource.ID,
					IDFremdeslager:  ebSource.IDFremdeslager,
					IDBuchung:       ebSource.IDBuchung,
					IDEilager:       ebSource.IDEilager,
					Buchungsdatum:   ebSource.Buchungsdatum,
					Jumbos:          ebSource.Jumbos + old.Jumbos - params.Jumbos,
					Xl:              ebSource.Xl + old.Xl - params.Xl,
					Large:           ebSource.Large + old.Large - params.Large,
					Medium:          ebSource.Medium + old.Medium - params.Medium,
					Small:           ebSource.Small + old.Small - params.Small,
					Volleikg:        ebSource.Volleikg + old.Volleikg - params.Volleikg,
					Schmutz:         ebSource.Schmutz + old.Schmutz - params.Schmutz,
					Knickeier:       ebSource.Knickeier + old.Knickeier - params.Knickeier,
					Brucheier:       ebSource.Brucheier + old.Brucheier - params.Brucheier,
					Buchungstyp:     ebSource.Buchungstyp,
					Charge:          ebSource.Charge,
					KzLager:         ebSource.KzLager,
					IDFremdebuchung: ebSource.IDFremdebuchung,
					Verkauf:         ebSource.Verkauf,
				}
				_, _ = queries.UpdateEilagerBuchung(c, updateParams)
			}
		}

		// 2. Verkauf-Toggle Logik
		if old.Verkauf == 1 && params.Verkauf == 0 {
			// Verkauf löschen
			v, err := queries.GetVerkaufByEilagerbuchung(c, idInt)
			if err == nil {
				if v.Verbucht {
					c.JSON(http.StatusForbidden, gin.H{"error": "Zugehöriger Verkauf ist bereits verbucht. Flag kann nicht entfernt werden."})
					return
				}
				_ = queries.DeleteVerkauf(c, v.ID)
			}
		} else if old.Verkauf == 0 && params.Verkauf == 1 {
			// Verkauf neu anlegen
			prices, _ := queries.ListEierpreise(c)
			var pS, pM, pL, pXL float64
			for _, p := range prices {
				switch p.Eierklasse {
				case "S":
					pS = extractFloat64(p.PreisVon)
				case "M":
					pM = extractFloat64(p.PreisVon)
				case "L":
					pL = extractFloat64(p.PreisVon)
				case "XL":
					pXL = extractFloat64(p.PreisVon)
				}
			}
			gesamt := ((float64(params.Small) * pS) + (float64(params.Medium) * pM) + (float64(params.Large) * pL) + (float64(params.Xl) * pXL)) / 100.0
			_, _ = queries.CreateVerkauf(c, db.CreateVerkaufParams{
				IDEilagerbuchung: idInt,
				IDBuchung:        params.IDBuchung,
				Buchungsdatum:    params.Buchungsdatum,
				Mengesmall:       params.Small,
				Mengemedium:      params.Medium,
				Mengelarge:       params.Large,
				Mengexl:          params.Xl,
				Preissmall:       pS / 100.0,
				Preismedium:      pM / 100.0,
				Preislarge:       pL / 100.0,
				Preisxl:          pXL / 100.0,
				Gesamtpreis:      gesamt,
				Bio:              false,
				Verbucht:         false,
				Charge:           params.Charge,
				Rabattprozent:    0,
			})
		}

		if params.KzLager == "" || params.KzLager == "x" {
			if params.IDEilager != 0 {
				if l, err := queries.GetEilager(c, params.IDEilager); err == nil {
					if s, ok := l.Kz.(string); ok {
						params.KzLager = s
					}
				}
			}
		}

		res, err := queries.UpdateEilagerBuchung(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/eilagerbuchungen/:id", func(c *gin.Context) {
		idInt, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		eb, err := queries.GetEilagerBuchung(c, idInt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Buchung nicht gefunden"})
			return
		}

		// 1. Verkauf check
		if eb.Verkauf == 1 {
			v, err := queries.GetVerkaufByEilagerbuchung(c, idInt)
			if err == nil {
				if v.Verbucht {
					c.JSON(http.StatusForbidden, gin.H{"error": "Zugehöriger Verkauf ist bereits verbucht. Löschen nicht möglich."})
					return
				}
				_ = queries.DeleteVerkauf(c, v.ID)
			}
		}

		// 2. Mengen an Quellbuchung zurückgeben
		if eb.IDFremdebuchung > 0 {
			ebSource, err := queries.GetEilagerBuchung(c, eb.IDFremdebuchung)
			if err == nil {
				updateParams := db.UpdateEilagerBuchungParams{
					ID:              ebSource.ID,
					IDFremdeslager:  ebSource.IDFremdeslager,
					IDBuchung:       ebSource.IDBuchung,
					IDEilager:       ebSource.IDEilager,
					Buchungsdatum:   ebSource.Buchungsdatum,
					Jumbos:          ebSource.Jumbos + eb.Jumbos,
					Xl:              ebSource.Xl + eb.Xl,
					Large:           ebSource.Large + eb.Large,
					Medium:          ebSource.Medium + eb.Medium,
					Small:           ebSource.Small + eb.Small,
					Volleikg:        ebSource.Volleikg + eb.Volleikg,
					Schmutz:         ebSource.Schmutz + eb.Schmutz,
					Knickeier:       ebSource.Knickeier + eb.Knickeier,
					Brucheier:       ebSource.Brucheier + eb.Brucheier,
					Buchungstyp:     ebSource.Buchungstyp,
					Charge:          ebSource.Charge,
					KzLager:         ebSource.KzLager,
					IDFremdebuchung: ebSource.IDFremdebuchung,
					Verkauf:         ebSource.Verkauf,
				}
				_, _ = queries.UpdateEilagerBuchung(c, updateParams)
			}
		}

		if err := queries.DeleteEilagerBuchung(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	r.POST("/api/eilager", func(c *gin.Context) {
		var req struct {
			Lagernummer   int64   `json:"lagernummer" binding:"required"`
			Kz            string  `json:"KZ"`
			Bezeichnung   string  `json:"BEZEICHNUNG"`
			LetzteBuchung string  `json:"LETZTE_BUCHUNG"`
			Jumbos        int64   `json:"JUMBOS"`
			Xl            int64   `json:"XL"`
			Large         int64   `json:"LARGE"`
			Medium        int64   `json:"MEDIUM"`
			Small         int64   `json:"SMALL"`
			Volleikg      float64 `json:"VOLLEIKG"`
			Aw            int64   `json:"AW"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.CreateEilagerParams{
			Lagernummer:   req.Lagernummer,
			Kz:            req.Kz,
			Bezeichnung:   req.Bezeichnung,
			LetzteBuchung: req.LetzteBuchung,
			Jumbos:        0, // Bestände bei Neuanlage auf 0 initialisieren
			Xl:            0,
			Large:         0,
			Medium:        0,
			Small:         0,
			Volleikg:      0,
			Aw:            req.Aw,
		}
		res, err := queries.CreateEilager(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/eilager/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		var req struct {
			Lagernummer   int64  `json:"lagernummer" binding:"required"`
			Kz            string `json:"KZ"`
			Bezeichnung   string `json:"BEZEICHNUNG"`
			LetzteBuchung string `json:"LETZTE_BUCHUNG"`
			Aw            int64  `json:"AW"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Bei Update holen wir erst den aktuellen Bestand, da dieser nur über Buchungen geändert werden darf
		current, err := queries.GetEilager(c, idInt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Eilager nicht gefunden"})
			return
		}

		params := db.UpdateEilagerParams{
			ID:            idInt,
			Lagernummer:   req.Lagernummer,
			Kz:            req.Kz,
			Bezeichnung:   req.Bezeichnung,
			LetzteBuchung: req.LetzteBuchung,
			Jumbos:        current.Jumbos,
			Xl:            current.Xl,
			Large:         current.Large,
			Medium:        current.Medium,
			Small:         current.Small,
			Volleikg:      current.Volleikg,
			Aw:            req.Aw,
		}
		res, err := queries.UpdateEilager(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/eilager/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		// Check if used in herds or bookings
		var countH, countB int64
		_ = conn.QueryRowContext(c, "SELECT COUNT(*) FROM HERDEN WHERE ID_EILAGER = ?", idInt).Scan(&countH)
		_ = conn.QueryRowContext(c, "SELECT COUNT(*) FROM EILAGERBUCHUNG WHERE ID_EILAGER = ?", idInt).Scan(&countB)

		if countH > 0 || countB > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Eilager kann nicht gelöscht werden, da es noch Herden oder Buchungen zugeordnet ist."})
			return
		}

		if err := queries.DeleteEilager(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Lagerplatz APIs
	r.GET("/api/lagerplatz", func(c *gin.Context) {
		res, err := queries.ListLagerplaetze(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/lagerplatz/eilager/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		res, err := queries.ListLagerplaetzeByEilager(c, idInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/lagerplatz", func(c *gin.Context) {
		var req struct {
			IDEilager   int64  `json:"id_eilager" binding:"required"`
			Bezeichnung string `json:"bezeichnung" binding:"required"`
			Bemerkung   string `json:"BEMERKUNG"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.CreateLagerplatzParams{
			IDEilager:   req.IDEilager,
			Bezeichnung: req.Bezeichnung,
			Bemerkung:   req.Bemerkung,
		}
		res, err := queries.CreateLagerplatz(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/lagerplatz/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		var req struct {
			IDEilager   int64  `json:"id_eilager" binding:"required"`
			Bezeichnung string `json:"bezeichnung" binding:"required"`
			Bemerkung   string `json:"BEMERKUNG"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateLagerplatzParams{
			ID:          idInt,
			IDEilager:   req.IDEilager,
			Bezeichnung: req.Bezeichnung,
			Bemerkung:   req.Bemerkung,
		}
		res, err := queries.UpdateLagerplatz(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/lagerplatz/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		if err := queries.DeleteLagerplatz(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	r.GET("/api/eilager/bestandsuebersicht", func(c *gin.Context) {
		idStr := c.Query("id_eilager")
		idInt := int64(0)
		if idStr != "" {
			var err error
			idInt, err = strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
				return
			}
		}

		res, err := queries.GetBestandsuebersicht(c, db.GetBestandsuebersichtParams{
			Column1:   idInt,
			IDEilager: idInt,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	// Person APIs
	r.GET("/api/person", func(c *gin.Context) {
		res, err := queries.ListPersonen(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/person/zuechter", func(c *gin.Context) {
		res, err := queries.ListZuechter(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/person/lieferant", func(c *gin.Context) {
		res, err := queries.ListLieferanten(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/company/person", func(c *gin.Context) {
		res, err := queries.GetCompanyPerson(c)
		if err != nil {
			if err == sql.ErrNoRows {
				// Create default company record if not exist
				defaultCompany, err := queries.CreatePerson(c, db.CreatePersonParams{
					Name:  "Meine Firma",
					Kz:    "F",
					Ort:   "Musterstadt",
					Email: "info@firma.de",
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create default company"})
					return
				}
				c.JSON(http.StatusOK, defaultCompany)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/person", func(c *gin.Context) {
		var req struct {
			IDTexte        int64  `json:"ID_TEXTE"`
			IDAnrede       int64  `json:"ID_ANREDE"`
			Personennummer int64  `json:"PERSONENNUMMER"`
			Kz             string `json:"KZ"`
			Postfach       string `json:"POSTFACH"`
			Name           string `json:"name" binding:"required"`
			Firma          string `json:"FIRMA"`
			Strasse        string `json:"STRASSE"`
			Plz            string `json:"PLZ"`
			Ort            string `json:"ORT"`
			Telefon        string `json:"TELEFON"`
			Mobiltelephon  string `json:"MOBILTELEPHON"`
			Email          string `json:"EMAIL"`
			Email2         string `json:"EMAIL2"`
			Foto           string `json:"FOTO"`
			Homepage       string `json:"HOMEPAGE"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.CreatePersonParams{
			IDTexte:        req.IDTexte,
			IDAnrede:       req.IDAnrede,
			Personennummer: req.Personennummer,
			Kz:             req.Kz,
			Postfach:       req.Postfach,
			Name:           req.Name,
			Firma:          req.Firma,
			Strasse:        req.Strasse,
			Plz:            req.Plz,
			Ort:            req.Ort,
			Telefon:        req.Telefon,
			Mobiltelephon:  req.Mobiltelephon,
			Email:          req.Email,
			Email2:         req.Email2,
			Foto:           decodeBase64(req.Foto),
			Homepage:       req.Homepage,
		}
		res, err := queries.CreatePerson(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/person/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		var req struct {
			IDTexte        int64  `json:"ID_TEXTE"`
			IDAnrede       int64  `json:"ID_ANREDE"`
			Personennummer int64  `json:"PERSONENNUMMER"`
			Kz             string `json:"KZ"`
			Postfach       string `json:"POSTFACH"`
			Name           string `json:"name" binding:"required"`
			Firma          string `json:"FIRMA"`
			Strasse        string `json:"STRASSE"`
			Plz            string `json:"PLZ"`
			Ort            string `json:"ORT"`
			Telefon        string `json:"TELEFON"`
			Mobiltelephon  string `json:"MOBILTELEPHON"`
			Email          string `json:"EMAIL"`
			Email2         string `json:"EMAIL2"`
			Foto           string `json:"FOTO"`
			Homepage       string `json:"HOMEPAGE"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdatePersonParams{
			ID:             idInt,
			IDTexte:        req.IDTexte,
			IDAnrede:       req.IDAnrede,
			Personennummer: req.Personennummer,
			Kz:             req.Kz,
			Postfach:       req.Postfach,
			Name:           req.Name,
			Firma:          req.Firma,
			Strasse:        req.Strasse,
			Plz:            req.Plz,
			Ort:            req.Ort,
			Telefon:        req.Telefon,
			Mobiltelephon:  req.Mobiltelephon,
			Email:          req.Email,
			Email2:         req.Email2,
			Foto:           decodeBase64(req.Foto),
			Homepage:       req.Homepage,
		}
		res, err := queries.UpdatePerson(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/person/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		if err := queries.DeletePerson(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Verkauf APIs
	r.GET("/api/verkauf", func(c *gin.Context) {
		res, err := queries.ListVerkauf(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/verkauf/:id", func(c *gin.Context) {
		idInt, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		res, err := queries.GetVerkauf(c, idInt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/verkauf", func(c *gin.Context) {
		var req struct {
			IDEilagerbuchung int64   `json:"ID_EILAGERBUCHUNG"`
			IDBuchung        int64   `json:"ID_BUCHUNG"`
			Buchungsdatum    string  `json:"BUCHUNGSDATUM"`
			Mengesmall       int64   `json:"MENGESMALL"`
			Mengemedium      int64   `json:"MENGEMEDIUM"`
			Mengelarge       int64   `json:"MENGELARGE"`
			Mengexl          int64   `json:"MENGEXL"`
			Preissmall       float64 `json:"PREISSMALL"`
			Preismedium      float64 `json:"PREISMEDIUM"`
			Preislarge       float64 `json:"PREISLARGE"`
			Preisxl          float64 `json:"PREISXL"`
			Gesamtpreis      float64 `json:"GESAMTPREIS"`
			Bio              bool    `json:"BIO"`
			Verbucht         bool    `json:"VERBUCHT"`
			Charge           string  `json:"CHARGE"`
			Rabattprozent    float64 `json:"RABATTPROZENT"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		prices, _ := queries.ListEierpreise(c)
		var pS, pM, pL, pXL float64
		for _, p := range prices {
			switch p.Eierklasse {
			case "S":
				pS = extractFloat64(p.PreisVon)
			case "M":
				pM = extractFloat64(p.PreisVon)
			case "L":
				pL = extractFloat64(p.PreisVon)
			case "XL":
				pXL = extractFloat64(p.PreisVon)
			}
		}

		gesamt := ((float64(req.Mengesmall) * pS) +
			(float64(req.Mengemedium) * pM) +
			(float64(req.Mengelarge) * pL) +
			(float64(req.Mengexl) * pXL)) / 100.0

		// Charge aus Eilagerbuchung holen, falls vorhanden und nicht überschrieben
		var chargeStr = req.Charge
		if req.IDEilagerbuchung > 0 && chargeStr == "" {
			_ = conn.QueryRowContext(c, "SELECT CHARGE FROM EILAGERBUCHUNG WHERE ID = ?", req.IDEilagerbuchung).Scan(&chargeStr)
		}

		params := db.CreateVerkaufParams{
			IDEilagerbuchung: req.IDEilagerbuchung,
			IDBuchung:        req.IDBuchung,
			Buchungsdatum:    req.Buchungsdatum,
			Mengesmall:       req.Mengesmall,
			Mengemedium:      req.Mengemedium,
			Mengelarge:       req.Mengelarge,
			Mengexl:          req.Mengexl,
			Preissmall:       pS / 100.0,
			Preismedium:      pM / 100.0,
			Preislarge:       pL / 100.0,
			Preisxl:          pXL / 100.0,
			Gesamtpreis:      gesamt,
			Bio:              req.Bio,
			Verbucht:         req.Verbucht,
			Charge:           chargeStr,
			Rabattprozent:    req.Rabattprozent,
		}
		res, err := queries.CreateVerkauf(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/verkauf/:id", func(c *gin.Context) {
		idInt, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req struct {
			IDEilagerbuchung int64   `json:"ID_EILAGERBUCHUNG"`
			Buchungsdatum    string  `json:"BUCHUNGSDATUM"`
			Mengesmall       int64   `json:"MENGESMALL"`
			Mengemedium      int64   `json:"MENGEMEDIUM"`
			Mengelarge       int64   `json:"MENGELARGE"`
			Mengexl          int64   `json:"MENGEXL"`
			Preissmall       float64 `json:"PREISSMALL"`
			Preismedium      float64 `json:"PREISMEDIUM"`
			Preislarge       float64 `json:"PREISLARGE"`
			Preisxl          float64 `json:"PREISXL"`
			Gesamtpreis      float64 `json:"GESAMTPREIS"`
			Bio              bool    `json:"BIO"`
			Verbucht         bool    `json:"VERBUCHT"`
			Charge           string  `json:"CHARGE"`
			Rabattprozent    float64 `json:"RABATTPROZENT"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		params := db.UpdateVerkaufParams{
			ID:               idInt,
			IDEilagerbuchung: req.IDEilagerbuchung,
			Buchungsdatum:    req.Buchungsdatum,
			Mengesmall:       req.Mengesmall,
			Mengemedium:      req.Mengemedium,
			Mengelarge:       req.Mengelarge,
			Mengexl:          req.Mengexl,
			Preissmall:       req.Preissmall,
			Preismedium:      req.Preismedium,
			Preislarge:       req.Preislarge,
			Preisxl:          req.Preisxl,
			Gesamtpreis:      req.Gesamtpreis,
			Bio:              req.Bio,
			Verbucht:         req.Verbucht,
			Charge:           req.Charge,
			Rabattprozent:    req.Rabattprozent,
		}
		res, err := queries.UpdateVerkauf(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/verkauf/:id", func(c *gin.Context) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Verkauf kann nur über die zugehörige Lagerbuchung gelöscht werden."})
	})

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now()})
	})

	// Stall APIs
	r.GET("/api/stall", func(c *gin.Context) {
		res, err := queries.ListStaelle(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/stall", func(c *gin.Context) {
		var req struct {
			Stallnummer int64  `json:"stallnummer" binding:"required"`
			Bezeichnung string `json:"bezeichnung" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.CreateStallParams{
			Stallnummer: req.Stallnummer,
			Bezeichnung: req.Bezeichnung,
		}
		res, err := queries.CreateStall(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/stall/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		var req struct {
			Stallnummer int64  `json:"stallnummer" binding:"required"`
			Bezeichnung string `json:"bezeichnung" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateStallParams{
			ID:          idInt,
			Stallnummer: req.Stallnummer,
			Bezeichnung: req.Bezeichnung,
		}
		res, err := queries.UpdateStall(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/stall/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		// Check if any herds are assigned to this stall
		var count int64
		err = conn.QueryRowContext(c, "SELECT COUNT(*) FROM HERDEN WHERE ID_STALL = ?", idInt).Scan(&count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler bei der Belegungsprüfung: " + err.Error()})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Stall kann nicht gelöscht werden, da ihm noch Herden zugeordnet sind."})
			return
		}

		if err := queries.DeleteStall(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Herden APIs
	r.GET("/api/herden", func(c *gin.Context) {
		herden, err := queries.ListHerden(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		type EnrichedHerde struct {
			db.ListHerdenRow
			EggStats    gin.H                    `json:"EGGSTATS"`
			WeeklyStats []map[string]interface{} `json:"WEEKLYSTATS"`
		}

		enriched := make([]EnrichedHerde, len(herden))
		for i, h := range herden {
			enriched[i] = EnrichedHerde{
				ListHerdenRow: h,
			}

			// EggStats
			params := h.ID
			eggRes, _ := queries.GetEggStatsByHerde(c, db.GetEggStatsByHerdeParams{
				ID:       params,
				IDHerden: params,
			})
			enriched[i].EggStats = gin.H{
				"SUM_KLASSE_A": toInt64(eggRes.SumKlasseA),
				"SUM_SMALL":    toInt64(eggRes.SumSmall),
				"SUM_MEDIUM":   toInt64(eggRes.SumMedium),
				"SUM_LARGE":    toInt64(eggRes.SumLarge),
				"SUM_XL":       toInt64(eggRes.SumXl),
				"SUM_VERLUSTE": toInt64(eggRes.SumVerluste),
			}

			// WeeklyStats
			weekRes, _ := queries.GetEggStatsWeeklyByHerde(c, h.ID)
			var weekly []map[string]interface{}
			for _, row := range weekRes {
				weekly = append(weekly, map[string]interface{}{
					"LEBENSWOCHE":   row.Lebenswoche,
					"LETZTES_DATUM": toString(row.LetztesDatum),
					"SUM_KLASSE_A":  toInt64(row.SumKlasseA),
					"SUM_SMALL":     toInt64(row.SumSmall),
					"SUM_MEDIUM":    toInt64(row.SumMedium),
					"SUM_LARGE":     toInt64(row.SumLarge),
					"SUM_XL":        toInt64(row.SumXl),
					"SUM_VERLUSTE":  toInt64(row.SumVerluste),
				})
			}
			enriched[i].WeeklyStats = weekly
		}

		c.JSON(http.StatusOK, enriched)
	})

	r.GET("/api/herden/lookup", func(c *gin.Context) {
		res, err := queries.ListHerdenLookup(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/herden/:id/eggstats", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		res, err := queries.GetEggStatsByHerde(c, db.GetEggStatsByHerdeParams{
			ID:       idInt,
			IDHerden: idInt,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// logic handled by package-level toInt64

		stats := gin.H{
			"SUM_KLASSE_A": toInt64(res.SumKlasseA),
			"SUM_SMALL":    toInt64(res.SumSmall),
			"SUM_MEDIUM":   toInt64(res.SumMedium),
			"SUM_LARGE":    toInt64(res.SumLarge),
			"SUM_XL":       toInt64(res.SumXl),
			"SUM_VERLUSTE": toInt64(res.SumVerluste),
		}

		c.JSON(http.StatusOK, stats)
	})

	r.GET("/api/herden/:id/eggstats/filtered", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		year := c.Query("year")
		quarterStr := c.DefaultQuery("quarter", "0")
		quarterInt, _ := strconv.ParseInt(quarterStr, 10, 64)
		monthStr := c.DefaultQuery("month", "0")
		monthInt, _ := strconv.ParseInt(monthStr, 10, 64)
		activeOnlyStr := c.DefaultQuery("onlyActive", "0")
		activeOnlyInt, _ := strconv.ParseInt(activeOnlyStr, 10, 64)

		res, err := queries.GetEggStatsByHerdeFiltered(c, db.GetEggStatsByHerdeFilteredParams{
			IDHerden:   idInt,
			OnlyActive: activeOnlyInt,
			Year:       year,
			Quarter:    quarterInt,
			Month:      monthInt,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		stats := gin.H{
			"SUM_KLASSE_A": toInt64(res.SumKlasseA),
			"SUM_SMALL":    toInt64(res.SumSmall),
			"SUM_MEDIUM":   toInt64(res.SumMedium),
			"SUM_LARGE":    toInt64(res.SumLarge),
			"SUM_XL":       toInt64(res.SumXl),
			"SUM_VERLUSTE": toInt64(res.SumVerluste),
		}
		c.JSON(http.StatusOK, stats)
	})

	r.GET("/api/herden/:id/latest_booking", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		res, err := queries.GetLatestBookingByHerde(c, idInt)
		if err != nil {
			// If no booking found, return 404 or empty object
			c.JSON(http.StatusNotFound, gin.H{"error": "no booking found"})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/herden/:id/years", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		activeOnlyStr := c.DefaultQuery("onlyActive", "0")
		activeOnlyInt, _ := strconv.ParseInt(activeOnlyStr, 10, 64)

		res, err := queries.GetEggBookingYears(c, db.GetEggBookingYearsParams{
			IDHerden:   idInt,
			OnlyActive: activeOnlyInt,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/herden/:id/weeklystats", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		res, err := queries.GetEggStatsWeeklyByHerde(c, idInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// logic handled by package-level toInt64

		// Ensure it's mapped to a list of dicts with numerical values
		var list []map[string]interface{}
		for _, row := range res {
			list = append(list, map[string]interface{}{
				"LEBENSWOCHE":   row.Lebenswoche,
				"LETZTES_DATUM": toString(row.LetztesDatum),
				"SUM_KLASSE_A":  toInt64(row.SumKlasseA),
				"SUM_SMALL":     toInt64(row.SumSmall),
				"SUM_MEDIUM":    toInt64(row.SumMedium),
				"SUM_LARGE":     toInt64(row.SumLarge),
				"SUM_XL":        toInt64(row.SumXl),
				"SUM_VERLUSTE":  toInt64(row.SumVerluste),
			})
		}

		c.JSON(http.StatusOK, list)
	})
	// Hilfsfunktion zur Bereinigung von Datumswerten
	sanitizeDate := func(d string) string {
		if len(d) > 10 {
			return d[:10]
		}
		return d
	}

	sanitizeDateTime := func(d string) string {
		// MariaDB mag kein "T" oder "Z" in DATETIME Feldern
		d = strings.ReplaceAll(d, "T", " ")
		d = strings.ReplaceAll(d, "Z", "")
		if len(d) > 19 {
			return d[:19]
		}
		return d
	}

	r.POST("/api/herden", func(c *gin.Context) {
		var req struct {
			Herdennummer          int64   `json:"herdennummer" binding:"required"`
			IdRasse               int64   `json:"id_rasse" binding:"required"`
			IDZuechter            int64   `json:"ID_ZUECHTER"`
			IDEilager             int64   `json:"ID_EILAGER"`
			Anfangsbestand        int64   `json:"ANFANGSBESTAND"`
			Einstalldatum         string  `json:"EINSTALLDATUM"`
			Legedatum             string  `json:"LEGEDATUM"`
			Einstallkosten        float64 `json:"EINSTALLKOSTEN"`
			IDSilo                int64   `json:"ID_SILO"`
			IDStall               int64   `json:"ID_STALL"`
			Aktiv                 int64   `json:"AKTIV"`
			Bezeichnung           string  `json:"BEZEICHNUNG"`
			AlleBuchungenMitDatum int64   `json:"ALLE_BUCHUNGEN_MIT_DATUM"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[API] POST /api/herden - Bind Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.CreateHerdeParams{
			Herdennummer:          req.Herdennummer,
			Bezeichnung:           req.Bezeichnung,
			IDRasse:               req.IdRasse,
			IDZuechter:            req.IDZuechter,
			IDEilager:             req.IDEilager,
			Anfangsbestand:        req.Anfangsbestand,
			Einstalldatum:         sanitizeDate(req.Einstalldatum),
			Legedatum:             sanitizeDate(req.Legedatum),
			Einstallkosten:        req.Einstallkosten,
			IDSilo:                req.IDSilo,
			IDStall:               req.IDStall,
			Aktiv:                 req.Aktiv,
			Allebuchungenmitdatum: req.AlleBuchungenMitDatum,
		}
		log.Printf("[API] POST /api/herden - Creating: %+v", params)
		res, err := queries.CreateHerde(c, params)
		if err != nil {
			log.Printf("[API] POST /api/herden - Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/herden/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req struct {
			Herdennummer          int64   `json:"herdennummer" binding:"required"`
			IdRasse               int64   `json:"id_rasse" binding:"required"`
			IDZuechter            int64   `json:"ID_ZUECHTER"`
			IDEilager             int64   `json:"ID_EILAGER"`
			Anfangsbestand        int64   `json:"ANFANGSBESTAND"`
			Einstalldatum         string  `json:"EINSTALLDATUM"`
			Legedatum             string  `json:"LEGEDATUM"`
			Einstallkosten        float64 `json:"EINSTALLKOSTEN"`
			IDSilo                int64   `json:"ID_SILO"`
			IDStall               int64   `json:"ID_STALL"`
			Aktiv                 int64   `json:"AKTIV"`
			Bezeichnung           string  `json:"BEZEICHNUNG"`
			AlleBuchungenMitDatum int64   `json:"ALLE_BUCHUNGEN_MIT_DATUM"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[API] PUT /api/herden/%d - Bind Error: %v", idInt, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateHerdeParams{
			ID:                    idInt,
			Herdennummer:          req.Herdennummer,
			Bezeichnung:           req.Bezeichnung,
			IDRasse:               req.IdRasse,
			IDZuechter:            req.IDZuechter,
			IDEilager:             req.IDEilager,
			Anfangsbestand:        req.Anfangsbestand,
			Einstalldatum:         sanitizeDate(req.Einstalldatum),
			Legedatum:             sanitizeDate(req.Legedatum),
			Einstallkosten:        req.Einstallkosten,
			IDSilo:                req.IDSilo,
			IDStall:               req.IDStall,
			Aktiv:                 req.Aktiv,
			Allebuchungenmitdatum: req.AlleBuchungenMitDatum,
		}
		log.Printf("[API] PUT /api/herden/%d - Updating: %+v", idInt, params)
		res, err := queries.UpdateHerde(c, params)
		if err != nil {
			log.Printf("[API] PUT /api/herden/%d - Error: %v", idInt, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/leistung/verlust", func(c *gin.Context) {
		var req struct {
			IDHerden int64  `json:"id_herden" binding:"required"`
			Verluste int64  `json:"verluste" binding:"required"`
			IDTexte  int64  `json:"id_texte" binding:"required"`
			Datum    string `json:"DATUM"`
			Memo     string `json:"MEMO"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get herde to obtain herdennummer
		herde, err := queries.GetHerde(c, req.IDHerden)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Herde nicht gefunden"})
			return
		}

		// 1. Tierbewegung erstellen (Typ 'V' = Verlust)
		_, err = queries.CreateTierbewegung(c, db.CreateTierbewegungParams{
			Herdennummer:   toNullInt64(herde.Herdennummer),
			IDBuchung:      sql.NullInt64{Valid: false},
			Typ:            "V",
			IDTexte:        toNullInt64(req.IDTexte),
			Bewegungsdatum: toNullString(req.Datum),
			Bewegungen:     toNullInt64(req.Verluste),
			IDHerdenVon:    sql.NullInt64{Valid: false},
			IDHerdenNach:   sql.NullInt64{Valid: false},
			Kosten:         toNullFloat64(0),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tierbewegung konnte nicht erstellt werden: " + err.Error()})
			return
		}

		// 2. Bestand in Herden-Tabelle reduzieren
		err = queries.UpdateHerdeStock(c, db.UpdateHerdeStockParams{
			Anfangsbestand: req.Verluste,
			ID:             req.IDHerden,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 3. Neuen Bestand laden für Feedback
		updatedHerde, err := queries.GetHerde(c, req.IDHerden)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Bestand konnte nicht gelesen werden"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"message":  "Verlust erfolgreich gebucht",
			"newStock": updatedHerde.Anfangsbestand,
		})
	})

	r.POST("/api/leistung/pseudo-booking", func(c *gin.Context) {
		var req struct {
			IDLeistung   int64  `json:"ID_LEISTUNG" binding:"required"`
			IDEilager    int64  `json:"ID_EILAGER" binding:"required"`
			KzVerwendung string `json:"KZ_VERWENDUNG" binding:"required"`
			Datum        string `json:"DATUM"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 1. Buchung (Leistung) laden
		buchung, err := queries.GetBuchung(c, req.IDLeistung)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Buchung nicht gefunden: " + err.Error()})
			return
		}

		// 2. Bereits verbuchte Mengen laden
		sums, err := queries.GetEilagerSumByBuchungID(c, req.IDLeistung)
		if err != nil && err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Summen konnten nicht geladen werden: " + err.Error()})
			return
		}

		// 3. Eilager (Pseudolager) laden für KZ_LAGER
		lager, err := queries.GetEilager(c, req.IDEilager)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Eilager nicht gefunden: " + err.Error()})
			return
		}

		// 4. Differenzen berechnen (was noch übrig ist)
		remJumbos := buchung.Kl6 - sums.Jumbos
		remXl := buchung.Xl - sums.Xl
		remLarge := buchung.Large - sums.Large
		remMedium := buchung.Medium - sums.Medium
		remSmall := buchung.Small - sums.Small
		remVollei := buchung.Vollei - sums.Volleikg
		remSchmutz := buchung.Schmutz - sums.Schmutz
		remKnick := buchung.Knickeier - sums.Knickeier
		remBruch := buchung.Brucheier - sums.Brucheier

		// Nur buchen, wenn überhaupt etwas übrig ist
		if remJumbos <= 0 && remXl <= 0 && remLarge <= 0 && remMedium <= 0 && remSmall <= 0 && remVollei <= 0 && remSchmutz <= 0 && remKnick <= 0 && remBruch <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Keine restlichen Eier zum Buchen vorhanden"})
			return
		}

		// 5. Neue Eilagerbuchung erstellen
		_, err = queries.AddEilagerBuchung(c, db.AddEilagerBuchungParams{
			IDBuchung:     req.IDLeistung,
			IDEilager:     req.IDEilager,
			Buchungsdatum: req.Datum,
			Jumbos:        remJumbos,
			Xl:            remXl,
			Large:         remLarge,
			Medium:        remMedium,
			Small:         remSmall,
			Volleikg:      remVollei,
			Schmutz:       remSchmutz,
			Knickeier:     remKnick,
			Brucheier:     remBruch,
			Buchungstyp:   req.KzVerwendung, // Wir speichern das Verwendungskennzeichen im Typ
			KzLager:       toString(lager.Kz),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Erstellen der Pseudo-Buchung: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Pseudolager-Buchung erfolgreich"})
	})

	r.POST("/api/herden/verlust", func(c *gin.Context) {
		var req struct {
			IDHerden int64  `json:"id_herden" binding:"required"`
			Verluste int64  `json:"verluste" binding:"required"`
			IDTexte  int64  `json:"ID_TEXTE"`
			Datum    string `json:"DATUM"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get herde to obtain herdennummer
		herde, err := queries.GetHerde(c, req.IDHerden)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Herde nicht gefunden"})
			return
		}

		// 1. Tierbewegung erstellen
		_, err = queries.CreateTierbewegung(c, db.CreateTierbewegungParams{
			Herdennummer:   toNullInt64(herde.Herdennummer),
			IDBuchung:      sql.NullInt64{Valid: false},
			Typ:            "V",
			IDTexte:        toNullInt64(req.IDTexte),
			Bewegungsdatum: toNullString(req.Datum),
			Bewegungen:     toNullInt64(req.Verluste),
			IDHerdenVon:    sql.NullInt64{Valid: false},
			IDHerdenNach:   sql.NullInt64{Valid: false},
			Kosten:         toNullFloat64(0),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tierbewegung konnte nicht erstellt werden: " + err.Error()})
			return
		}

		// 2. Bestand in Herden-Tabelle reduzieren
		err = queries.UpdateHerdeStock(c, db.UpdateHerdeStockParams{
			Anfangsbestand: req.Verluste,
			ID:             req.IDHerden,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Bestand konnte nicht aktualisiert werden: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Verlust erfolgreich gebucht"})
	})

	r.GET("/api/admin/latest-date", func(c *gin.Context) {
		var maxDate sql.NullString
		// Wir ermitteln das aktuellste Datum aus der Tabelle 'buchung'
		err := conn.QueryRowContext(c, "SELECT max(BUCHUNGSDATUM) FROM BUCHUNG").Scan(&maxDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"max_date": maxDate})
	})

	r.POST("/api/admin/shift-test-dates", func(c *gin.Context) {
		var req struct {
			Days int `json:"DAYS"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Transaktion starten
		tx, err := conn.BeginTx(c, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback()

		days := req.Days

		// 1. Alle relevanten UNIQUE Indizes finden, die Datums-Spalten enthalten
		type idxDef struct {
			Name string
			SQL  string
		}
		var indices []idxDef

		idxRows, err := tx.QueryContext(c, `
			SELECT name, sql FROM sqlite_master 
			WHERE type = 'index' AND sql IS NOT NULL AND sql LIKE '%UNIQUE%' AND (
				sql LIKE '%BUCHUNGSDATUM%' OR sql LIKE '%LIEFERDATUM%' OR sql LIKE '%DATUM%' OR 
				sql LIKE '%LEGEDATUM%' OR sql LIKE '%EINSTALLDATUM%' OR sql LIKE '%INVENTURDATUMALT%' OR 
				sql LIKE '%INVENTURDATUMNEU%' OR sql LIKE '%BEWEGUNGSDATUM%' OR sql LIKE '%letzte_buchung%' OR 
				sql LIKE '%Anlagedatum%'
			)
		`)

		if err == nil {
			for idxRows.Next() {
				var idm idxDef
				if err := idxRows.Scan(&idm.Name, &idm.SQL); err == nil {
					indices = append(indices, idm)
				}
			}
			idxRows.Close()
		}

		// 2. Diese Indizes temporär droppen
		for _, idx := range indices {
			_, _ = tx.ExecContext(c, fmt.Sprintf("DROP INDEX IF EXISTS %s", idx.Name))
		}

		// Liste der Updates mit exaktem Schema-Casing
		updates := []struct {
			sql      string
			mysqlSql string
			args     []interface{}
		}{
			{
				"UPDATE BUCHUNG SET BUCHUNGSDATUM = date(BUCHUNGSDATUM, ? || ' days')",
				"UPDATE BUCHUNG SET BUCHUNGSDATUM = DATE(DATE_ADD(BUCHUNGSDATUM, INTERVAL ? DAY))",
				[]interface{}{days},
			},
			{
				"UPDATE FUTTER SET LIEFERDATUM = date(LIEFERDATUM, ? || ' days'), DATUM = date(DATUM, ? || ' days')",
				"UPDATE FUTTER SET LIEFERDATUM = DATE(DATE_ADD(LIEFERDATUM, INTERVAL ? DAY)), DATUM = DATE(DATE_ADD(DATUM, INTERVAL ? DAY))",
				[]interface{}{days, days},
			},
			{
				"UPDATE HERDEN SET LEGEDATUM = date(LEGEDATUM, ? || ' days'), EINSTALLDATUM = date(EINSTALLDATUM, ? || ' days')",
				"UPDATE HERDEN SET LEGEDATUM = DATE(DATE_ADD(LEGEDATUM, INTERVAL ? DAY)), EINSTALLDATUM = DATE(DATE_ADD(EINSTALLDATUM, INTERVAL ? DAY))",
				[]interface{}{days, days},
			},
			{
				"UPDATE SILO SET INVENTURDATUMALT = date(INVENTURDATUMALT, ? || ' days'), INVENTURDATUMNEU = date(INVENTURDATUMNEU, ? || ' days')",
				"UPDATE SILO SET INVENTURDATUMALT = DATE(DATE_ADD(INVENTURDATUMALT, INTERVAL ? DAY)), INVENTURDATUMNEU = DATE(DATE_ADD(INVENTURDATUMNEU, INTERVAL ? DAY))",
				[]interface{}{days, days},
			},
			{
				"UPDATE TIERBEWEGUNGEN SET BEWEGUNGSDATUM = date(BEWEGUNGSDATUM, ? || ' days')",
				"UPDATE TIERBEWEGUNGEN SET BEWEGUNGSDATUM = DATE(DATE_ADD(BEWEGUNGSDATUM, INTERVAL ? DAY))",
				[]interface{}{days},
			},
			{
				"UPDATE EILAGERBUCHUNG SET BUCHUNGSDATUM = date(BUCHUNGSDATUM, ? || ' days')",
				"UPDATE EILAGERBUCHUNG SET BUCHUNGSDATUM = DATE(DATE_ADD(BUCHUNGSDATUM, INTERVAL ? DAY))",
				[]interface{}{days},
			},
			{
				"UPDATE EILAGER SET letzte_buchung = date(letzte_buchung, ? || ' days')",
				"UPDATE EILAGER SET letzte_buchung = DATE(DATE_ADD(letzte_buchung, INTERVAL ? DAY))",
				[]interface{}{days},
			},
			{
				"UPDATE KOSTEN SET BUCHUNGSDATUM = date(BUCHUNGSDATUM, ? || ' days')",
				"UPDATE KOSTEN SET BUCHUNGSDATUM = DATE(DATE_ADD(KOSTEN.BUCHUNGSDATUM, INTERVAL ? DAY))",
				[]interface{}{days},
			},
			{
				"UPDATE TABELLENKOPF SET Anlagedatum = date(Anlagedatum, ? || ' days'), DATUM = date(DATUM, ? || ' days')",
				"UPDATE TABELLENKOPF SET Anlagedatum = DATE(DATE_ADD(Anlagedatum, INTERVAL ? DAY)), DATUM = DATE(DATE_ADD(DATUM, INTERVAL ? DAY))",
				[]interface{}{days, days},
			},
			{
				"UPDATE AKTIONEN SET AKTIONSDATUM = date(AKTIONSDATUM, ? || ' days')",
				"UPDATE AKTIONEN SET AKTIONSDATUM = DATE(DATE_ADD(AKTIONSDATUM, INTERVAL ? DAY))",
				[]interface{}{days},
			},
			{
				"UPDATE VERKAUF SET BUCHUNGSDATUM = date(BUCHUNGSDATUM, ? || ' days')",
				"UPDATE VERKAUF SET BUCHUNGSDATUM = DATE(DATE_ADD(BUCHUNGSDATUM, INTERVAL ? DAY))",
				[]interface{}{days},
			},
		}

		for _, up := range updates {
			query := up.sql
			if database.Engine == "mysql" {
				query = up.mysqlSql
			}
			_, err = tx.ExecContext(c, query, up.args...)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("SQL-Fehler in [%s]: %v", query, err)})
				return
			}
		}

		// 3. Alle Indizes wiederherstellen
		for _, idx := range indices {
			_, err = tx.ExecContext(c, idx.SQL)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Fehler beim Wiederherstellen des Index %s: %v", idx.Name, err)})
				return
			}
		}

		if err = tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "days_added": days})
	})

	r.DELETE("/api/herden/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		if err := queries.DeleteHerde(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// LegacyBuchungRow stellt sicher, dass die JSON-Tags großgeschrieben werden (für das Frontend)
	type LegacyBuchungRow struct {
		ID                   int64   `json:"ID"`
		IDHerden             int64   `json:"ID_HERDEN"`
		Lw                   int64   `json:"LW"`
		Herdennummer         int64   `json:"HERDENNUMMER"`
		Buchungsdatum        string  `json:"BUCHUNGSDATUM"`
		Gewichtprobe         int64   `json:"GEWICHTPROBE"`
		Kontrollgewicht      float64 `json:"KONTROLLGEWICHT"`
		Klassea              int64   `json:"KLASSEA"`
		Verluste             int64   `json:"VERLUSTE"`
		Eimasse              float64 `json:"EIMASSE"`
		Schmutz              int64   `json:"SCHMUTZ"`
		Knickeier            int64   `json:"KNICKEIER"`
		Vollei               float64 `json:"VOLLEI"`
		Brucheier            int64   `json:"BRUCHEIER"`
		Tierbestand          int64   `json:"TIERBESTAND"`
		IDEitabelle          int64   `json:"ID_EITABELLE"`
		IDDgewichttab        int64   `json:"ID_DGEWICHTTAB"`
		Futterktag           int64   `json:"FUTTERKTAG"`
		Silonr               int64   `json:"SILONR"`
		Kl6                  int64   `json:"KL6"`
		Vermitteltam         string  `json:"VERMITTELTAM"`
		Small                int64   `json:"SMALL"`
		Large                int64   `json:"LARGE"`
		Medium               int64   `json:"MEDIUM"`
		Xl                   int64   `json:"XL"`
		Zeitstempel          string  `json:"ZEITSTEMPEL"`
		Dgewichtei           float64 `json:"DGEWICHTEI"`
		Aw                   int64   `json:"AW"`
		Vermittelt           string  `json:"VERMITTELT"`
		Futterverbrauchtier  int64   `json:"FUTTERVERBRAUCHTIER"`
		HerdenNummerRel      int64   `json:"HERDEN_NUMMER_REL"`
		HerdenBezeichnungRel string  `json:"HERDEN_BEZEICHNUNG_REL"`
	}

	// Buchung APIs
	r.GET("/api/buchung", func(c *gin.Context) {
		res, err := queries.ListBuchungen(c)
		if err != nil {
			log.Printf("[API] GET /api/buchung - Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Mapping auf Legacy-Struktur mit Großbuchstaben
		legacyList := make([]LegacyBuchungRow, len(res))
		for i, v := range res {
			legacyList[i] = LegacyBuchungRow{
				ID:              v.ID,
				IDHerden:        v.IDHerden,
				Lw:              v.Lw,
				Herdennummer:    v.Herdennummer,
				Buchungsdatum:   sanitizeDate(v.Buchungsdatum),
				Gewichtprobe:    v.Gewichtprobe,
				Kontrollgewicht: v.Kontrollgewicht,
				Klassea:         v.Klassea,
				Verluste:        v.Verluste,
				Eimasse:         v.Eimasse,
				Schmutz:         v.Schmutz,
				Knickeier:       v.Knickeier,
				Vollei:          v.Vollei,
				Brucheier:       v.Brucheier,
				Tierbestand:     v.Tierbestand,
				IDEitabelle:     v.IDEitabelle,
				IDDgewichttab:   v.IDDgewichttab,
				Futterktag:      v.Futterktag,
				Silonr:          v.Silonr,
				Kl6:             v.Kl6,
				Vermitteltam:    sanitizeDate(v.Vermitteltam),
				Small:           v.Small,
				Large:           v.Large,
				Medium:          v.Medium,
				Xl:              v.Xl,
				Zeitstempel:     v.Zeitstempel,
				Dgewichtei:      v.Dgewichtei,
				Aw:              v.Aw,
				Vermittelt:      toString(v.Vermittelt),
				Futterverbrauchtier: v.Futterverbrauchtier,
				HerdenNummerRel: v.HerdenNummerRel.Int64,
			}
		}

		log.Printf("[API] GET /api/buchung - Found %d records, mapped to Legacy format", len(legacyList))
		c.JSON(http.StatusOK, legacyList)
	})

	r.GET("/api/buchung/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		res, err := queries.GetBuchung(c, idInt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/buchung/last-info/:id_herden", func(c *gin.Context) {
		idStr := c.Param("id_herden")
		herdeID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		res, err := queries.GetLatestBookingByHerde(c, herdeID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusOK, gin.H{"buchungsdatum": nil, "tierbestand": nil})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"buchungsdatum": res.Buchungsdatum,
			"tierbestand":   res.Tierbestand,
		})
	})

	r.POST("/api/calculate-distribution", func(c *gin.Context) {
		var req struct {
			IDHerden        int64   `json:"ID_HERDEN"`
			Kontrollgewicht float64 `json:"KONTROLLGEWICHT"`
			Gewichtprobe    int64   `json:"GEWICHTPROBE"`
			Verpackung      float64 `json:"VERPACKUNG"`
			Klassea         int64   `json:"KLASSEA"`
			Lw              int     `json:"LW"`
			Aw              int     `json:"AW"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 1. Parameter laden
		params, err := queries.GetFirmenparameterByHerde(c, req.IDHerden)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Parameter konnten nicht geladen werden"})
			return
		}

		// Erlaube Berechnung wenn entweder Klasseaerfassen aktiv ist ODER eine der Aufteilungen
		canCalculate := params.Klasseaerfassen == 1 || params.Aufteilunggewicht == 1 || params.Aufteilungalter == 1
		if !canCalculate {
			c.JSON(http.StatusOK, gin.H{"active": false, "reason": "No classification mode active"})
			return
		}

		var dWeight float64
		// Pfad A (AufteilungGewicht = true)
		if params.Aufteilunggewicht == 1 {
			anzahl := req.Gewichtprobe
			if anzahl <= 0 {
				anzahl = params.Anzahlkontrollw
			}
			verpackung := req.Verpackung
			if verpackung <= 0 {
				verpackung = params.Verpackungkg
			}

			if req.Kontrollgewicht > 0 && anzahl > 0 {
				dWeight = GetDurchschnittsgewicht(req.Kontrollgewicht, verpackung, anzahl)
			}
		} else {
			// Pfad B (Else): Nutze KlassenAufteilungAlter(Lebenswoche)
			idTabAlter := params.IDTabellealter
			if idTabAlter > 0 && req.Aw > 0 {
				dWeight, err = GetEigewichtByAlter(c, queries, idTabAlter, req.Aw)
				if err != nil {
					dWeight = 0
				}
			}
		}

		if dWeight > 0 {
			idTabGewicht := params.IDTabellegewicht
			if idTabGewicht > 0 {
				s, m, l, xl, err := GetEierverteilungByGewicht(c, queries, req.Klassea, idTabGewicht, dWeight)
				if err != nil {
					if err == sql.ErrNoRows {
						c.JSON(http.StatusOK, gin.H{
							"active":     true,
							"dgewicht":   dWeight,
							"error":      fmt.Sprintf("Kein Tabelleneintrag für %.1f g gefunden", dWeight),
							"calculated": false,
						})
					} else {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Verteilungsfehler: " + err.Error()})
					}
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"active":     true,
					"dgewicht":   dWeight,
					"small":      s,
					"medium":     m,
					"large":      l,
					"xl":         xl,
					"calculated": true,
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"active": true, "calculated": false})
	})

	r.POST("/api/buchung", func(c *gin.Context) {
		var req struct {
			IDHerden        int64   `json:"ID_HERDEN"`
			Lw              int64   `json:"LW"`
			Herdennummer    int64   `json:"HERDENNUMMER"`
			Buchungsdatum   string  `json:"BUCHUNGSDATUM"`
			Gewichtprobe    int64   `json:"GEWICHTPROBE"`
			Kontrollgewicht float64 `json:"KONTROLLGEWICHT"`
			Klassea         int64   `json:"KLASSEA"`
			Verluste        int64   `json:"VERLUSTE"`
			Eimasse         float64 `json:"EIMASSE"`
			Schmutz         int64   `json:"SCHMUTZ"`
			Knickeier       int64   `json:"KNICKEIER"`
			Vollei          float64 `json:"VOLLEI"`
			Brucheier       int64   `json:"BRUCHEIER"`
			Tierbestand     int64   `json:"TIERBESTAND"`
			IDEitabelle     int64   `json:"ID_EITABELLE"`
			IDDgewichttab   int64   `json:"ID_DGEWICHTTAB"`
			Futterktag      int64   `json:"FUTTERKTAG"`
			Silonr          int64   `json:"SILONR"`
			Kl6             int64   `json:"KL6"`
			Vermitteltam    string  `json:"VERMITTELTAM"`
			Small           int64   `json:"SMALL"`
			Large           int64   `json:"LARGE"`
			Medium          int64   `json:"MEDIUM"`
			Xl              int64   `json:"XL"`
			Zeitstempel     string  `json:"ZEITSTEMPEL"`
			Dgewichtei      float64 `json:"DGEWICHTEI"`
			Aw              int64   `json:"AW"`
			Vermittelt      string  `json:"VERMITTELT"`
			Futterverbrauchtier int64 `json:"FUTTERVERBRAUCHTIER"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[API] POST /api/buchung - Bind Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		req.Buchungsdatum = sanitizeDate(req.Buchungsdatum)
		req.Vermitteltam = sanitizeDate(req.Vermitteltam)
		req.Zeitstempel = sanitizeDateTime(req.Zeitstempel)

		siloNr := req.Silonr
		if siloNr <= 0 {
			siloNr = getSiloNrForHerd(c, conn, req.IDHerden)
		}

		log.Printf("[API] POST /api/buchung - Received: %+v, Resolved SiloNr: %d", req, siloNr)

		// Dublettenprüfung
		var count int64
		err := conn.QueryRowContext(c, "SELECT COUNT(*) FROM BUCHUNG WHERE ID_HERDEN = ? AND BUCHUNGSDATUM = ?", req.IDHerden, req.Buchungsdatum).Scan(&count)
		if err != nil {
			log.Printf("[API] POST /api/buchung - Duplicate Check Error: %v", err)
		}
		if err == nil && count > 0 {
			log.Printf("[API] POST /api/buchung - Duplicate found for Herde %d at %s", req.IDHerden, req.Buchungsdatum)
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Doppelte Buchung nicht zulässig (Herde: %d, Datum: %s)", req.IDHerden, req.Buchungsdatum)})
			return
		}

		log.Printf("[API] POST /api/buchung - Loading parameters for Herde %d...", req.IDHerden)
		// Parameter und letzte Buchung laden
		paramsRow, err := queries.GetFirmenparameterByHerde(c, req.IDHerden)
		if err != nil {
			log.Printf("[API] POST /api/buchung - GetFirmenparameter Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Laden der Parameter: " + err.Error()})
			return
		}
		log.Printf("[API] POST /api/buchung - Parameters loaded: %+v", paramsRow)

		log.Printf("[API] POST /api/buchung - Loading latest booking for Herde %d...", req.IDHerden)
		lastBooking, err := queries.GetLatestBookingByHerde(c, req.IDHerden)

		doVermittlung := false
		var diffDays int
		var lastDate time.Time
		var newDate time.Time

		// Das aktuelle Buchungsdatum parsen (vorher säubern!)
		newDateStr := sanitizeDate(req.Buchungsdatum)
		newDate, _ = time.Parse("2006-01-02", newDateStr)

		if err == nil && lastBooking.Buchungsdatum != "" {
			log.Printf("[API] POST /api/buchung - Latest booking found: %+v", lastBooking)
			lastDate, _ = time.Parse("2006-01-02", sanitizeDate(lastBooking.Buchungsdatum))
			diffDays = int(newDate.Sub(lastDate).Hours() / 24)
		} else {
			log.Printf("[API] POST /api/buchung - No latest booking found, trying fallback to herd start date...")
			// NEU: Fallback auf das Legedatum der Herde, falls noch gar keine Buchung existiert
			if h, err := queries.GetHerde(c, req.IDHerden); err == nil && h.Legedatum != "" {
				log.Printf("[API] POST /api/buchung - Fallback to herd Legedatum: %s", h.Legedatum)
				lastDate, _ = time.Parse("2006-01-02", sanitizeDate(h.Legedatum))
				lastDate = lastDate.AddDate(0, 0, -1) // ALT = Legedatum - 1 Tag
				diffDays = int(newDate.Sub(lastDate).Hours() / 24)
			} else {
				log.Printf("[API] POST /api/buchung - Fallback failed, no Legedatum found for Herde %d", req.IDHerden)
			}
		}

		log.Printf("[API] POST /api/buchung - Calculated diffDays: %d (Max allowed: %d)", diffDays, paramsRow.Maxtagevermitteln)

		// Lückenlosigkeits-Prüfung / Vermittlung
		if diffDays > 0 {
			maxDays := paramsRow.Maxtagevermitteln
			if maxDays > 0 && diffDays > int(maxDays) {
				c.JSON(http.StatusConflict, gin.H{
					"error": fmt.Sprintf("Buchung nicht möglich: Der Abstand zum letzten Datum (%d Tage) überschreitet das Limit von %d Tagen (Parameter 'MaxtageVermitteln').", diffDays, maxDays),
				})
				return
			}

			if paramsRow.Klasseavermitteln == 1 {
				if diffDays > 1 {
					doVermittlung = true
				}
			}
		}

		if doVermittlung {
			// VERMITTLUNGS-LOGIK: Erzeuge mehrere Sätze
			vGroupDate := time.Now().Format("2006-01-02")

			// NEU: Lücke säubern, um Doppelerfassungen zu vermeiden
			// Zuerst verknüpfte Lagerbuchungen löschen
			_, err = conn.ExecContext(c, `
				DELETE FROM EILAGERBUCHUNG 
				WHERE ID_BUCHUNG IN (
					SELECT ID FROM BUCHUNG 
					WHERE ID_HERDEN = ? AND BUCHUNGSDATUM > ? AND BUCHUNGSDATUM <= ?
				)`, req.IDHerden, lastDate.Format("2006-01-02"), req.Buchungsdatum)
			if err != nil {
				log.Printf("Warnung: Konnte verknüpfte Lagerbuchungen nicht säubern: %v", err)
			}

			// Dann die Buchungen selbst löschen
			_, err = conn.ExecContext(c, "DELETE FROM BUCHUNG WHERE ID_HERDEN = ? AND BUCHUNGSDATUM > ? AND BUCHUNGSDATUM <= ?",
				req.IDHerden, lastDate.Format("2006-01-02"), req.Buchungsdatum)
			if err != nil {
				log.Printf("Warnung: Konnte Lücke nicht säubern: %v", err)
			}

			// Hilfsvariablen zur Verteilung
			distInt := func(total int64, isLast bool) int64 {
				base := total / int64(diffDays)
				if isLast {
					return base + (total % int64(diffDays))
				}
				return base
			}
			distFloat := func(total float64, isLast bool) float64 {
				base := float64(int(total*100/float64(diffDays))) / 100.0 // Auf 2 Stellen kürzen
				if isLast {
					// Den kompletten Rest auf den letzten Tag
					usedSoFar := base * float64(diffDays-1)
					return total - usedSoFar
				}
				return base
			}

			log.Printf("[API] POST /api/buchung - Starting Vermittlung for %d days...", diffDays)
			var lastCreatedRes db.Buchung
			for i := 1; i <= diffDays; i++ {
				isLast := (i == diffDays)
				currentDate := lastDate.AddDate(0, 0, i).Format("2006-01-02")

				p := db.CreateBuchungParams{
					IDHerden:        req.IDHerden,
					Lw:              req.Lw,
					Herdennummer:    req.Herdennummer,
					Buchungsdatum:   currentDate,
					Gewichtprobe:    req.Gewichtprobe,
					Kontrollgewicht: req.Kontrollgewicht,
					Klassea:         distInt(req.Klassea, isLast),
					Verluste:        distInt(req.Verluste, isLast),
					Eimasse:         distFloat(req.Eimasse, isLast),
					Schmutz:         distInt(req.Schmutz, isLast),
					Knickeier:       distInt(req.Knickeier, isLast),
					Vollei:          distFloat(req.Vollei, isLast),
					Brucheier:       distInt(req.Brucheier, isLast),
					Tierbestand:     req.Tierbestand,
					IDEitabelle:     req.IDEitabelle,
					IDDgewichttab:   req.IDDgewichttab,
					Futterktag:      req.Futterktag,
					Silonr:          siloNr,
					Kl6:             distInt(req.Kl6, isLast),
					Vermitteltam:    vGroupDate,
					Small:           distInt(req.Small, isLast),
					Large:           distInt(req.Large, isLast),
					Medium:          distInt(req.Medium, isLast),
					Xl:              distInt(req.Xl, isLast),
					Dgewichtei:      req.Dgewichtei,
					Zeitstempel:     req.Zeitstempel,
					Aw:              req.Aw,
					Vermittelt:      "V", // Vermittelt
					Futterverbrauchtier: req.Futterverbrauchtier,
				}
				res, err := queries.CreateBuchung(c, p)
				if err != nil {
					log.Printf("[API] POST /api/buchung - Vermittlung Error at day %d (%s): %v", i, currentDate, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler bei Vermittlungsschleife: " + err.Error()})
					return
				}
				lastCreatedRes = res
				// Jede Teilbuchung automatisch ins Eilager buchen (wenn Parameter aktiv)
				_ = doAutomaticEilagerBuchung(c, conn, queries, res.ID)
			}
			log.Printf("[API] POST /api/buchung - Vermittlung finished successfully.")
			c.JSON(http.StatusOK, lastCreatedRes)
			return
		}

		// NORMALE BUCHUNG: Nur ein Satz
		params := db.CreateBuchungParams{
			IDHerden:        req.IDHerden,
			Lw:              req.Lw,
			Herdennummer:    req.Herdennummer,
			Buchungsdatum:   req.Buchungsdatum,
			Gewichtprobe:    req.Gewichtprobe,
			Kontrollgewicht: req.Kontrollgewicht,
			Klassea:         req.Klassea,
			Verluste:        req.Verluste,
			Eimasse:         req.Eimasse,
			Schmutz:         req.Schmutz,
			Knickeier:       req.Knickeier,
			Vollei:          req.Vollei,
			Brucheier:       req.Brucheier,
			Tierbestand:     req.Tierbestand,
			IDEitabelle:     req.IDEitabelle,
			IDDgewichttab:   req.IDDgewichttab,
			Futterktag:      req.Futterktag,
			Silonr:          siloNr,
			Kl6:             req.Kl6,
			Vermitteltam:    req.Vermitteltam,
			Small:           req.Small,
			Large:           req.Large,
			Medium:          req.Medium,
			Xl:              req.Xl,
			Dgewichtei:      req.Dgewichtei,
			Zeitstempel:     req.Zeitstempel,
			Aw:              req.Aw,
			Vermittelt:      "N", // Normal
			Futterverbrauchtier: req.Futterverbrauchtier,
		}
		res, err := queries.CreateBuchung(c, params)
		if err != nil {
			log.Printf("[API] POST /api/buchung - Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		_ = doAutomaticEilagerBuchung(c, conn, queries, res.ID)
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/buchung/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req struct {
			IDHerden        int64   `json:"ID_HERDEN"`
			Lw              int64   `json:"LW"`
			Herdennummer    int64   `json:"HERDENNUMMER"`
			Buchungsdatum   string  `json:"BUCHUNGSDATUM"`
			Gewichtprobe    int64   `json:"GEWICHTPROBE"`
			Kontrollgewicht float64 `json:"KONTROLLGEWICHT"`
			Klassea         int64   `json:"KLASSEA"`
			Verluste        int64   `json:"VERLUSTE"`
			Eimasse         float64 `json:"EIMASSE"`
			Schmutz         int64   `json:"SCHMUTZ"`
			Knickeier       int64   `json:"KNICKEIER"`
			Vollei          float64 `json:"VOLLEI"`
			Brucheier       int64   `json:"BRUCHEIER"`
			Tierbestand     int64   `json:"TIERBESTAND"`
			IDEitabelle     int64   `json:"ID_EITABELLE"`
			IDDgewichttab   int64   `json:"ID_DGEWICHTTAB"`
			Futterktag      int64   `json:"FUTTERKTAG"`
			Silonr          int64   `json:"SILONR"`
			Kl6             int64   `json:"KL6"`
			Vermitteltam    string  `json:"VERMITTELTAM"`
			Small           int64   `json:"SMALL"`
			Large           int64   `json:"LARGE"`
			Medium          int64   `json:"MEDIUM"`
			Xl              int64   `json:"XL"`
			Zeitstempel     string  `json:"ZEITSTEMPEL"`
			Dgewichtei      float64 `json:"DGEWICHTEI"`
			Aw              int64   `json:"AW"`
			Vermittelt      string  `json:"VERMITTELT"`
			Futterverbrauchtier int64 `json:"FUTTERVERBRAUCHTIER"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		req.Buchungsdatum = sanitizeDate(req.Buchungsdatum)
		req.Vermitteltam = sanitizeDate(req.Vermitteltam)

		existing, err := queries.GetBuchung(c, idInt)
		var existingSilonr, existingFvt, existingFkt int64
		if err == nil {
			existingSilonr = existing.Silonr
			existingFvt = existing.Futterverbrauchtier
			existingFkt = existing.Futterktag
		}

		siloNr := req.Silonr
		if siloNr <= 0 {
			siloNr = existingSilonr
		}
		if siloNr <= 0 {
			siloNr = getSiloNrForHerd(c, conn, req.IDHerden)
		}

		fvt := req.Futterverbrauchtier
		if fvt <= 0 {
			fvt = existingFvt
		}

		fkt := req.Futterktag
		if fkt <= 0 {
			fkt = existingFkt
		}

		// Dublettenprüfung (unter Ausschluss des eigenen Datensatzes)
		var count int64
		err = conn.QueryRowContext(c, "SELECT COUNT(*) FROM BUCHUNG WHERE ID_HERDEN = ? AND BUCHUNGSDATUM = ? AND ID != ?", req.IDHerden, req.Buchungsdatum, idInt).Scan(&count)
		if err == nil && count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Doppelte Buchung nicht zulässig"})
			return
		}

		params := db.UpdateBuchungParams{
			ID:              idInt,
			IDHerden:        req.IDHerden,
			Lw:              req.Lw,
			Herdennummer:    req.Herdennummer,
			Buchungsdatum:   req.Buchungsdatum,
			Gewichtprobe:    req.Gewichtprobe,
			Kontrollgewicht: req.Kontrollgewicht,
			Klassea:         req.Klassea,
			Verluste:        req.Verluste,
			Eimasse:         req.Eimasse,
			Schmutz:         req.Schmutz,
			Knickeier:       req.Knickeier,
			Vollei:          req.Vollei,
			Brucheier:       req.Brucheier,
			Tierbestand:     req.Tierbestand,
			IDEitabelle:     req.IDEitabelle,
			IDDgewichttab:   req.IDDgewichttab,
			Futterktag:      fkt,
			Silonr:          siloNr,
			Kl6:             req.Kl6,
			Vermitteltam:    req.Vermitteltam,
			Small:           req.Small,
			Large:           req.Large,
			Medium:          req.Medium,
			Xl:              req.Xl,
			Dgewichtei:      req.Dgewichtei,
			Zeitstempel:     req.Zeitstempel,
			Aw:              req.Aw,
			Vermittelt:      req.Vermittelt,
			Futterverbrauchtier: fvt,
		}
		res, err := queries.UpdateBuchung(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		_ = doAutomaticEilagerBuchung(c, conn, queries, res.ID)
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/buchung/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		// 1. Zuerst den Datensatz laden, um zu prüfen ob es eine Vermittlung ist
		res, err := queries.GetBuchung(c, idInt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Buchung nicht gefunden"})
			return
		}

		// KZ auslesen
		kz := toString(res.Vermittelt)
		vam := toString(res.Vermitteltam)

		// 2. Wenn S oder V, dann die ganze Gruppe löschen
		if (kz == "S" || kz == "V") && vam != "" {
			// Zuerst verknüpfte Lagerbuchungen der ganzen Gruppe löschen
			_, err := conn.ExecContext(c, `
				DELETE FROM EILAGERBUCHUNG 
				WHERE ID_BUCHUNG IN (
					SELECT ID FROM BUCHUNG 
					WHERE ID_HERDEN = ? AND VERMITTELTAM = ? AND VERMITTELT IN ('S', 'V')
				)`, res.IDHerden, vam)
			if err != nil {
				log.Printf("Warnung: Lagerbuchungen der Gruppe konnten nicht gelöscht werden: %v", err)
			}

			// Dann die Buchungsgruppe löschen
			_, err = conn.ExecContext(c, "DELETE FROM BUCHUNG WHERE ID_HERDEN = ? AND VERMITTELTAM = ? AND VERMITTELT IN ('S', 'V')", res.IDHerden, vam)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Löschen der Vermittlungs-Gruppe: " + err.Error()})
				return
			}
		} else {
			// Normaler Einzelsatz
			// Zuerst verknüpfte Lagerbuchung löschen
			_, err = conn.ExecContext(c, "DELETE FROM EILAGERBUCHUNG WHERE ID_BUCHUNG = ?", idInt)
			if err != nil {
				log.Printf("Warnung: Lagerbuchung konnte nicht gelöscht werden: %v", err)
			}

			if err := queries.DeleteBuchung(c, idInt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Futter APIs
	r.GET("/api/futter", func(c *gin.Context) {
		res, err := queries.ListFutterBuchungen(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/futter", func(c *gin.Context) {
		var req struct {
			IDSilo         int64   `json:"ID_SILO"`
			Silonummer     int64   `json:"SILONUMMER"`
			Herdenr        int64   `json:"HERDENR"`
			IDPerson       int64   `json:"ID_PERSON"`
			Lieferdatum    string  `json:"LIEFERDATUM"`
			Liefermenge    float64 `json:"LIEFERMENGE"`
			Preisdt        float64 `json:"PREISDT"`
			Rabattproz     float64 `json:"RABATTPROZ"`
			Netto          float64 `json:"NETTO"`
			Brutto         float64 `json:"BRUTTO"`
			Mwstproz       float64 `json:"MWSTPROZ"`
			Mwstkz         string  `json:"MWSTKZ"`
			Datum          string  `json:"DATUM"`
			Zeitstempel    string  `json:"ZEITSTEMPEL"`
			Aw             int64   `json:"AW"`
			IDFuttersorten int64   `json:"ID_FUTTERSORTEN"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.CreateFutterBuchungParams{
			IDSilo:         req.IDSilo,
			Silonummer:     req.Silonummer,
			Herdenr:        req.Herdenr,
			IDPerson:       req.IDPerson,
			Lieferdatum:    req.Lieferdatum,
			Liefermenge:    req.Liefermenge,
			Preisdt:        req.Preisdt,
			Rabattproz:     req.Rabattproz,
			Netto:          req.Netto,
			Brutto:         req.Brutto,
			Mwstproz:       req.Mwstproz,
			Mwstkz:         req.Mwstkz,
			Datum:          req.Datum,
			Zeitstempel:    req.Zeitstempel,
			Aw:             req.Aw,
			IDFuttersorten: req.IDFuttersorten,
		}
		res, err := queries.CreateFutterBuchung(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/futter/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req struct {
			IDSilo         int64   `json:"ID_SILO"`
			Silonummer     int64   `json:"SILONUMMER"`
			Herdenr        int64   `json:"HERDENR"`
			IDPerson       int64   `json:"ID_PERSON"`
			Lieferdatum    string  `json:"LIEFERDATUM"`
			Liefermenge    float64 `json:"LIEFERMENGE"`
			Preisdt        float64 `json:"PREISDT"`
			Rabattproz     float64 `json:"RABATTPROZ"`
			Netto          float64 `json:"NETTO"`
			Brutto         float64 `json:"BRUTTO"`
			Mwstproz       float64 `json:"MWSTPROZ"`
			Mwstkz         string  `json:"MWSTKZ"`
			Datum          string  `json:"DATUM"`
			Zeitstempel    string  `json:"ZEITSTEMPEL"`
			Aw             int64   `json:"AW"`
			IDFuttersorten int64   `json:"ID_FUTTERSORTEN"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateFutterBuchungParams{
			ID:             idInt,
			IDSilo:         req.IDSilo,
			Silonummer:     req.Silonummer,
			Herdenr:        req.Herdenr,
			IDPerson:       req.IDPerson,
			Lieferdatum:    req.Lieferdatum,
			Liefermenge:    req.Liefermenge,
			Preisdt:        req.Preisdt,
			Rabattproz:     req.Rabattproz,
			Netto:          req.Netto,
			Brutto:         req.Brutto,
			Mwstproz:       req.Mwstproz,
			Mwstkz:         req.Mwstkz,
			Datum:          req.Datum,
			Zeitstempel:    req.Zeitstempel,
			Aw:             req.Aw,
			IDFuttersorten: req.IDFuttersorten,
		}
		res, err := queries.UpdateFutterBuchung(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/futter/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		if err := queries.DeleteFutterBuchung(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	r.GET("/api/company/params", func(c *gin.Context) {
		res, err := queries.GetGlobalFirmenparameter(c)
		if err != nil {
			if err == sql.ErrNoRows {
				// If global params don't exist, create them
				defaultParams, err := queries.CreateFirmenparameter(c, db.CreateFirmenparameterParams{
					IDHerden: -1,
					Kz:       "F",
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create default params"})
					return
				}
				c.JSON(http.StatusOK, defaultParams)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	// Tierbewegungen APIs
	r.GET("/api/tierbewegungen", func(c *gin.Context) {
		lang := c.DefaultQuery("lang", "de")
		res, err := queries.ListTierbewegungen(c, lang)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/tierbewegungen", func(c *gin.Context) {
		var req struct {
			Herdennummer   int64   `json:"HERDENNUMMER"`
			IDBuchung      int64   `json:"ID_BUCHUNG"`
			Typ            string  `json:"TYP"`
			IDTexte        int64   `json:"ID_TEXTE"`
			Bewegungsdatum string  `json:"BEWEGUNGSDATUM"`
			Bewegungen     int64   `json:"BEWEGUNGEN"`
			IDHerdenVon    int64   `json:"ID_HERDEN_VON"`
			IDHerdenNach   int64   `json:"ID_HERDEN_NACH"`
			Kosten         float64 `json:"KOSTEN"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.CreateTierbewegungParams{
			Herdennummer:   toNullInt64(req.Herdennummer),
			IDBuchung:      toNullInt64(req.IDBuchung),
			Typ:            req.Typ,
			IDTexte:        toNullInt64(req.IDTexte),
			Bewegungsdatum: toNullString(req.Bewegungsdatum),
			Bewegungen:     toNullInt64(req.Bewegungen),
			IDHerdenVon:    toNullInt64(req.IDHerdenVon),
			IDHerdenNach:   toNullInt64(req.IDHerdenNach),
			Kosten:         toNullFloat64(req.Kosten),
		}
		res, err := queries.CreateTierbewegung(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Fall 1: Zugang (Typ 'Z')
		if req.Typ == "Z" {
			// 1. Herde aktualisieren (Bestand + Kosten)
			err = queries.IncreaseHerdeStockAndCosts(c, db.IncreaseHerdeStockAndCostsParams{
				Anfangsbestand: req.Bewegungen,
				Einstallkosten: req.Kosten,
				Herdennummer:   req.Herdennummer,
			})
			if err != nil {
				log.Printf("Warning: Fall 1 Herde-Update fehlgeschlagen: %v\n", err)
			}

			// 2. Buchung aktualisieren (Bestand)
			err = queries.UpdateBuchungStock(c, db.UpdateBuchungStockParams{
				Tierbestand:   req.Bewegungen,
				Herdennummer:  req.Herdennummer,
				Buchungsdatum: req.Bewegungsdatum,
			})
			if err != nil {
				log.Printf("Warning: Fall 1 Buchung-Update fehlgeschlagen: %v\n", err)
			}
		}

		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/tierbewegungen/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req struct {
			Herdennummer   int64   `json:"HERDENNUMMER"`
			IDBuchung      int64   `json:"ID_BUCHUNG"`
			Typ            string  `json:"TYP"`
			IDTexte        int64   `json:"ID_TEXTE"`
			Bewegungsdatum string  `json:"BEWEGUNGSDATUM"`
			Bewegungen     int64   `json:"BEWEGUNGEN"`
			IDHerdenVon    int64   `json:"ID_HERDEN_VON"`
			IDHerdenNach   int64   `json:"ID_HERDEN_NACH"`
			Kosten         float64 `json:"KOSTEN"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateTierbewegungParams{
			ID:             idInt,
			Herdennummer:   toNullInt64(req.Herdennummer),
			IDBuchung:      toNullInt64(req.IDBuchung),
			Typ:            req.Typ,
			IDTexte:        toNullInt64(req.IDTexte),
			Bewegungsdatum: toNullString(req.Bewegungsdatum),
			Bewegungen:     toNullInt64(req.Bewegungen),
			IDHerdenVon:    toNullInt64(req.IDHerdenVon),
			IDHerdenNach:   toNullInt64(req.IDHerdenNach),
			Kosten:         toNullFloat64(req.Kosten),
		}
		res, err := queries.UpdateTierbewegung(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/buchungen/detailed", func(c *gin.Context) {
		res, err := queries.ListBuchungenWithHerde(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/tierbewegungen/umbuchung", func(c *gin.Context) {
		var req struct {
			ID_BUCHUNG_VON  int64  `json:"ID_BUCHUNG_VON"`
			ID_HERDEN_VON   int64  `json:"ID_HERDEN_VON"`
			ID_BUCHUNG_NACH int64  `json:"ID_BUCHUNG_NACH"`
			ID_HERDEN_NACH  int64  `json:"ID_HERDEN_NACH"`
			MENGE           int64  `json:"MENGE"`
			DATUM           string `json:"DATUM"`
			ID_TEXTE        int64  `json:"ID_TEXTE"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Transaktion starten
		tx, err := conn.BeginTx(c, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback()

		qtx := db.New(tx)

		// 1. Bestände in der Tabelle HERDEN (Stammdaten) anpassen
		log.Printf("[UMB] Umbuchung von ID %d nach ID %d, Menge %d\n", req.ID_HERDEN_VON, req.ID_HERDEN_NACH, req.MENGE)

		// Abgebende Herde verringern
		err = qtx.AdjustHerdeStock(c, db.AdjustHerdeStockParams{
			Anfangsbestand: -req.MENGE,
			ID:             req.ID_HERDEN_VON,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Verringern des Bestands (Links): " + err.Error()})
			return
		}

		// Empfangende Herde erhöhen
		err = qtx.AdjustHerdeStock(c, db.AdjustHerdeStockParams{
			Anfangsbestand: req.MENGE,
			ID:             req.ID_HERDEN_NACH,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Erhöhen des Bestands (Rechts): " + err.Error()})
			return
		}

		// 2. Bestände in der BUCHUNG anpassen
		// 2a. Bestand in der BUCHUNG links verringern (Geben)
		bVon, _ := qtx.GetBuchung(c, req.ID_BUCHUNG_VON)
		log.Printf("[UMB] Buchung Links (ID %d) Vorher: %d, Änderung: %d\n", req.ID_BUCHUNG_VON, bVon.Tierbestand, -req.MENGE)

		err = qtx.UpdateBuchungStockById(c, db.UpdateBuchungStockByIdParams{
			Tierbestand: -req.MENGE,
			ID:          req.ID_BUCHUNG_VON,
		})

		// 2b. Bestand in der BUCHUNG rechts erhöhen (Nehmen)
		bNach, _ := qtx.GetBuchung(c, req.ID_BUCHUNG_NACH)
		log.Printf("[UMB] Buchung Rechts (ID %d) Vorher: %d, Änderung: %d\n", req.ID_BUCHUNG_NACH, bNach.Tierbestand, req.MENGE)

		err = qtx.UpdateBuchungStockById(c, db.UpdateBuchungStockByIdParams{
			Tierbestand: req.MENGE,
			ID:          req.ID_BUCHUNG_NACH,
		})

		// Kontrolle nachher
		bVonNeu, _ := qtx.GetBuchung(c, req.ID_BUCHUNG_VON)
		bNachNeu, _ := qtx.GetBuchung(c, req.ID_BUCHUNG_NACH)
		log.Printf("[UMB] Nachher - Links: %d, Rechts: %d\n", bVonNeu.Tierbestand, bNachNeu.Tierbestand)

		// 3. Tierbewegung erstellen
		// User: ID_HERDE_VON = ID_Herden von rechter Seite, ID_HERDEN_NACH = ID_HERDEN Linke Seite
		// Typ 'T', ID_TEXTE = 41 (Umbuchungen)
		herdeLinks, _ := qtx.GetHerde(c, req.ID_HERDEN_VON)
		_, err = qtx.CreateTierbewegung(c, db.CreateTierbewegungParams{
			Herdennummer:   toNullInt64(herdeLinks.Herdennummer),
			IDBuchung:      toNullInt64(req.ID_BUCHUNG_VON),
			Typ:            "U",
			IDTexte:        toNullInt64(req.ID_TEXTE),
			Bewegungsdatum: toNullString(req.DATUM),
			Bewegungen:     toNullInt64(req.MENGE),
			IDHerdenVon:    toNullInt64(req.ID_HERDEN_VON),  // Abgebende Herde
			IDHerdenNach:   toNullInt64(req.ID_HERDEN_NACH), // Empfangende Herde
			Kosten:         toNullFloat64(0),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Erstellen der Tierbewegung: " + err.Error()})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	r.DELETE("/api/tierbewegungen/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		if err := queries.DeleteTierbewegung(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Kostentabkopf APIs
	r.GET("/api/kostentabkopf", func(c *gin.Context) {
		res, err := queries.GetKostentabkopf(c)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/kostentabkopf", func(c *gin.Context) {
		var req struct {
			Schlachterloes   float64 `json:"SCHLACHTERLOES"`
			Proddauergeplant int64   `json:"PRODDAUERGEPLANT"`
			Gebaeudewert     float64 `json:"GEBAEUDEWERT"`
			AbschreibungG    float64 `json:"ABSCHREIBUNG_G"`
			Geraetewert      float64 `json:"GERAETEWERT"`
			AbschreibungR    float64 `json:"ABSCHREIBUNG_R"`
			Kostentag        float64 `json:"KOSTENTAG"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateKostentabkopfParams{
			Schlachterloes:   req.Schlachterloes,
			Proddauergeplant: req.Proddauergeplant,
			Gebaeudewert:     req.Gebaeudewert,
			AbschreibungG:    req.AbschreibungG,
			Geraetewert:      req.Geraetewert,
			AbschreibungR:    req.AbschreibungR,
			Kostentag:        req.Kostentag,
		}
		res, err := queries.UpdateKostentabkopf(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	// Kosten APIs
	r.GET("/api/kosten", func(c *gin.Context) {
		res, err := queries.ListKosten(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/kosten", func(c *gin.Context) {
		var req struct {
			Kostentyp     string  `json:"KOSTENTYP"`
			Bezeichnung   string  `json:"BEZEICHNUNG"`
			Kosten        float64 `json:"KOSTEN"`
			Tage          int64   `json:"TAGE"`
			Buchungsdatum string  `json:"BUCHUNGSDATUM"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.CreateKostenParams{
			Kostentyp:     req.Kostentyp,
			Bezeichnung:   req.Bezeichnung,
			Kosten:        req.Kosten,
			Tage:          req.Tage,
			Buchungsdatum: req.Buchungsdatum,
		}
		res, err := queries.CreateKosten(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/kosten/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req struct {
			Kostentyp     string  `json:"KOSTENTYP"`
			Bezeichnung   string  `json:"BEZEICHNUNG"`
			Kosten        float64 `json:"KOSTEN"`
			Tage          int64   `json:"TAGE"`
			Buchungsdatum string  `json:"BUCHUNGSDATUM"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateKostenParams{
			ID:            idInt,
			Kostentyp:     req.Kostentyp,
			Bezeichnung:   req.Bezeichnung,
			Kosten:        req.Kosten,
			Tage:          req.Tage,
			Buchungsdatum: req.Buchungsdatum,
		}
		res, err := queries.UpdateKosten(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/kosten/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		if err := queries.DeleteKosten(c, idInt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Admin APIs - Testdaten aktualisieren
	// TextTypen APIs
	r.GET("/api/texttypen", func(c *gin.Context) {
		lang := c.DefaultQuery("lang", "de")
		res, err := queries.ListTextTypen(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if lang != "de" {
			for i, item := range res {
				switch item.Kz {
				case "L":
					if lang == "it" {
						res[i].Bezeichnung = "Tipi di deposito uova"
					} else {
						res[i].Bezeichnung = "Egg Storage Types"
					}
				case "E":
					if lang == "it" {
						res[i].Bezeichnung = "Tipi di registrazione deposito uova"
					} else {
						res[i].Bezeichnung = "Egg Storage Booking Types"
					}
				case "V":
					if lang == "it" {
						res[i].Bezeichnung = "Tipi di perdite"
					} else {
						res[i].Bezeichnung = "Loss Types"
					}
				case "T":
					if lang == "it" {
						res[i].Bezeichnung = "Tipi di movimento animali"
					} else {
						res[i].Bezeichnung = "Animal Movement Types"
					}
				case "H":
					if lang == "it" {
						res[i].Bezeichnung = "Tipi di origine / partenza"
					} else {
						res[i].Bezeichnung = "Origin / Departure Types"
					}
				case "N":
					if lang == "it" {
						res[i].Bezeichnung = "Tipi di fornitori / allevatori"
					} else {
						res[i].Bezeichnung = "Supplier / Breeder Types"
					}
				case "P":
					if lang == "it" {
						res[i].Bezeichnung = "Tipi di personale"
					} else {
						res[i].Bezeichnung = "Personnel Types"
					}
				}
			}
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/texttypen", func(c *gin.Context) {
		var req struct {
			Kz          string      `json:"kz" binding:"required"`
			Bezeichnung string      `json:"bezeichnung"`
			System      interface{} `json:"system"`
			Status      interface{} `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sysVal := toInt64(req.System)
		statusVal := toInt64(req.Status)

		kzClean := wailsdb.SanitizeKZ(req.Kz)

		res, err := queries.CreateTextTyp(c, db.CreateTextTypParams{
			Kz:          kzClean,
			Bezeichnung: req.Bezeichnung,
			SystemKz:    sysVal,
			Status:      statusVal,
		})
		if err != nil {
			log.Printf("[API] POST /api/texttypen - DB Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		log.Printf("[API] POST /api/texttypen - Success, New ID: %d", res.ID)
		c.JSON(http.StatusOK, gin.H{"ID": res.ID, "KZ": res.Kz})
	})

	r.PUT("/api/texttypen/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		if appconfig.LoadConfig().System == 0 {
			items, err := queries.ListTextTypen(c)
			if err == nil {
				for _, item := range items {
					if item.ID == id && item.SystemKz == 1 {
						log.Printf("[API] PUT /api/texttypen/%d - Blocked by system=0 lock", id)
						c.JSON(http.StatusForbidden, gin.H{"error": "System-Einträge können nicht geändert werden (system=0 in settings.json)."})
						return
					}
				}
			}
		}

		var req struct {
			Kz          *string     `json:"kz"`
			Bezeichnung *string     `json:"bezeichnung"`
			System      interface{} `json:"system"`
			Status      interface{} `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[API] PUT /api/texttypen/%d - JSON binding error: %v", id, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON error: " + err.Error()})
			return
		}

		kzVal := ""
		if req.Kz != nil {
			kzVal = wailsdb.SanitizeKZ(*req.Kz)
		}
		bezVal := ""
		if req.Bezeichnung != nil {
			bezVal = *req.Bezeichnung
		}
		sysVal := toInt64(req.System)
		statusVal := toInt64(req.Status)

		log.Printf("[API] PUT /api/texttypen/%d - Mapped: KZ=%s, Bez=%s, Sys=%d", id, kzVal, bezVal, sysVal)

		updatedRes, err := queries.UpdateTextTyp(c, db.UpdateTextTypParams{
			ID:          id,
			Kz:          kzVal,
			Bezeichnung: bezVal,
			SystemKz:    sysVal,
			Status:      statusVal,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": updatedRes.ID, "status": "updated", "rows": 1})
	})

	r.DELETE("/api/texttypen/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		if appconfig.LoadConfig().System == 0 {
			items, err := queries.ListTextTypen(c)
			if err == nil {
				for _, item := range items {
					if item.ID == id && item.SystemKz == 1 {
						log.Printf("[API] DELETE /api/texttypen/%d - Blocked by system=0 lock", id)
						c.JSON(http.StatusForbidden, gin.H{"error": "System-Einträge können nicht gelöscht werden (system=0 in settings.json)."})
						return
					}
				}
			}
		}

		if err := queries.DeleteTextTyp(c, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Texte APIs
	r.GET("/api/texte", func(c *gin.Context) {
		lang := c.DefaultQuery("lang", "de")
		res, err := queries.ListTexte(c, lang)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/texte/typ/:typ", func(c *gin.Context) {
		lang := c.DefaultQuery("lang", "de")
		res, err := queries.ListTexteByType(c, db.ListTexteByTypeParams{
			SpracheKz: lang,
			TextTypKz: c.Param("typ"),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/texte/verwendung", func(c *gin.Context) {
		lang := c.DefaultQuery("lang", "de")
		res, err := queries.ListVerwendungsTexte(c, lang)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/texte/kz/:kz", func(c *gin.Context) {
		lang := c.DefaultQuery("lang", "de")
		res, err := queries.ListTexteByKZ(c, db.ListTexteByKZParams{
			SpracheKz: lang,
			Kz:        c.Param("kz"),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/texte", func(c *gin.Context) {
		var req struct {
			TextTypKz string      `json:"text_typ_kz" binding:"required"`
			Kz        string      `json:"kz"`
			Betreff   string      `json:"betreff"`
			Inhalt    string      `json:"inhalt"`
			System    interface{} `json:"system"`
			Status    interface{} `json:"status"`
		}
		lang := c.DefaultQuery("lang", "de")
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON error: " + err.Error()})
			return
		}

		sysVal := toInt64(req.System)
		statusVal := toInt64(req.Status)

		tx, err := conn.BeginTx(c, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback()

		// 1. Create text entry
		kzClean := wailsdb.SanitizeKZ(req.Kz)

		var qtx db.Querier
		if database.Engine == "mysql" {
			qtx = queries.(*wailsdb.MySQLWrapper).WithTx(tx)
		} else {
			qtx = queries.(*db.Queries).WithTx(tx)
		}

		res, err := qtx.CreateText(c, db.CreateTextParams{
			TextTypKz: req.TextTypKz,
			Kz:        kzClean,
			SystemKz:  sysVal,
			Status:    statusVal,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Create text failed: " + err.Error()})
			return
		}

		// 2. Create translation
		_, err = qtx.CreateUebersetzung(c, db.CreateUebersetzungParams{
			IDTexte:   res.ID,
			SpracheKz: lang,
			Betreff:   req.Betreff,
			Inhalt:    req.Inhalt,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Create translation failed: " + err.Error()})
			return
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"ID": res.ID})
	})

	r.PUT("/api/texte/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		if appconfig.LoadConfig().System == 0 {
			items, err := queries.ListTexte(c, "de")
			if err == nil {
				for _, item := range items {
					if item.ID == id && item.SystemKz == 1 {
						log.Printf("[API] PUT /api/texte/%d - Blocked by system=0 lock", id)
						c.JSON(http.StatusForbidden, gin.H{"error": "System-Einträge können nicht geändert werden (system=0 in settings.json)."})
						return
					}
				}
			}
		}

		lang := c.DefaultQuery("lang", "de")
		var req struct {
			TextTypKz string      `json:"text_typ_kz"`
			Kz        string      `json:"kz"`
			Betreff   string      `json:"betreff"`
			Inhalt    string      `json:"inhalt"`
			System    interface{} `json:"system"`
			Status    interface{} `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[API] PUT /api/texte/%d - JSON binding error: %v", id, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON error: " + err.Error()})
			return
		}

		sysVal := toInt64(req.System)
		statusVal := toInt64(req.Status)

		log.Printf("[API] PUT /api/texte/%d - Mapped Data: %+v, Final SystemKz: %d", id, req, sysVal)

		tx, err := conn.BeginTx(c, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback()

		var qtx db.Querier
		if database.Engine == "mysql" {
			qtx = queries.(*wailsdb.MySQLWrapper).WithTx(tx)
		} else {
			qtx = queries.(*db.Queries).WithTx(tx)
		}

		// 1. Update text entry
		kzClean := wailsdb.SanitizeKZ(req.Kz)
		_, err = qtx.UpdateText(c, db.UpdateTextParams{
			ID:        id,
			TextTypKz: req.TextTypKz,
			Kz:        kzClean,
			SystemKz:  sysVal,
			Status:    statusVal,
		})
		if err != nil {
			log.Printf("[API] PUT /api/texte/%d - DB Error: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update text failed: " + err.Error()})
			return
		}

		// 2. Upsert translation
		_, err = qtx.UpsertUebersetzung(c, db.UpsertUebersetzungParams{
			IDTexte:   id,
			SpracheKz: lang,
			Betreff:   req.Betreff,
			Inhalt:    req.Inhalt,
		})
		if err != nil {
			log.Printf("[API] PUT /api/texte/%d - Upsert translation Error: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upsert translation failed: " + err.Error()})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id, "status": "updated"})
	})

	r.DELETE("/api/texte/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		if appconfig.LoadConfig().System == 0 {
			items, err := queries.ListTexte(c, "de")
			if err == nil {
				for _, item := range items {
					if item.ID == id && item.SystemKz == 1 {
						log.Printf("[API] DELETE /api/texte/%d - Blocked by system=0 lock", id)
						c.JSON(http.StatusForbidden, gin.H{"error": "System-Einträge können nicht gelöscht werden (system=0 in settings.json)."})
						return
					}
				}
			}
		}

		if err := queries.DeleteText(c, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Backup Endpunkt
	// Backup Endpunkt mit Zeitstempel
	r.POST("/api/backup", func(c *gin.Context) {
		if database.Engine == "mysql" {
			c.JSON(http.StatusOK, gin.H{
				"status":  "info",
				"message": "MariaDB Backups können nicht über diese Anwendung erstellt werden. Bitte führen Sie die Sicherung über Ihre externe Datenbank-Administration (z.B. mysqldump, phpMyAdmin, HeidiSQL) durch.",
			})
			return
		}

		timestamp := time.Now().Format("0601021504")
		destPath := filepath.Join(backupDir, fmt.Sprintf("HuhnLite%s.db", timestamp))

		err := copyFile(currentDBPath, destPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Backup fehlgeschlagen: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "filename": filepath.Base(destPath), "path": destPath})
	})

	// DB Management Endpunkte
	r.GET("/api/db/list", func(c *gin.Context) {
		var files []string
		dirs := []string{filepath.Dir(currentDBPath), backupDir}

		seen := make(map[string]bool)
		for _, dir := range dirs {
			matches, _ := filepath.Glob(filepath.Join(dir, "*.db"))
			for _, m := range matches {
				p := filepath.ToSlash(m)
				if !seen[p] {
					files = append(files, p)
					seen[p] = true
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"files":   files,
			"current": currentDBPath,
		})
	})

	r.POST("/api/db/switch", func(c *gin.Context) {
		var req struct {
			Path string `json:"PATH"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// 1. Sicherheits-Backup der aktuellen DB
		timestamp := time.Now().Format("0601021504")
		safetyBackup := filepath.Join(backupDir, fmt.Sprintf("SafetyBackup_before_switch_%s.db", timestamp))

		err := copyFile(currentDBPath, safetyBackup)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Sicherheits-Backup fehlgeschlagen: " + err.Error()})
			return
		}

		// 2. Verbindung wechseln
		newConn, err := sql.Open("sqlite", req.Path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Neue DB konnte nicht geöffnet werden: " + err.Error()})
			return
		}

		// Teste Verbindung
		if err = newConn.Ping(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Neue DB antwortet nicht: " + err.Error()})
			return
		}

		// Schließe alte Verbindung
		if conn != nil {
			conn.Close()
		}

		conn = newConn
		queries = db.New(conn)
		currentDBPath = req.Path

		c.JSON(http.StatusOK, gin.H{
			"status":        "switched",
			"current":       currentDBPath,
			"safety_backup": filepath.Base(safetyBackup),
		})
	})

	r.POST("/api/db/restore", func(c *gin.Context) {
		if database.Engine == "mysql" {
			c.JSON(http.StatusOK, gin.H{
				"status":  "info",
				"message": "Ein Restore von MariaDB-Daten muss extern über die Datenbank-Administration durchgeführt werden. Diese Funktion ist nur für die lokale SQLite-Datenbank verfügbar.",
			})
			return
		}

		var req struct {
			Path string `json:"PATH"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// 1. Sicherheits-Backup der aktuellen DB
		timestamp := time.Now().Format("0601021504")
		safetyBackup := filepath.Join(backupDir, fmt.Sprintf("SafetyBeforeRestore_%s.db", timestamp))

		dbMutex.Lock()
		defer dbMutex.Unlock()

		err := copyFile(currentDBPath, safetyBackup)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Sicherheits-Backup vor Restore fehlgeschlagen: " + err.Error()})
			return
		}

		// 2. Verbindung schließen
		if conn != nil {
			conn.Close()
		}

		// 3. Datei überschreiben
		err = copyFile(req.Path, currentDBPath)
		if err != nil {
			// Versuche alte Verbindung wieder zu öffnen
			conn, _ = sql.Open("sqlite", currentDBPath)
			queries = db.New(conn)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Datei-Wiederherstellung fehlgeschlagen: " + err.Error()})
			return
		}

		// 4. Verbindung neu öffnen
		newConn, err := sql.Open("sqlite", currentDBPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Verbindung nach Restore fehlgeschlagen: " + err.Error()})
			return
		}

		conn = newConn
		queries = db.New(conn)

		c.JSON(http.StatusOK, gin.H{
			"status":  "restored",
			"current": currentDBPath,
			"safety":  filepath.Base(safetyBackup),
		})
	})

	// System-Reconnect via absolutem Pfad
	r.POST("/api/system/reconnect", func(c *gin.Context) {
		var req struct {
			Path string `json:"PATH"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// 1. Datei-Check
		info, err := os.Stat(req.Path)
		if os.IsNotExist(err) || info.IsDir() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Datei existiert nicht oder ist ein Verzeichnis"})
			return
		}

		// 2. Validiere SQLite Header (minimal)
		f, err := os.Open(req.Path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Datei konnte nicht geöffnet werden: " + err.Error()})
			return
		}
		header := make([]byte, 16)
		_, err = f.Read(header)
		f.Close()
		if err != nil || string(header) != "SQLite format 3\x00" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Keine valide SQLite Datenbank"})
			return
		}

		// 3. Sicherheits-Backup der aktuellen DB
		timestamp := time.Now().Format("0601021504")
		safetyBackup := filepath.Join(backupDir, fmt.Sprintf("SafetyBackup_before_manual_switch_%s.db", timestamp))

		err = copyFile(currentDBPath, safetyBackup)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Sicherheits-Backup fehlgeschlagen: " + err.Error()})
			return
		}

		// 4. Verbindung wechseln
		newConn, err := sql.Open("sqlite", req.Path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Neue DB konnte nicht geöffnet werden: " + err.Error()})
			return
		}

		if err = newConn.Ping(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Neue DB antwortet nicht: " + err.Error()})
			return
		}

		// Schließe alte Verbindung
		if conn != nil {
			conn.Close()
		}

		conn = newConn
		queries = db.New(conn)
		currentDBPath = req.Path

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"current": currentDBPath,
			"safety":  filepath.Base(safetyBackup),
		})
	})

	r.GET("/api/solltabellen", func(c *gin.Context) {
		res, err := queries.ListSolltabellen(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/tabellenkopf/typ/:typ", func(c *gin.Context) {
		typ := c.Param("typ")
		res, err := queries.ListTabellenkopfByType(c, typ)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		log.Printf("[API] ListTabellenkopfByType: type=%s, found=%d", typ, len(res))
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/tabellenkopf/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req struct {
			Tabellennummer int64  `json:"TABELLENNUMMER"`
			Bezeichnung    string `json:"BEZEICHNUNG"`
			Anlagedatum    string `json:"ANLAGEDATUM"`
			Datum          string `json:"DATUM"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.UpdateTabellenkopfParams{
			ID:             idInt,
			Tabellennummer: req.Tabellennummer,
			Bezeichnung:    req.Bezeichnung,
			Anlagedatum:    req.Anlagedatum,
			Datum:          req.Datum,
		}
		res, err := queries.UpdateTabellenkopf(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/tabellenkopf", func(c *gin.Context) {
		var req struct {
			Tabellentyp    string `json:"TABELLENTYP"`
			Tabellennummer int64  `json:"TABELLENNUMMER"`
			Bezeichnung    string `json:"BEZEICHNUNG"`
			Anlagedatum    string `json:"ANLAGEDATUM"`
			Datum          string `json:"DATUM"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := db.CreateTabellenkopfParams{
			Tabellentyp:    req.Tabellentyp,
			Tabellennummer: req.Tabellennummer,
			Bezeichnung:    req.Bezeichnung,
			Anlagedatum:    req.Anlagedatum,
			Datum:          req.Datum,
		}
		res, err := queries.CreateTabellenkopf(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/tabellenkopf/:id", func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		// 1. Get the Tabellenkopf to know Typ and Nummer
		var ttyp interface{}
		var tnum int64
		err := conn.QueryRowContext(c, "SELECT TABELLENTYP, TABELLENNUMMER FROM TABELLENKOPF WHERE ID = ?", id).Scan(&ttyp, &tnum)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tabelle nicht gefunden"})
			return
		}

		typStr := fmt.Sprintf("%s", ttyp)

		// 2. Delete rows in associated tables
		if typStr == "G" {
			_, _ = conn.ExecContext(c, "DELETE FROM GEWICHTTABELLE WHERE TABELLENNUMMER = ?", tnum)
		} else if typStr == "L" {
			_, _ = conn.ExecContext(c, "DELETE FROM LSLKLASSIK WHERE TABELLENNUMMER = ?", tnum)
		}

		// 3. Delete the header
		_, err = conn.ExecContext(c, "DELETE FROM TABELLENKOPF WHERE ID = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	r.GET("/api/lsl_klassik/tabnum/:tabnum", func(c *gin.Context) {
		tabnum, _ := strconv.ParseInt(c.Param("tabnum"), 10, 64)
		res, err := queries.ListLSLKlassikByTabNum(c, tabnum)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/lsl_klassik", func(c *gin.Context) {
		var req struct {
			Tabellennummer int64   `json:"TABELLENNUMMER"`
			Alterinwochen  int64   `json:"ALTERINWOCHEN"`
			Eizahlkum      float64 `json:"EIZAHLKUM"`
			Legerateah     float64 `json:"LEGERATEAH"`
			Legeratedh     float64 `json:"LEGERATEDH"`
			Eigewichtwo    float64 `json:"EIGEWICHTWO"`
			Eigewichtkum   float64 `json:"EIGEWICHTKUM"`
			Eimassewo      float64 `json:"EIMASSEWO"`
			Eimassekum     float64 `json:"EIMASSEKUM"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := queries.CreateLSLKlassik(c, db.CreateLSLKlassikParams{
			Tabellennummer: req.Tabellennummer,
			Alterinwochen:  req.Alterinwochen,
			Eizahlkum:      req.Eizahlkum,
			Legerateah:     req.Legerateah,
			Legeratedh:     req.Legeratedh,
			Eigewichtwo:    req.Eigewichtwo,
			Eigewichtkum:   req.Eigewichtkum,
			Eimassewo:      req.Eimassewo,
			Eimassekum:     req.Eimassekum,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/lsl_klassik/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige ID: " + idStr})
			return
		}
		var req struct {
			Tabellennummer int64   `json:"TABELLENNUMMER"`
			Alterinwochen  int64   `json:"ALTERINWOCHEN"`
			Eizahlkum      float64 `json:"EIZAHLKUM"`
			Legerateah     float64 `json:"LEGERATEAH"`
			Legeratedh     float64 `json:"LEGERATEDH"`
			Eigewichtwo    float64 `json:"EIGEWICHTWO"`
			Eigewichtkum   float64 `json:"EIGEWICHTKUM"`
			Eimassewo      float64 `json:"EIMASSEWO"`
			Eimassekum     float64 `json:"EIMASSEKUM"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := queries.UpdateLSLKlassik(c, db.UpdateLSLKlassikParams{
			ID:             id,
			Tabellennummer: req.Tabellennummer,
			Alterinwochen:  req.Alterinwochen,
			Eizahlkum:      req.Eizahlkum,
			Legerateah:     req.Legerateah,
			Legeratedh:     req.Legeratedh,
			Eigewichtwo:    req.Eigewichtwo,
			Eigewichtkum:   req.Eigewichtkum,
			Eimassewo:      req.Eimassewo,
			Eimassekum:     req.Eimassekum,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/lsl_klassik/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige ID: " + idStr})
			return
		}
		if err := queries.DeleteLSLKlassik(c, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	r.GET("/api/gewichttabelle/tabnum/:tabnum", func(c *gin.Context) {
		tabnum, _ := strconv.ParseInt(c.Param("tabnum"), 10, 64)
		res, err := queries.ListGewichtByTabNum(c, tabnum)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/gewichttabelle", func(c *gin.Context) {
		var req struct {
			Tabellennummer int64   `json:"TABELLENNUMMER"`
			Eigewicht      float64 `json:"EIGEWICHT"`
			Klasse1        float64 `json:"KLASSE1"`
			Klasse2        float64 `json:"KLASSE2"`
			Klasse3        float64 `json:"KLASSE3"`
			Klasse4        float64 `json:"KLASSE4"`
			Klasse5        float64 `json:"KLASSE5"`
			Klasse6        float64 `json:"KLASSE6"`
			Klasse7        float64 `json:"KLASSE7"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := queries.CreateGewichtTabelle(c, db.CreateGewichtTabelleParams{
			Tabellennummer: req.Tabellennummer,
			Eigewicht:      req.Eigewicht,
			Klasse1:        req.Klasse1,
			Klasse2:        req.Klasse2,
			Klasse3:        req.Klasse3,
			Klasse4:        req.Klasse4,
			Klasse5:        req.Klasse5,
			Klasse6:        req.Klasse6,
			Klasse7:        req.Klasse7,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/gewichttabelle/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige ID: " + idStr})
			return
		}
		var req struct {
			Tabellennummer int64   `json:"TABELLENNUMMER"`
			Eigewicht      float64 `json:"EIGEWICHT"`
			Klasse1        float64 `json:"KLASSE1"`
			Klasse2        float64 `json:"KLASSE2"`
			Klasse3        float64 `json:"KLASSE3"`
			Klasse4        float64 `json:"KLASSE4"`
			Klasse5        float64 `json:"KLASSE5"`
			Klasse6        float64 `json:"KLASSE6"`
			Klasse7        float64 `json:"KLASSE7"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := queries.UpdateGewichtTabelle(c, db.UpdateGewichtTabelleParams{
			ID:             id,
			Tabellennummer: req.Tabellennummer,
			Eigewicht:      req.Eigewicht,
			Klasse1:        req.Klasse1,
			Klasse2:        req.Klasse2,
			Klasse3:        req.Klasse3,
			Klasse4:        req.Klasse4,
			Klasse5:        req.Klasse5,
			Klasse6:        req.Klasse6,
			Klasse7:        req.Klasse7,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/gewichttabelle/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige ID: " + idStr})
			return
		}
		if err := queries.DeleteGewichtTabelle(c, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	r.GET("/api/tabellen/gewicht/:tabnr/:weight", func(c *gin.Context) {
		tabNr, _ := strconv.ParseInt(c.Param("tabNr"), 10, 64)
		weight, _ := strconv.ParseFloat(c.Param("weight"), 64)
		res, err := queries.GetGewichtByTabNumAndWeight(c, db.GetGewichtByTabNumAndWeightParams{
			Tabellennummer: tabNr,
			ROUND:          weight,
		})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Keine Daten gefunden"})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/tabellen/lsl/:tabnr/:age", func(c *gin.Context) {
		tabNr, _ := strconv.ParseInt(c.Param("tabNr"), 10, 64)
		age, _ := strconv.ParseInt(c.Param("age"), 10, 64)
		res, err := queries.GetLSLByTabNumAndAge(c, db.GetLSLByTabNumAndAgeParams{
			Tabellennummer: tabNr,
			Alterinwochen:  age,
		})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Keine Daten gefunden"})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	// Preistabelle (Eierpreise)
	r.GET("/api/preise", func(c *gin.Context) {
		res, err := queries.ListEierpreise(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/preise", func(c *gin.Context) {
		var req struct {
			KzHaltungstyp string  `json:"KZ_HALTUNGSTYP"`
			Eierklasse    string  `json:"EIERKLASSE"`
			GewichtVon    float64 `json:"GEWICHT_VON"`
			GewichtBis    float64 `json:"GEWICHT_BIS"`
			PreisVon      float64 `json:"PREIS_VON"`
			PreisBis      float64 `json:"PREIS_BIS"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := queries.CreateEierpreis(c, db.CreateEierpreisParams{
			KzHaltungstyp: req.KzHaltungstyp,
			Eierklasse:    req.Eierklasse,
			GewichtVon:    req.GewichtVon,
			GewichtBis:    req.GewichtBis,
			PreisVon:      req.PreisVon,
			PreisBis:      req.PreisBis,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/preise/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige ID: " + idStr})
			return
		}
		var req struct {
			KzHaltungstyp string  `json:"KZ_HALTUNGSTYP"`
			Eierklasse    string  `json:"EIERKLASSE"`
			GewichtVon    float64 `json:"GEWICHT_VON"`
			GewichtBis    float64 `json:"GEWICHT_BIS"`
			PreisVon      float64 `json:"PREIS_VON"`
			PreisBis      float64 `json:"PREIS_BIS"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := queries.UpdateEierpreis(c, db.UpdateEierpreisParams{
			ID:            id,
			KzHaltungstyp: req.KzHaltungstyp,
			Eierklasse:    req.Eierklasse,
			GewichtVon:    req.GewichtVon,
			GewichtBis:    req.GewichtBis,
			PreisVon:      req.PreisVon,
			PreisBis:      req.PreisBis,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/preise/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige ID: " + idStr})
			return
		}
		if err := queries.DeleteEierpreis(c, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// --- FELD-KONFIGURATION ---
	r.GET("/api/field-configs", func(c *gin.Context) {
		res, err := queries.ListFieldTranslations(c, "de")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/field-configs/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige ID: " + idStr})
			return
		}
		var req struct {
			Inhalt   string `json:"INHALT"`
			Betreff  string `json:"BETREFF"`
			Nameindb int64  `json:"NAMEINDB"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Edit catalog entry (NAMEINDB)
		_, _ = conn.Exec("UPDATE FELD_KATALOG SET NAMEINDB = ? WHERE ID = ?", req.Nameindb, id)

		// Edit translation
		res, err := queries.UpsertFieldTranslation(c, db.UpsertFieldTranslationParams{
			IDFeldKatalog: id,
			SpracheKz:     "de",
			Betreff:       req.Betreff,
			Inhalt:        req.Inhalt,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/field-configs/catalog", func(c *gin.Context) {
		var req struct {
			Kz       string `json:"KZ"`
			Feldname string `json:"FELDNAME"`
			Nameindb int64  `json:"NAMEINDB"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := queries.CreateFieldKatalog(c, db.CreateFieldKatalogParams{
			Kz:       req.Kz,
			Feldname: req.Feldname,
			Nameindb: req.Nameindb,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/field-configs/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige ID: " + idStr})
			return
		}
		_, err = conn.Exec("DELETE FROM FELD_KATALOG WHERE ID = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})
	r.POST("/api/field-configs/sync", func(c *gin.Context) {
		// 1. Alle Tabellennamen abfragen (System-Tabellen ausschliessen)
		var tableRows *sql.Rows
		var err error
		if database.Engine == "mysql" {
			tableRows, err = conn.Query("SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE()")
		} else {
			tableRows, err = conn.Query("SELECT NAME FROM sqlite_master WHERE type='table' AND NAME NOT LIKE 'sqlite_%'")
		}

		if err != nil {
			log.Printf("Sync-Fehler bei Tabellenabfrage: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Lesen der Tabellenstruktur: " + err.Error()})
			return
		}
		defer tableRows.Close()

		var tableNames []string
		for tableRows.Next() {
			var name string
			if err := tableRows.Scan(&name); err == nil {
				tableNames = append(tableNames, name)
			}
		}

		// 2. Alle Spaltennamen sammeln
		fieldNames := make(map[string]bool)
		for _, tableName := range tableNames {
			var cols *sql.Rows
			var err error
			if database.Engine == "mysql" {
				cols, err = conn.Query("SELECT 0 as cid, COLUMN_NAME as name, DATA_TYPE as dtype, IF(IS_NULLABLE='NO', 1, 0) as notnull, COLUMN_DEFAULT as dfltValue, IF(COLUMN_KEY='PRI', 1, 0) as pk FROM information_schema.COLUMNS WHERE TABLE_NAME = ? AND TABLE_SCHEMA = DATABASE()", tableName)
			} else {
				cols, err = conn.Query(fmt.Sprintf("PRAGMA table_info('%s')", tableName))
			}
			if err != nil {
				continue
			}
			for cols.Next() {
				var cid int
				var name, dtype string
				var notnull, pk int
				var dfltValue interface{}
				if err := cols.Scan(&cid, &name, &dtype, &notnull, &dfltValue, &pk); err == nil {
					fieldNames[strings.ToUpper(name)] = true
				}
			}
			cols.Close()
		}

		// 3. Neue Felder in einer Transaktion hinzufügen
		tx, err := conn.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaktionsfehler: " + err.Error()})
			return
		}
		defer tx.Rollback()

		newCount := 0
		for fieldName := range fieldNames {
			// Bereinigen
			safeName := strings.TrimSpace(fieldName)

			// Leere Namen ignorieren (verhindert ID 782 Problem)
			if safeName == "" {
				continue
			}

			// FELDNAME ist auf 25 Zeichen begrenzt - abschneiden falls nötig
			if len(safeName) > 25 {
				safeName = safeName[:25]
			}

			// In den Katalog einfügen (IGNORE falls schon da)
			res, err := tx.Exec("INSERT OR IGNORE INTO FELD_KATALOG (FELDNAME, KZ, NAMEINDB) VALUES (?, 'X', 1)", safeName)
			if err != nil {
				continue
			}

			rowsAffected, _ := res.RowsAffected()
			if rowsAffected > 0 {
				newCount++
				// ID holen
				var lastID int64
				err = tx.QueryRow("SELECT ID FROM FELD_KATALOG WHERE FELDNAME = ?", safeName).Scan(&lastID)
				if err == nil {
					// Initiale Übersetzung 'de'
					tx.Exec("INSERT OR IGNORE INTO TRANSLATEFELDNAMEN (id_feld_katalog, sprache_kz, BETREFF, INHALT) VALUES (?, 'de', ?, ?)",
						lastID, safeName, safeName)
				}
			}
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit-Fehler: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   fmt.Sprintf("Synchronisation abgeschlossen. %d neue Feldnamen wurden hinzugefügt.", newCount),
			"new_count": newCount,
		})
	})

	// --- DYNAMISCHE REPORTS ---
	r.GET("/api/reports", func(c *gin.Context) {
		res, err := queries.ListDynamischeSQL(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Laden der Berichtsliste: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/reports/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Report-ID: " + idStr})
			return
		}
		res, err := queries.GetDynamischeSQL(c, id)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Report nicht gefunden"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Datenbankfehler: " + err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/reports", func(c *gin.Context) {
		var req struct {
			Beschreibung       string      `json:"BESCHREIBUNG"`
			Sqlstatement       string      `json:"SQLSTATEMENT"`
			KategorieKz        string      `json:"KATEGORIE_KZ"`
			GruppenKz          string      `json:"GRUPPEN_KZ"`
			TypKz              string      `json:"TYP_KZ"`
			TemplateName       string      `json:"TEMPLATE_NAME"`
			ParamDef           string      `json:"PARAM_DEF"`
			DetailSql          string      `json:"DETAIL_SQL"`
			LinkLogic          string      `json:"LINK_LOGIC"`
			GroupField         string      `json:"GROUP_FIELD"`
			RowsPerPage        int64       `json:"ROWS_PER_PAGE"`
			PageOrientation    string      `json:"PAGE_ORIENTATION"`
			ShowMasterGrid     int64       `json:"SHOW_MASTER_GRID"`
			ShowDetailGrid     int64       `json:"SHOW_DETAIL_GRID"`
			SystemKz           interface{} `json:"SYSTEM_KZ"`
			SqlstatementNative string      `json:"SQLSTATEMENT_NATIVE"`
			DetailSqlNative    string      `json:"DETAIL_SQL_NATIVE"`
			RootKz             string      `json:"ROOT_KZ"`
			Summenzeile        interface{} `json:"SUMMENZEILE"`
			IstSummenzeile     interface{} `json:"IST_SUMMENZEILE"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Feldkatalog synchronisieren und SQL normalisieren (Aliase auf Großschreibung "AS")
		req.SqlstatementNative = syncFieldCatalog(conn, req.SqlstatementNative)
		req.DetailSqlNative = syncFieldCatalog(conn, req.DetailSqlNative)

		res, err := queries.CreateDynamischeSQL(c, db.CreateDynamischeSQLParams{
			Beschreibung:       req.Beschreibung,
			Sqlstatement:       req.Sqlstatement,
			KategorieKz:        req.KategorieKz,
			GruppenKz:          req.GruppenKz,
			TypKz:              req.TypKz,
			TemplateName:       req.TemplateName,
			ParamDef:           req.ParamDef,
			DetailSql:          req.DetailSql,
			LinkLogic:          req.LinkLogic,
			GroupField:         req.GroupField,
			RowsPerPage:        req.RowsPerPage,
			PageOrientation:    req.PageOrientation,
			ShowMasterGrid:     req.ShowMasterGrid,
			ShowDetailGrid:     req.ShowDetailGrid,
			SystemKz:           toString(req.SystemKz),
			SqlstatementNative: req.SqlstatementNative,
			DetailSqlNative:    req.DetailSqlNative,
			RootKz:             req.RootKz,
			Summenzeile:        toString(req.Summenzeile),
			IstSummenzeile:     toInt64(req.IstSummenzeile),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/reports/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Report-ID: " + idStr})
			return
		}

		if appconfig.LoadConfig().System == 0 {
			item, err := queries.GetDynamischeSQL(c, id)
			if err == nil {
				sysStr := strings.TrimSpace(item.SystemKz)
				if sysStr != "" && sysStr != "0" && strings.ToLower(sysStr) != "false" {
					log.Printf("[API] PUT /api/reports/%d - Blocked by system=0 lock", id)
					c.JSON(http.StatusForbidden, gin.H{"error": "System-Einträge können nicht geändert werden (system=0 in settings.json)."})
					return
				}
			}
		}

		var req struct {
			Beschreibung       string      `json:"BESCHREIBUNG"`
			Sqlstatement       string      `json:"SQLSTATEMENT"`
			KategorieKz        string      `json:"KATEGORIE_KZ"`
			GruppenKz          string      `json:"GRUPPEN_KZ"`
			TypKz              string      `json:"TYP_KZ"`
			TemplateName       string      `json:"TEMPLATE_NAME"`
			ParamDef           string      `json:"PARAM_DEF"`
			DetailSql          string      `json:"DETAIL_SQL"`
			LinkLogic          string      `json:"LINK_LOGIC"`
			GroupField         string      `json:"GROUP_FIELD"`
			RowsPerPage        int64       `json:"ROWS_PER_PAGE"`
			PageOrientation    string      `json:"PAGE_ORIENTATION"`
			ShowMasterGrid     int64       `json:"SHOW_MASTER_GRID"`
			ShowDetailGrid     int64       `json:"SHOW_DETAIL_GRID"`
			SystemKz           interface{} `json:"SYSTEM_KZ"`
			SqlstatementNative string      `json:"SQLSTATEMENT_NATIVE"`
			DetailSqlNative    string      `json:"DETAIL_SQL_NATIVE"`
			RootKz             string      `json:"ROOT_KZ"`
			Summenzeile        interface{} `json:"SUMMENZEILE"`
			IstSummenzeile     interface{} `json:"IST_SUMMENZEILE"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Feldkatalog synchronisieren und SQL normalisieren (Aliase auf Großschreibung "AS")
		req.SqlstatementNative = syncFieldCatalog(conn, req.SqlstatementNative)
		req.DetailSqlNative = syncFieldCatalog(conn, req.DetailSqlNative)

		res, err := queries.UpdateDynamischeSQL(c, db.UpdateDynamischeSQLParams{
			ID:                 id,
			Beschreibung:       req.Beschreibung,
			Sqlstatement:       req.Sqlstatement,
			KategorieKz:        req.KategorieKz,
			GruppenKz:          req.GruppenKz,
			TypKz:              req.TypKz,
			TemplateName:       req.TemplateName,
			ParamDef:           req.ParamDef,
			DetailSql:          req.DetailSql,
			LinkLogic:          req.LinkLogic,
			GroupField:         req.GroupField,
			RowsPerPage:        req.RowsPerPage,
			PageOrientation:    req.PageOrientation,
			ShowMasterGrid:     req.ShowMasterGrid,
			ShowDetailGrid:     req.ShowDetailGrid,
			SystemKz:           toString(req.SystemKz),
			SqlstatementNative: req.SqlstatementNative,
			DetailSqlNative:    req.DetailSqlNative,
			RootKz:             req.RootKz,
			Summenzeile:        toString(req.Summenzeile),
			IstSummenzeile:     toInt64(req.IstSummenzeile),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/reports/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Report-ID: " + idStr})
			return
		}

		if appconfig.LoadConfig().System == 0 {
			item, err := queries.GetDynamischeSQL(c, id)
			if err == nil {
				sysStr := strings.TrimSpace(item.SystemKz)
				if sysStr != "" && sysStr != "0" && strings.ToLower(sysStr) != "false" {
					log.Printf("[API] DELETE /api/reports/%d - Blocked by system=0 lock", id)
					c.JSON(http.StatusForbidden, gin.H{"error": "System-Einträge können nicht gelöscht werden (system=0 in settings.json)."})
					return
				}
			}
		}

		if err := queries.DeleteDynamischeSQL(c, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	r.POST("/api/reports/test-sql", func(c *gin.Context) {
		var req struct {
			MasterSql string                 `json:"MASTER_SQL"`
			DetailSql string                 `json:"DETAIL_SQL"`
			Params    map[string]interface{} `json:"PARAMS"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Anfrage"})
			return
		}

		lang := c.DefaultQuery("lang", "de")
		if req.Params == nil {
			req.Params = make(map[string]interface{})
		}
		// System-Parameter injizieren (für SQL-Platzhalter)
		req.Params["LANG"] = lang
		req.Params["SPR_KZ"] = lang

		translations, _ := queries.ListTranslateFeldnamen(c, lang)

		runAdHocSql := func(isql string) (string, []interface{}, error) {
			isql = translateSqlAliases(c, conn, isql, lang)
			// Parameter-Map normalisieren (Upper)
			paramsUpper := make(map[string]interface{})
			for k, v := range req.Params {
				paramsUpper[strings.ToUpper(k)] = v
			}

			// 1. Übersetzungen (Case-Insensitive Ersatz)
			for _, t := range translations {
				replacement := t.Inhalt
				if replacement == "" {
					replacement = t.Betreff
				}
				// Wir nutzen Regex für Case-Insensitive Replace von §...§
				re := regexp.MustCompile("(?i)§" + regexp.QuoteMeta(t.Betreff) + "§")
				isql = re.ReplaceAllString(isql, replacement)
			}
			isql = reSection.ReplaceAllString(isql, "$1") // Reste entfernen falls nicht gefunden

			// 2. Backticks auflösen
			btMatches := reBacktick.FindAllStringSubmatch(isql, -1)
			for _, m := range btMatches {
				isql = strings.Replace(isql, m[0], "%"+m[1]+";TEXT%", 1)
			}

			// 3. Parameter extrahieren und konvertieren
			var args []interface{}
			finalSQL := isql
			paramMatches := rePara.FindAllStringSubmatch(isql, -1)
			for _, m := range paramMatches {
				label := strings.ToUpper(m[2])
				pType := strings.ToUpper(m[3])
				val, ok := paramsUpper[label]
				if !ok {
					return "", nil, fmt.Errorf("%s", label)
				}

				var validatedVal interface{}
				switch pType {
				case "NUMBER":
					f, _ := strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
					validatedVal = f
				case "DATE":
					validatedVal = fmt.Sprintf("%v", val)
				case "BOOL", "CHECKBOX":
					b := false
					if bv, ok := val.(bool); ok {
						b = bv
					} else if sv, ok := val.(string); ok {
						b = (sv == "1" || sv == "true")
					}
					if b {
						validatedVal = 1
					} else {
						validatedVal = 0
					}
				default:
					validatedVal = fmt.Sprintf("%v", val)
				}

				args = append(args, validatedVal)
				finalSQL = strings.Replace(finalSQL, m[0], "?", 1)
			}

			// 4. Einfache Platzhalter %PARAM% ohne Typ (Fallback)
			reSimple := regexp.MustCompile(`%([A-Za-z0-9_]{3,20})%`)
			simpleMatches := reSimple.FindAllStringSubmatch(finalSQL, -1)
			for _, m := range simpleMatches {
				label := strings.ToUpper(m[1])
				if label == "FIRMA." || label == "MASTER." || label == "DETAIL." {
					continue
				}
				val, ok := paramsUpper[label]
				if ok {
					args = append(args, fmt.Sprintf("%v", val))
					finalSQL = strings.Replace(finalSQL, m[0], "?", 1)
				}
			}

			return finalSQL, args, nil
		}

		// --- PARAMETER ERKENNUNG ---
		fullSql := req.MasterSql + " " + req.DetailSql
		requiredParams := make([]map[string]string, 0)
		seen := make(map[string]bool)

		// 1. §...§ Platzhalter prüfen (was nicht übersetzbar ist, ist ein Parameter)
		reSection := regexp.MustCompile(`(?i)§([^§]+)§`)
		sectionMatches := reSection.FindAllStringSubmatch(fullSql, -1)
		for _, m := range sectionMatches {
			label := strings.ToUpper(m[1])
			if seen[label] {
				continue
			}
			// Prüfen ob in Übersetzungen vorhanden (Case-Insensitive)
			found := false
			for _, t := range translations {
				if strings.ToUpper(t.Betreff) == label {
					found = true
					break
				}
			}
			if !found {
				requiredParams = append(requiredParams, map[string]string{
					"raw":   m[0],
					"label": label,
					"type":  "TEXT",
				})
				seen[label] = true
			}
		}

		// 2. Backticks auflösen für Scan
		btMatches := reBacktick.FindAllStringSubmatch(fullSql, -1)
		for _, m := range btMatches {
			if !seen[m[1]] {
				requiredParams = append(requiredParams, map[string]string{
					"raw":   m[0],
					"label": m[1],
					"type":  "TEXT",
				})
				seen[m[1]] = true
			}
		}

		// 3. %...;TYPE% und auch einfaches %...% (Fallback)
		paramMatches := rePara.FindAllStringSubmatch(fullSql, -1)
		for _, m := range paramMatches {
			label := m[2]
			if seen[label] {
				continue
			}
			requiredParams = append(requiredParams, map[string]string{
				"raw":   m[0],
				"label": label,
				"type":  strings.ToUpper(m[3]),
			})
			seen[label] = true
		}

		// Zusätzlicher Scan für %PARAM% ohne Semikolon (falls vorhanden)
		reSimplePara := regexp.MustCompile(`%([A-Za-z0-9_]{3,20})%`)
		simpleMatches := reSimplePara.FindAllStringSubmatch(fullSql, -1)
		for _, m := range simpleMatches {
			label := m[1]
			if seen[label] || label == "Firma." || label == "Master." || label == "Detail." {
				continue
			}
			requiredParams = append(requiredParams, map[string]string{
				"raw":   m[0],
				"label": label,
				"type":  "TEXT",
			})
			seen[label] = true
		}

		// Parameter-Maps für case-insensitive Suche vorbereiten
		providedParamsUpper := make(map[string]interface{})
		for k, v := range req.Params {
			providedParamsUpper[strings.ToUpper(k)] = v
		}

		// Überprüfung ob ALLE benötigten Parameter in req.Params vorhanden sind
		allPresent := true
		for _, rp := range requiredParams {
			if _, ok := providedParamsUpper[strings.ToUpper(rp["label"])]; !ok {
				allPresent = false
				break
			}
		}

		if len(requiredParams) > 0 && (!allPresent || len(req.Params) == 0) {
			c.JSON(http.StatusOK, gin.H{"needs_params": true, "params": requiredParams})
			return
		}
		// --- ENDE ERKENNUNG ---

		// Master ausführen
		var masterRows []map[string]interface{}
		var mSqlFinal string
		var err error
		if req.MasterSql != "" && req.MasterSql != "FOLDER" {
			var mArgs []interface{}
			mSqlFinal, mArgs, err = runAdHocSql(req.MasterSql)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Master-SQL: " + err.Error()})
				return
			}
			masterRows, err = executeQueryToMaps(c, conn, mSqlFinal, mArgs)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Master-EXEC: " + err.Error()})
				return
			}
		}

		// Detail ausführen
		var detailRows []map[string]interface{}
		var dSqlFinal string
		if req.DetailSql != "" {
			var dArgs []interface{}
			dSqlFinal, dArgs, err = runAdHocSql(req.DetailSql)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Detail-SQL: " + err.Error()})
				return
			}
			detailRows, err = executeQueryToMaps(c, conn, dSqlFinal, dArgs)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Detail-EXEC: " + err.Error()})
				return
			}
		}

		_, masterRows = translateResponseKeys(nil, masterRows, lang, conn)
		_, detailRows = translateResponseKeys(nil, detailRows, lang, conn)

		c.JSON(http.StatusOK, gin.H{
			"master":     masterRows,
			"details":    detailRows,
			"master_sql": mSqlFinal,
			"detail_sql": dSqlFinal,
		})
	})

	r.POST("/api/reports/execute/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Report-ID: " + idStr})
			return
		}

		// 1. Request Body für Parameter lesen
		var req struct {
			Params map[string]interface{} `json:"PARAMS"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			// Falls der Body leer ist oder kein JSON, ignorieren wir das (req.Params bleibt empty map)
			if req.Params == nil {
				req.Params = make(map[string]interface{})
			}
		}

		lang := c.DefaultQuery("lang", "de")
		if req.Params == nil {
			req.Params = make(map[string]interface{})
		}
		// System-Parameter injizieren (für SQL-Platzhalter)
		req.Params["LANG"] = lang
		req.Params["SPR_KZ"] = lang

		// 2. Report laden
		report, err := queries.GetDynamischeSQL(c, id)
		if err == nil {
			log.Printf("[REPORTS] Report geladen: ID=%d, Typ=%v, IstSumme=%d, SumSQL=%q", report.ID, report.TypKz, report.IstSummenzeile, report.Summenzeile)
		}
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Report nicht gefunden"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Datenbankfehler: " + err.Error()})
			}
			return
		}

		// 3. Typ bestimmen
		typ := strings.ToUpper(toString(report.TypKz))
		if typ == "" {
			typ = "S"
		}

		// --- STEP 1: PARAMS GLOBAL EXTRAHIEREN (FÜR ALLE TYPEN) ---
		combinedSQL := report.Sqlstatement + " " + toString(report.DetailSql)
		// Platzhalter §...§ vorab ersetzen für die Parameter-Suche
		translations, _ := queries.ListTranslateFeldnamen(c, lang)
		for _, t := range translations {
			placeholder := "§" + strings.ToUpper(t.Betreff) + "§"
			combinedSQL = strings.ReplaceAll(combinedSQL, placeholder, t.Inhalt)
		}

		// Robuster Parameter-Regex (Unterstützt :%LABEL;TYPE%, :NAME und `BACKTICK`)
		reBoth := regexp.MustCompile(`(['"]?):?%([^;%]+);([^;%]+)%['"]?|:([a-zA-Z0-9_]+)|` + "`([^`]+)`")
		paramMatches := reBoth.FindAllStringSubmatch(combinedSQL, -1)

		// Hilfsfunktion zur intelligenten Typ-Erkennung
		guessType := func(name string) string {
			un := strings.ToUpper(name)
			if strings.Contains(un, "DATUM") {
				return "DATE"
			}
			if strings.Contains(un, "AKTIV") || strings.HasSuffix(un, "_KZ") {
				return "BOOL"
			}
			if strings.HasSuffix(un, "_ID") || strings.Contains(un, "NUMMER") || strings.Contains(un, "KOSTEN") || strings.Contains(un, "BETRAG") {
				return "NUMBER"
			}
			return "TEXT"
		}

		requiredParams := make([]map[string]string, 0)
		seenParams := make(map[string]bool)
		for _, m := range paramMatches {
			label := ""
			pType := ""
			raw := m[0]

			if m[2] != "" {
				label = m[2]
				pType = strings.ToUpper(m[3])
			} else if m[4] != "" {
				label = m[4]
				pType = guessType(m[4])
			} else if m[5] != "" {
				label = m[5]
				if strings.Contains(label, ";") {
					parts := strings.Split(label, ";")
					label = parts[0]
					pType = strings.ToUpper(parts[1])
				} else {
					pType = guessType(label)
				}
			}

			if label != "" && !seenParams[label] {
				requiredParams = append(requiredParams, map[string]string{
					"raw":   raw,
					"label": label,
					"type":  pType,
				})
				seenParams[label] = true
			}
		}

		// Pre-Flight Check: Wenn Parameter benötigt, aber keine geschickt -> Frontend anpingen
		hasMissing := false
		if req.Params == nil {
			if len(requiredParams) > 0 {
				hasMissing = true
			}
		} else {
			for _, p := range requiredParams {
				label := p["label"]
				if _, ok := req.Params[label]; !ok {
					// Fallback check (case-insensitive)
					found := false
					for k := range req.Params {
						if strings.EqualFold(k, label) {
							found = true
							break
						}
					}
					if !found {
						hasMissing = true
						break
					}
				}
			}
		}

		if hasMissing {
			c.JSON(http.StatusOK, gin.H{"needs_params": true, "params": requiredParams})
			return
		}

		// Helper für SQL-Ausführung mit Parametern & Locking & Translations
		runSqlWithParams := func(isql string, currentParams map[string]interface{}) ([]map[string]interface{}, []string, error) {
			dbMutex.Lock()
			defer dbMutex.Unlock()

			// 1. Übersetzungen in diesem konkreten SQL-String
			for _, t := range translations {
				repl := t.Inhalt
				if repl == "" {
					repl = t.Betreff
				}
				isql = strings.ReplaceAll(isql, "§"+strings.ToUpper(t.Betreff)+"§", repl)
			}
			// Sektions-Marker entfernen (z.B. <!-- BEGIN Detail -->)
			isql = reSection.ReplaceAllString(isql, "$1")

			// 2. Parameter-Ersetzung durch Platzhalter (?)
			var localArgs []interface{}
			finalSQL := isql
			pMatches := reBoth.FindAllStringSubmatch(isql, -1)

			// Wir müssen die Ersetzung von links nach rechts machen, aber mit ?
			// Problem: strings.Replace ersetzt alle Vorkommen. Wir nutzen strings.Replace(..., 1) in der Schleife.
			for _, m := range pMatches {
				label := ""
				pType := ""

				if m[2] != "" {
					label = m[2]
					pType = strings.ToUpper(m[3])
				} else if m[4] != "" {
					label = m[4]
					pType = "TEXT"
				} else if m[5] != "" {
					label = m[5]
					pType = "TEXT"
				}

				// Globaler Check auf Semikolon (für alle Formate)
				if strings.Contains(label, ";") {
					parts := strings.Split(label, ";")
					label = parts[0]
					if pType == "TEXT" || pType == "" {
						pType = strings.ToUpper(parts[1])
					}
				}

				if pType == "" {
					pType = "TEXT"
				}

				val, ok := currentParams[label]
				if !ok {
					// Fallback: Vielleicht wurde es am Frontend anders gecased?
					for k, v := range currentParams {
						if strings.EqualFold(k, label) {
							val = v
							ok = true
							break
						}
					}
				}

				if !ok {
					return nil, nil, fmt.Errorf("fehlender Parameter: %s", label)
				}

				var validatedVal interface{}
				vStr := fmt.Sprintf("%v", val)
				switch pType {
				case "NUMBER":
					f, _ := strconv.ParseFloat(vStr, 64)
					validatedVal = f
				case "DATE", "TEXT":
					validatedVal = vStr
				default:
					validatedVal = val
				}
				localArgs = append(localArgs, validatedVal)
				finalSQL = strings.Replace(finalSQL, m[0], "?", 1)
			}

			// 3. Query ausführen
			log.Printf("[REPORTS] Executing SQL (Lang=%s): %s with params: %v", lang, finalSQL, localArgs)
			rows, err := conn.QueryContext(c, finalSQL, localArgs...)
			if err != nil {
				return nil, nil, err
			}
			defer rows.Close()

			cols, _ := rows.Columns()
			upperCols := make([]string, len(cols))
			for i, col := range cols {
				cleanCol := col
				if idx := strings.LastIndex(col, "."); idx != -1 {
					cleanCol = col[idx+1:]
				} else if idx := strings.LastIndex(col, ":"); idx != -1 {
					cleanCol = col[idx+1:]
				}
				upperCols[i] = strings.ToUpper(cleanCol)
			}

			var res []map[string]interface{}
			for rows.Next() {
				vals := make([]interface{}, len(cols))
				for i := range vals {
					vals[i] = new(interface{})
				}
				if err := rows.Scan(vals...); err != nil {
					continue
				}

				rowMap := make(map[string]interface{})
				for i, name := range upperCols {
					v := *(vals[i].(*interface{}))
					if b, ok := v.([]byte); ok {
						rowMap[name] = string(b)
					} else {
						rowMap[name] = v
					}
				}
				res = append(res, rowMap)
			}
			return res, upperCols, nil
		}

		// --- WEICHE NACH TYP ---

		// FALL T/M/G/L: Master-Detail-Reporting
		if typ == "T" || typ == "M" || typ == "G" || typ == "L" {
			// A: Global (Firma & Firmenparameter)
			firmaRows, _ := executeQueryToMaps(c, conn, "SELECT * FROM PERSON WHERE KZ = 'F' LIMIT 1", nil)
			var firma map[string]interface{}
			if len(firmaRows) > 0 {
				firma = firmaRows[0]
			}
			paramRows, _ := executeQueryToMaps(c, conn, "SELECT * FROM FIRMENPARAMETER LIMIT 1", nil)
			var firmenparams map[string]interface{}
			if len(paramRows) > 0 {
				firmenparams = paramRows[0]
			}

			// B: Master-SQL ausführen
			mSql := strings.TrimSpace(toString(report.SqlstatementNative))
			if mSql == "" {
				mSql = strings.TrimSpace(toString(report.Sqlstatement))
			}
			mSql = translateSqlAliases(c, conn, mSql, lang)
			masterRows, mCols, err := runSqlWithParams(mSql, req.Params)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Master-SQL Fehler: " + err.Error()})
				return
			}
			populateOriginalKeys(masterRows, lang, conn)

			// C: Template laden
			var tmplContent []byte
			tmplPath := toString(report.TemplateName)
			if tmplPath != "" {
				if !strings.HasSuffix(tmplPath, ".html") {
					tmplPath += ".html"
				}
				fullPath := filepath.Join("DOKU", tmplPath)
				tmplContent, _ = os.ReadFile(fullPath)
				if tmplContent == nil {
					fullPath = filepath.Join("templates", tmplPath)
					tmplContent, _ = os.ReadFile(fullPath)
				}
			}

			// Metadaten & Übersetzungen
			catalogFields, _ := queries.ListFieldTranslations(c, "de")
			metaMap := make(map[string]int64)
			for _, f := range catalogFields {
				metaMap[strings.ToUpper(f.Feldname)] = f.Nameindb
			}
			transMap := make(map[string]string)
			for _, t := range translations {
				transMap[strings.ToUpper(t.Betreff)] = t.Inhalt
			}

			// Initial leer, wird pro Master oder global befüllt
			var globalSumRows []map[string]interface{}
			var globalSCols []string
			masterSumsMap := make(map[interface{}][]map[string]interface{})

			if report.IstSummenzeile == 1 && toString(report.Summenzeile) != "" && !strings.Contains(strings.ToUpper(toString(report.Summenzeile)), "§MASTER.") {
				sRows, sc, err := runSqlWithParams(translateSqlAliases(c, conn, toString(report.Summenzeile), lang), req.Params)
				if err != nil {
					log.Printf("[REPORTS] Global Summenzeile Fehler: %v", err)
				} else {
					globalSumRows = sRows
					populateOriginalKeys(globalSumRows, lang, conn)
					globalSCols = sc
					log.Printf("[REPORTS] Globale Summenzeile geladen: %d Zeilen", len(globalSumRows))
				}
			}

			// D: Detail-Iteration
			allDetailRows := []map[string]interface{}{}
			filteredMasterRows := []map[string]interface{}{}
			var combinedHtmlParts []string
			baseDSql := strings.TrimSpace(toString(report.DetailSqlNative))
			if baseDSql == "" {
				baseDSql = strings.TrimSpace(toString(report.DetailSql))
			}
			baseDSql = translateSqlAliases(c, conn, baseDSql, lang)

			// Automatische Verknüpfung ermitteln (Konvention: ID_TABELLENNAME)
			var dCols []string
			masterTable := ""
			mSql = toString(report.SqlstatementNative)
			if mSql == "" {
				mSql = toString(report.Sqlstatement)
			}
			reFrom := regexp.MustCompile(`(?i)FROM\s+([a-zA-Z0-9_]+)`)
			if m := reFrom.FindStringSubmatch(mSql); len(m) > 1 {
				masterTable = strings.ToUpper(m[1])
			}

			for i, master := range masterRows {
				// Robustes Extrahieren der Master-ID
				var mID interface{}
				for k, v := range master {
					cleanK := strings.TrimSpace(strings.ToUpper(k))
					if cleanK == "ID" || cleanK == "ID_HERDEN" || cleanK == "HERDEN_ID" {
						mID = v
						break
					}
				}
				if mID == nil {
					mID = master["ID"]
				}
				if mID == nil {
					mID = master["id"]
				}

				// Falls immer noch keine ID da ist, nutzen wir den Index als virtuelle ID
				if mID == nil {
					mID = fmt.Sprintf("v-%d", i)
					master["_MASTER_ID_"] = mID // Stempel für Master
				}

				// Diagnose: Alle verfügbaren Keys ausgeben
				keys := []string{}
				for k := range master {
					keys = append(keys, k)
				}
				log.Printf("[REPORTS] Verfügbare Master-Felder: %v", keys)

				dSql := baseDSql

				// Automatischer Link falls nichts definiert ist
				if dSql != "" && !strings.Contains(strings.ToUpper(dSql), "§MASTER.") && masterTable != "" {
					linkCol := "ID_" + masterTable
					connector := " WHERE "
					if strings.Contains(strings.ToUpper(dSql), " WHERE ") {
						connector = " AND "
					}
					dSql += connector + linkCol + " = §MASTER.ID§"
					log.Printf("[REPORTS] Auto-Link angewendet: %s", dSql)
				}

				var dRows []map[string]interface{}
				if dSql != "" {
					// §MASTER.FELD§ Platzhalter
					reM := regexp.MustCompile(`(?i)§MASTER\.([^§]+)§`)
					dSql = reM.ReplaceAllStringFunc(dSql, func(m string) string {
						fieldName := strings.TrimSpace(strings.ToUpper(reM.FindStringSubmatch(m)[1]))

						var val interface{}
						found := false
						for k, v := range master {
							if strings.TrimSpace(strings.ToUpper(k)) == fieldName {
								val = v
								found = true
								break
							}
						}

						if found && val != nil {
							vS := fmt.Sprintf("%v", val)
							if _, err := strconv.ParseFloat(vS, 64); err == nil {
								return vS
							}
							return "'" + strings.ReplaceAll(vS, "'", "''") + "'"
						}
						return "NULL"
					})

					log.Printf("[REPORTS] Link-ID: %v | Generated Detail SQL: %s", mID, dSql)

					// Automatischer Join
					masterTable := ""
					mSqlUpper := strings.ToUpper(mSql)
					fromIdx := strings.Index(mSqlUpper, "FROM ")
					if fromIdx != -1 {
						parts := strings.Fields(mSql[fromIdx+5:])
						if len(parts) > 0 {
							masterTable = strings.ToUpper(strings.Trim(parts[0], " \n\r\t\"`[]"))
						}
					}
					hasManualLink := strings.Contains(strings.ToUpper(baseDSql), "§MASTER.") || (masterTable != "" && strings.Contains(strings.ToUpper(baseDSql), "ID_"+masterTable))
					if masterTable != "" && mID != nil && !hasManualLink {
						fkName := "ID_" + masterTable
						joinClause := fmt.Sprintf(" %s = %v ", fkName, mID)
						dSqlUpper := strings.ToUpper(dSql)
						insIdx := len(dSql)
						for _, marker := range []string{" GROUP BY ", " ORDER BY ", " LIMIT ", " OFFSET "} {
							if idx := strings.Index(dSqlUpper, marker); idx != -1 && idx < insIdx {
								insIdx = idx
							}
						}
						if strings.Contains(strings.ToUpper(dSql[:insIdx]), " WHERE ") {
							dSql = dSql[:insIdx] + " AND " + joinClause + dSql[insIdx:]
						} else {
							dSql = dSql[:insIdx] + " WHERE " + joinClause + dSql[insIdx:]
						}
					}

					rows, cols, err := runSqlWithParams(dSql, req.Params)
					if err == nil {
						dRows = rows
						populateOriginalKeys(dRows, lang, conn)
						if len(dCols) == 0 {
							dCols = cols
						}
					}
				}

				for i := range dRows {
					dRows[i]["_MASTER_ID_"] = mID
				}

				// Wenn ein Detail-SQL vorhanden ist, aber keine Sätze gefunden wurden, den Master überspringen
				if dSql != "" && len(dRows) == 0 {
					continue
				}

				filteredMasterRows = append(filteredMasterRows, master)
				allDetailRows = append(allDetailRows, dRows...)

				// --- Summenzeile pro Master (Subtotals) ---
				var currentSumRows []map[string]interface{}
				if report.IstSummenzeile == 1 && toString(report.Summenzeile) != "" {
					sumSql := translateSqlAliases(c, conn, toString(report.Summenzeile), lang)
					if strings.Contains(strings.ToUpper(sumSql), "§MASTER.") {
						// Platzhalter ersetzen
						reM := regexp.MustCompile(`(?i)§MASTER\.([^§]+)§`)
						sumSql = reM.ReplaceAllStringFunc(sumSql, func(m string) string {
							fieldName := strings.TrimSpace(strings.ToUpper(reM.FindStringSubmatch(m)[1]))
							for k, v := range master {
								if strings.TrimSpace(strings.ToUpper(k)) == fieldName {
									vS := fmt.Sprintf("%v", v)
									if _, err := strconv.ParseFloat(vS, 64); err == nil {
										return vS
									}
									return "'" + strings.ReplaceAll(vS, "'", "''") + "'"
								}
							}
							return "NULL"
						})
						sRows, _, err := runSqlWithParams(sumSql, req.Params)
						if err == nil {
							currentSumRows = sRows
							populateOriginalKeys(currentSumRows, lang, conn)
							if mID != nil {
								masterSumsMap[mID] = sRows
							}
						}
					} else {
						// Globale Summen nutzen
						currentSumRows = globalSumRows
					}
				}

				if len(tmplContent) > 0 {
					html := processTemplateReport(string(tmplContent), firma, firmenparams, master, dRows, metaMap, report, transMap, currentSumRows)
					combinedHtmlParts = append(combinedHtmlParts, html)
				}
			}

			// E: HTML Zusammenbau
			finalHTML := ""
			isInteractiveGrid := (typ == "M" || typ == "G")
			if len(combinedHtmlParts) > 0 {
				finalHTML = combinedHtmlParts[0]
				if len(combinedHtmlParts) > 1 {
					var inner strings.Builder
					for i := 1; i < len(combinedHtmlParts); i++ {
						html := combinedHtmlParts[i]

						startKey := "<div class=\"main-content\">"
						if !strings.Contains(html, startKey) {
							startKey = "<body"
						}

						endKey := "<!-- FOOTER -->"
						if !strings.Contains(html, endKey) {
							endKey = "</body>"
						}

						startIdx := strings.Index(html, startKey)
						if startIdx != -1 && startKey == "<body" {
							closingBracket := strings.Index(html[startIdx:], ">")
							if closingBracket != -1 {
								startIdx += closingBracket + 1
							}
						}
						endIdx := strings.Index(html, endKey)

						if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
							inner.WriteString("<div style='page-break-before: always; padding-top: 1cm; border-top: 3px solid #333;' class='no-print'></div>")
							inner.WriteString(html[startIdx:endIdx])
						}
					}
					finalHTML = strings.Replace(finalHTML, "</body>", inner.String()+"</body>", 1)
				}
			} else if (isInteractiveGrid || typ == "T") && len(filteredMasterRows) > 0 {
				finalHTML = generateDefaultMasterDetailHtml(report, firma, firmenparams, filteredMasterRows, mCols, allDetailRows, dCols, metaMap, transMap, globalSumRows, globalSCols, masterSumsMap)
			} else if typ == "S" || typ == "L" {
				if len(filteredMasterRows) > 0 {
					finalHTML = generateDefaultListHtml(report, firma, firmenparams, filteredMasterRows, mCols, metaMap, globalSumRows, globalSCols)
				}
			}

			mCols, filteredMasterRows = translateResponseKeys(mCols, filteredMasterRows, lang, conn)
			dCols, allDetailRows = translateResponseKeys(dCols, allDetailRows, lang, conn)
			_, globalSumRows = translateResponseKeys(nil, globalSumRows, lang, conn)

			c.JSON(http.StatusOK, gin.H{
				"typ":              typ,
				"html":             finalHTML,
				"columns":          mCols,
				"masterRows":       filteredMasterRows,
				"details":          allDetailRows,
				"detail_columns":   dCols,
				"sums":             globalSumRows,
				"show_master_grid": toInt64(report.ShowMasterGrid) == 1,
				"show_detail_grid": toInt64(report.ShowDetailGrid) == 1,
			})
			return
		}

		if typ == "S" || typ == "L" {
			translatedSQL := translateSqlAliases(c, conn, report.Sqlstatement, lang)
			res, cols, err := runSqlWithParams(translatedSQL, req.Params)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Lesen der Daten: " + err.Error()})
				return
			}
			populateOriginalKeys(res, lang, conn)

			var sumRows []map[string]interface{}
			var sCols []string
			if report.IstSummenzeile == 1 && toString(report.Summenzeile) != "" {
				sumSql := translateSqlAliases(c, conn, toString(report.Summenzeile), lang)
				sRows, sc, err := runSqlWithParams(sumSql, req.Params)
				if err != nil {
					log.Printf("[REPORTS] Summenzeile Fehler: %v", err)
				} else {
					sumRows = sRows
					populateOriginalKeys(sumRows, lang, conn)
					sCols = sc
				}
			}
			cols, res = translateResponseKeys(cols, res, lang, conn)
			_, sumRows = translateResponseKeys(nil, sumRows, lang, conn)

			// Falls HTML angefordert wurde (für Druck)
			if isTrue(req.Params["_PRINT_"]) {
				firmaRows, _ := executeQueryToMaps(c, conn, "SELECT * FROM PERSON WHERE KZ = 'F' LIMIT 1", nil)
				var firma map[string]interface{}
				if len(firmaRows) > 0 {
					firma = firmaRows[0]
				}
				paramRows, _ := executeQueryToMaps(c, conn, "SELECT * FROM FIRMENPARAMETER LIMIT 1", nil)
				var firmenparams map[string]interface{}
				if len(paramRows) > 0 {
					firmenparams = paramRows[0]
				}
				catalogFields, _ := queries.ListFieldTranslations(c, "de")
				metaMap := make(map[string]int64)
				for _, f := range catalogFields {
					metaMap[strings.ToUpper(f.Feldname)] = f.Nameindb
				}

				html := generateDefaultListHtml(report, firma, firmenparams, res, cols, metaMap, sumRows, sCols)
				c.JSON(http.StatusOK, gin.H{
					"typ":              "S",
					"data":             res,
					"sums":             sumRows,
					"columns":          cols,
					"html":             html,
					"show_master_grid": true,
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"typ":              "S",
				"data":             res,
				"sums":             sumRows,
				"columns":          cols,
				"show_master_grid": true,
				"show_detail_grid": false,
			})
			return
		}

		// FALL H: Kategorie
		if typ == "H" {
			c.JSON(http.StatusOK, gin.H{"typ": "H", "message": "Kategorie-Record (No SQL)"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "Unbekannter Report-Typ"})
	})

	r.POST("/api/reports/preview/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Report-ID: " + idStr})
			return
		}
		var req struct {
			Params map[string]interface{} `json:"PARAMS"`
		}
		_ = c.ShouldBindJSON(&req)

		report, err := queries.GetDynamischeSQL(c, id)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Report nicht gefunden"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Datenbankfehler: " + err.Error()})
			}
			return
		}

		lang := c.DefaultQuery("lang", "de")
		if req.Params == nil {
			req.Params = make(map[string]interface{})
		}
		req.Params["LANG"] = lang
		req.Params["SPR_KZ"] = lang

		translations, _ := queries.ListTranslateFeldnamen(c, lang)
		sqlStr := translateSqlAliases(c, conn, report.Sqlstatement, lang)
		for _, t := range translations {
			placeholder := "§" + strings.ToUpper(t.Betreff) + "§"
			replacement := t.Inhalt
			if replacement == "" {
				replacement = t.Betreff
			}
			sqlStr = strings.ReplaceAll(sqlStr, placeholder, replacement)
		}

		matches := rePara.FindAllStringSubmatch(sqlStr, -1)
		btMatches := reBacktick.FindAllStringSubmatch(sqlStr, -1)

		var paramDefMap map[string]interface{}
		if toString(report.ParamDef) != "" {
			json.Unmarshal([]byte(toString(report.ParamDef)), &paramDefMap)
		}

		finalSQL := sqlStr

		// NEU: Logik für Standard SQL Parameter :PARAM
		reParaNormal := regexp.MustCompile(`(?i):([A-Z0-9_]+)`)
		paraMatches := reParaNormal.FindAllStringSubmatch(sqlStr, -1)

		// Helper zum Mappen und Ersetzen
		replaceInPreview := func(raw, label, typ string) {
			val, ok := req.Params[label]
			if ok {
				var sVal string
				// Falls der Wert ein Datum/String ist, brauchen wir Anführungszeichen,
				// außer der Platzhalter im SQL hatte bereits welche (raw enthält dann quotes)
				hasQuotes := strings.HasPrefix(raw, "'") || strings.HasPrefix(raw, "\"")

				switch typ {
				case "NUMBER", "BOOLEAN", "BOOL", "CHOICE":
					b := false
					if bv, ok := val.(bool); ok {
						b = bv
					} else if fv, ok := val.(float64); ok {
						b = fv != 0
					} else if sv, ok := val.(string); ok {
						b = (sv == "1" || strings.ToLower(sv) == "true")
					}
					if typ == "NUMBER" {
						sVal = fmt.Sprintf("%v", val)
					} else if typ == "CHOICE" {
						sVal = fmt.Sprintf("%v", val)
						if strings.ToUpper(sVal) == "NULL" {
							sVal = "NULL"
						}
					} else {
						if b {
							sVal = "1"
						} else {
							sVal = "0"
						}
					}
				case "DATE", "STRING", "TEXT":
					str := fmt.Sprintf("%v", val)
					str = strings.ReplaceAll(str, "'", "''") // SQL-Escape
					if hasQuotes {
						sVal = str
					} else {
						sVal = "'" + str + "'"
					}
				default:
					str := fmt.Sprintf("%v", val)
					if hasQuotes {
						sVal = str
					} else {
						sVal = "'" + str + "'"
					}
				}
				finalSQL = strings.Replace(finalSQL, raw, sVal, 1)
			} else {
				finalSQL = strings.Replace(finalSQL, raw, "[FEHLT]", 1)
			}
		}

		for _, p := range matches {
			replaceInPreview(p[0], p[2], strings.ToUpper(p[3]))
		}
		for _, p := range btMatches {
			term := p[1]
			pLabel := term
			pType := "TEXT"
			if pDef, ok := paramDefMap[term].(map[string]interface{}); ok {
				if l, ok := pDef["label"].(string); ok && l != "" {
					pLabel = l
				}
				if t, ok := pDef["type"].(string); ok && t != "" {
					pType = strings.ToUpper(t)
				}
			}
			replaceInPreview(p[0], pLabel, pType)
		}
		for _, p := range paraMatches {
			pLabel := strings.ToUpper(p[1])
			pType := "TEXT"
			if strings.Contains(pLabel, "DATUM") {
				pType = "DATE"
			} else if strings.HasPrefix(pLabel, "ID_") || strings.HasSuffix(pLabel, "_ID") || pLabel == "ID" {
				pType = "NUMBER"
			}
			if pDef, ok := paramDefMap[pLabel].(map[string]interface{}); ok {
				if t, ok := pDef["type"].(string); ok && t != "" {
					pType = strings.ToUpper(t)
				}
			}
			replaceInPreview(p[0], pLabel, pType)
		}

		// Typ-Weiche für Preview-Response
		typ := ""
		if v, ok := report.TypKz.(string); ok {
			typ = v
		} else if v, ok := report.TypKz.([]uint8); ok {
			typ = string(v)
		}

		if typ == "T" {
			// Für Typ T liefern wir beide SQLs zurück
			masterSQL := finalSQL
			detailSQLStr := toString(report.DetailSql)

			// Übersetzungen für DetailSQL ebenfalls machen
			for _, t := range translations {
				detailSQLStr = strings.ReplaceAll(detailSQLStr, "§"+strings.ToUpper(t.Betreff)+"§", t.Inhalt)
			}
			// Parameter in DetailSQL ersetzen (ähnlich wie oben für Master)
			for _, p := range matches {
				val, ok := req.Params[p[2]]
				if ok {
					replacement := fmt.Sprintf("'%v'", val)
					detailSQLStr = strings.ReplaceAll(detailSQLStr, p[0], replacement)
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"typ":        "T",
				"master_sql": masterSQL,
				"detail_sql": detailSQLStr,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"sql": finalSQL})
	})

	r.POST("/api/reports/preview", func(c *gin.Context) {
		var req struct {
			SQL    string                 `json:"sql"`
			Params map[string]interface{} `json:"params"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Anfrage"})
			return
		}

		if req.SQL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Kein SQL-Statement angegeben"})
			return
		}

		lang := c.DefaultQuery("lang", "de")
		finalSQL := translateSqlAliases(c, conn, req.SQL, lang)

		// 1. Übersetzungen (§...§)
		translations, _ := queries.ListTranslateFeldnamen(c, lang)
		for _, t := range translations {
			placeholder := "§" + strings.ToUpper(t.Betreff) + "§"
			replacement := t.Inhalt
			if replacement == "" {
				replacement = t.Betreff
			}
			finalSQL = strings.ReplaceAll(finalSQL, placeholder, replacement)
		}

		// 2. Explizite Parameter (Backticks, §...§ oder :...)
		for k, v := range req.Params {
			sVal := fmt.Sprintf("%v", v)
			isNumeric := true
			if _, err := strconv.ParseFloat(sVal, 64); err != nil {
				isNumeric = false
			}

			finalVal := sVal
			if !isNumeric && !strings.HasPrefix(sVal, "'") {
				finalVal = "'" + sVal + "'"
			}

			// Case-Insensitive Regex für alle drei Varianten mit Wortgrenze (\b)
			re1 := regexp.MustCompile("(?i)`" + regexp.QuoteMeta(k) + "`")
			re2 := regexp.MustCompile("(?i)§" + regexp.QuoteMeta(k) + "§")
			re3 := regexp.MustCompile("(?i):" + regexp.QuoteMeta(k) + "\\b")

			finalSQL = re1.ReplaceAllString(finalSQL, finalVal)
			finalSQL = re2.ReplaceAllString(finalSQL, finalVal)
			finalSQL = re3.ReplaceAllString(finalSQL, finalVal)
		}

		rows, err := conn.QueryContext(c, finalSQL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("SQL Fehler: %v", err)})
			return
		}
		defer rows.Close()

		columns, _ := rows.Columns()
		var results []map[string]interface{}

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			row := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			}
			results = append(results, row)
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    results,
			"columns": columns,
		})
	})

	r.GET("/api/reports/template/:name", func(c *gin.Context) {
		tmplName := c.Param("name")
		if tmplName == "" || tmplName == "undefined" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiger Template-Name"})
			return
		}
		if !strings.HasSuffix(tmplName, ".html") {
			tmplName += ".html"
		}
		path := filepath.Join("DOKU", tmplName)
		content, err := os.ReadFile(path)
		if err != nil {
			path = filepath.Join("templates", tmplName)
			content, err = os.ReadFile(path)
		}
		if err != nil {
			// Falls es gar nicht existiert, liefern wir ein leeres Skelett zurück
			skeleton := "<!DOCTYPE html>\n<html>\n<body>\n  <h1>Neues Template</h1>\n</body>\n</html>"
			c.JSON(http.StatusOK, gin.H{"content": skeleton, "is_new": true})
			return
		}
		c.JSON(http.StatusOK, gin.H{"content": string(content)})
	})

	r.POST("/api/reports/template/:name", func(c *gin.Context) {
		tmplName := c.Param("name")
		if tmplName == "" || tmplName == "undefined" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiger Template-Name"})
			return
		}
		if !strings.HasSuffix(tmplName, ".html") {
			tmplName += ".html"
		}
		var req struct {
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		path := filepath.Join("DOKU", tmplName)
		os.MkdirAll("DOKU", 0755)
		if err := os.WriteFile(path, []byte(req.Content), 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Speicherfehler: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Template gespeichert"})
	})

	// TABELLEN-SEEDING (Optional: INITIALE DATEN FÜR FELD_KATALOG)
	fields, _ := queries.ListFieldTranslations(context.Background(), "de")
	if len(fields) == 0 {
		defaultFields := []struct {
			KZ   string
			FELD string
			DESC string
		}{
			{"H", "ID_HERDEN", "Herde"},
			{"D", "BUCHUNGSDATUM", "Datum"},
			{"K", "KLASSEA", "Klasse A"},
			{"X", "XL", "Größe XL"},
			{"L", "LARGE", "Größe L"},
			{"M", "MEDIUM", "Größe M"},
			{"S", "SMALL", "Größe S"},
			{"V", "VERLUSTE", "Verluste"},
			{"G", "KONTROLLGEWICHT", "Kontrollgewicht"},
			{"J", "KL6", "Jumbo"},
		}
		for _, f := range defaultFields {
			res, _ := queries.CreateFieldKatalog(context.Background(), db.CreateFieldKatalogParams{
				Kz:       f.KZ,
				Feldname: f.FELD,
				Nameindb: 1,
			})
			if res.ID > 0 {
				_, _ = queries.UpsertFieldTranslation(context.Background(), db.UpsertFieldTranslationParams{
					IDFeldKatalog: res.ID,
					SpracheKz:     "de",
					Inhalt:        f.DESC,
					Betreff:       f.FELD,
				})
			}
		}
	}

	// --- BENUTZER PROFILE & BERECHTIGUNGEN ---

	adminProfile, _ := queries.GetBenutzerProfilByKZ(context.Background(), "A")
	if adminProfile.ID == 0 {
		_, _ = queries.CreateBenutzerProfil(context.Background(), db.CreateBenutzerProfilParams{
			ProfilKz:                "A",
			Beschreibung:            "Administrator mit Vollzugriff",
			FDashboard:              1,
			FHerdenVerwalten:        1,
			FEinrichtungenVerwalten: 1,
			FPersonenVerwalten:      1,
			FBuchungenErfassen:      1,
			FAuswertungenAnzeigen:   1,
			FSqlStrukturVerwalten:   1,
			FBenutzerProfile:        1,
			FParameterEditieren:     1,
			FKostenVerwalten:        1,
			FTabellenAnzeigen:       1,
			FTexteVerwalten:         1,
			FSystemVerwaltung:       1,
			FBackupErstellen:        1,
		})

		log.Println("Default Admin-Profil 'A' angelegt.")
	}

	// Seeding Default Admin User if not exists
	adminUser, _ := queries.GetBenutzerByUsername(context.Background(), "Admin")
	if adminUser.ID == 0 {
		pass, _ := encryptAES("1234", "HuhnLite")
		_, _ = queries.CreateBenutzer(context.Background(), db.CreateBenutzerParams{
			Username:          "Admin",
			Passwort:          pass,
			IDBenutzerProfile: adminProfile.ID,
		})
		log.Println("Default Admin-User 'Admin' angelegt (Passwort: 1234).")
	}

	r.GET("/api/userprofiles", func(c *gin.Context) {
		res, err := queries.ListBenutzerProfile(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/userprofiles/:kz", func(c *gin.Context) {
		kz := c.Param("kz")
		res, err := queries.GetBenutzerProfilByKZ(c, kz)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Profil nicht gefunden"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/api/permissions/:kz", func(c *gin.Context) {
		kz := c.Param("kz")
		res, err := queries.GetBenutzerProfilByKZ(c, kz)
		if err != nil {
			// Falls Profil nicht existiert, geben wir leere Rechte zurück
			c.JSON(http.StatusOK, gin.H{})
			return
		}

		// Erstelle eine einfache Map für das Frontend
		permissions := map[string]bool{
			"dashboard":               res.FDashboard == 1,
			"herden_verwalten":        res.FHerdenVerwalten == 1,
			"einrichtungen_verwalten": res.FEinrichtungenVerwalten == 1,
			"personen_verwalten":      res.FPersonenVerwalten == 1,
			"buchungen_erfassen":      res.FBuchungenErfassen == 1,
			"auswertungen_anzeigen":   res.FAuswertungenAnzeigen == 1,
			"sql_struktur_verwalten":  res.FSqlStrukturVerwalten == 1,
			"benutzer_profile":        res.FBenutzerProfile == 1,
			"parameter_editieren":     res.FParameterEditieren == 1,
			"kosten_verwalten":        res.FKostenVerwalten == 1,
			"tabellen_anzeigen":       res.FTabellenAnzeigen == 1,
			"texte_verwalten":         res.FTexteVerwalten == 1,
			"system_verwaltung":       res.FSystemVerwaltung == 1,
			"backup_erstellen":        res.FBackupErstellen == 1,
		}
		c.JSON(http.StatusOK, permissions)
	})

	r.POST("/api/userprofiles", func(c *gin.Context) {
		var req struct {
			ProfilKz     string `json:"PROFIL_KZ"`
			Beschreibung string `json:"BESCHREIBUNG"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		params := db.CreateBenutzerProfilParams{
			ProfilKz:     req.ProfilKz,
			Beschreibung: req.Beschreibung,
		}
		res, err := queries.CreateBenutzerProfil(c, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/userprofiles/:kz", func(c *gin.Context) {
		kz := c.Param("kz")
		var req struct {
			Beschreibung            string `json:"BESCHREIBUNG"`
			FDashboard              int64  `json:"F_DASHBOARD"`
			FHerdenVerwalten        int64  `json:"F_HERDEN_VERWALTEN"`
			FEinrichtungenVerwalten int64  `json:"F_EINRICHTUNGEN_VERWALTEN"`
			FPersonenVerwalten      int64  `json:"F_PERSONEN_VERWALTEN"`
			FBuchungenErfassen      int64  `json:"F_BUCHUNGEN_ERFASSEN"`
			FAuswertungenAnzeigen   int64  `json:"F_AUSWERTUNGEN_ANZEIGEN"`
			FSqlStrukturVerwalten   int64  `json:"F_SQL_STRUKTUR_VERWALTEN"`
			FBenutzerProfile        int64  `json:"F_BENUTZER_PROFILE"`
			FParameterEditieren     int64  `json:"F_PARAMETER_EDITIEREN"`
			FKostenVerwalten        int64  `json:"F_KOSTEN_VERWALTEN"`
			FTabellenAnzeigen       int64  `json:"F_TABELLEN_ANZEIGEN"`
			FTexteVerwalten         int64  `json:"F_TEXTE_VERWALTEN"`
			FSystemVerwaltung       int64  `json:"F_SYSTEM_VERWALTUNG"`
			FBackupErstellen        int64  `json:"F_BACKUP_ERSTELLEN"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("PUT /api/userprofiles/%s: BindJSON Error: %v", kz, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("PUT /api/userprofiles/%s: Saving %v", kz, req)

		params := db.UpdateBenutzerProfilParams{
			Beschreibung:            req.Beschreibung,
			FDashboard:              req.FDashboard,
			FHerdenVerwalten:        req.FHerdenVerwalten,
			FEinrichtungenVerwalten: req.FEinrichtungenVerwalten,
			FPersonenVerwalten:      req.FPersonenVerwalten,
			FBuchungenErfassen:      req.FBuchungenErfassen,
			FAuswertungenAnzeigen:   req.FAuswertungenAnzeigen,
			FSqlStrukturVerwalten:   req.FSqlStrukturVerwalten,
			FBenutzerProfile:        req.FBenutzerProfile,
			FParameterEditieren:     req.FParameterEditieren,
			FKostenVerwalten:        req.FKostenVerwalten,
			FTabellenAnzeigen:       req.FTabellenAnzeigen,
			FTexteVerwalten:         req.FTexteVerwalten,
			FSystemVerwaltung:       req.FSystemVerwaltung,
			FBackupErstellen:        req.FBackupErstellen,
			ProfilKz:                kz,
		}

		res, err := queries.UpdateBenutzerProfil(c, params)

		if err != nil {
			log.Printf("PUT /api/userprofiles/%s: DB Update Error: %v", kz, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/userprofiles/:kz", func(c *gin.Context) {
		kz := c.Param("kz")
		if kz == "A" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Das Admin-Profil kann nicht gelöscht werden"})
			return
		}
		err := queries.DeleteBenutzerProfil(c, kz)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Profil gelöscht"})
	})

	// --- BENUTZER VERWALTUNG ---

	r.GET("/api/benutzer", func(c *gin.Context) {
		res, err := queries.ListBenutzer(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.POST("/api/benutzer", func(c *gin.Context) {
		var req struct {
			Username          string `json:"USERNAME"`
			Passwort          string `json:"PASSWORT"`
			IDBenutzerProfile int64  `json:"ID_BENUTZER_PROFILE"`
			Klarname          string `json:"KLARNAME"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		encryptedPass, err := encryptAES(req.Passwort, "HuhnLite")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Verschlüsselungsfehler"})
			return
		}

		res, err := queries.CreateBenutzer(c, db.CreateBenutzerParams{
			Username:          req.Username,
			Passwort:          encryptedPass,
			IDBenutzerProfile: req.IDBenutzerProfile,
			Klarname:          req.Klarname,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/benutzer/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		var req struct {
			Passwort          string `json:"PASSWORT"`
			IDBenutzerProfile int64  `json:"ID_BENUTZER_PROFILE"`
			Klarname          string `json:"KLARNAME"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		encryptedPass, err := encryptAES(req.Passwort, "HuhnLite")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Verschlüsselungsfehler"})
			return
		}

		err = queries.UpdateBenutzer(c, db.UpdateBenutzerParams{
			ID:                idInt,
			Passwort:          encryptedPass,
			IDBenutzerProfile: req.IDBenutzerProfile,
			Klarname:          req.Klarname,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Benutzer aktualisiert"})
	})

	r.DELETE("/api/benutzer/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		idInt, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		err = queries.DeleteBenutzer(c, idInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Benutzer gelöscht"})
	})

	r.POST("/api/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"USERNAME"`
			Passwort string `json:"PASSWORT"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := queries.GetBenutzerByUsername(c, req.Username)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer oder Passwort"})
			return
		}

		decryptedPass, err := decryptAES(user.Passwort, "HuhnLite")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer oder Passwort"})
			return
		}

		if decryptedPass != req.Passwort {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Ungültiger Benutzer oder Passwort"})
			return
		}

		// Profil & Berechtigungen laden
		profile, err := queries.GetBenutzerProfilByID(c, user.IDBenutzerProfile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Berechtigungen konnten nicht geladen werden"})
			return
		}

		permissions := map[string]bool{
			"dashboard":               profile.FDashboard == 1,
			"herden_verwalten":        profile.FHerdenVerwalten == 1,
			"einrichtungen_verwalten": profile.FEinrichtungenVerwalten == 1,
			"personen_verwalten":      profile.FPersonenVerwalten == 1,
			"buchungen_erfassen":      profile.FBuchungenErfassen == 1,
			"auswertungen_anzeigen":   profile.FAuswertungenAnzeigen == 1,
			"sql_struktur_verwalten":  profile.FSqlStrukturVerwalten == 1,
			"benutzer_profile":        profile.FBenutzerProfile == 1,
			"parameter_editieren":     profile.FParameterEditieren == 1,
			"kosten_verwalten":        profile.FKostenVerwalten == 1,
			"tabellen_anzeigen":       profile.FTabellenAnzeigen == 1,
			"texte_verwalten":         profile.FTexteVerwalten == 1,
			"system_verwaltung":       profile.FSystemVerwaltung == 1,
			"backup_erstellen":        profile.FBackupErstellen == 1,
		}

		c.JSON(http.StatusOK, gin.H{
			"id":          user.ID,
			"username":    user.Username,
			"klarname":    user.Klarname,
			"profile_kz":  profile.ProfilKz,
			"permissions": permissions,
		})
	})

	// --- SCHEMA INFO ENDPOINTS FOR REPORT BUILDER ---
	r.GET("/api/schema/tables", func(c *gin.Context) {
		rows, err := conn.Query("SELECT name FROM sqlite_master JOIN SHOWTV ON sqlite_master.name = SHOWTV.TVNAME WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%' AND SHOWTV.SHOWIT = 1 ORDER BY name")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var tables []string
		for rows.Next() {
			var name string
			rows.Scan(&name)
			tables = append(tables, name)
		}
		c.JSON(http.StatusOK, tables)
	})

	r.GET("/api/schema/columns/:table", func(c *gin.Context) {
		table := c.Param("table")
		var rows *sql.Rows
		var err error
		if database.Engine == "mysql" {
			rows, err = conn.Query("SELECT 0 as cid, COLUMN_NAME as name, DATA_TYPE as dtype, IF(IS_NULLABLE='NO', 1, 0) as notnull, COLUMN_DEFAULT as dfltValue, IF(COLUMN_KEY='PRI', 1, 0) as pk FROM information_schema.COLUMNS WHERE TABLE_NAME = ? AND TABLE_SCHEMA = DATABASE()", table)
		} else {
			rows, err = conn.Query(fmt.Sprintf("PRAGMA table_info(\"%s\")", table))
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var columns []string
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dflt_value interface{}
			rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk)
			columns = append(columns, name)
		}
		c.JSON(http.StatusOK, columns)
	})

	r.GET("/api/user-state/:key", func(c *gin.Context) {
		key := c.Param("key")
		username := c.Query("username")
		if username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username erforderlich"})
			return
		}

		var value string
		err := conn.QueryRow("SELECT VALUE FROM USER_STATE WHERE USERNAME = ? AND KEY = ?", username, key).Scan(&value)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"value": ""})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"value": value})
	})

	r.POST("/api/user-state", func(c *gin.Context) {
		var req struct {
			Username string `json:"USERNAME"`
			Key      string `json:"KEY"`
			Value    string `json:"VALUE"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		upsertQuery := `INSERT INTO USER_STATE (USERNAME, KEY, VALUE) VALUES (?, ?, ?)
			ON CONFLICT(USERNAME, KEY) DO UPDATE SET VALUE = excluded.VALUE`
		if database.Engine == "mysql" {
			upsertQuery = `INSERT INTO USER_STATE (USERNAME, KEY, VALUE) VALUES (?, ?, ?)
				ON DUPLICATE KEY UPDATE VALUE = VALUES(VALUE)`
		}
		_, err := conn.Exec(upsertQuery, req.Username, req.Key, req.Value)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Status gespeichert"})
	})

	r.GET("/api/aktionen", func(c *gin.Context) {
		idUser, _ := strconv.ParseInt(c.Query("id_user"), 10, 64)
		start := c.Query("start")
		end := c.Query("end")
		kz := c.Query("kz")
		
		showErledigtRaw := c.Query("show_erledigt")
		
		// Status filter mapping:
		// showErledigtRaw = "0" -> status = 0 (Open)
		// showErledigtRaw = "1" -> status = 1 (Completed)
		// showErledigtRaw = "2" -> status = 2 (All)
		// If frontend sends true/false, we might need to map it carefully.
		// The previous logic was: checkbox unchecked -> 0 (Open). checked -> 1 (All).
		// Now we want: unchecked -> 0 (Open). checked -> 1 (Completed). or maybe the frontend will be updated to send 0, 1, 2.
		// Let's support 0, 1, 2 directly.
		statusInt := 0
		if showErledigtRaw == "1" || showErledigtRaw == "true" {
			statusInt = 1
		} else if showErledigtRaw == "2" {
			statusInt = 2
		}
		
		fmt.Printf("[API] GET /api/aktionen: user=%d, start='%s', end='%s', kz='%s', raw_erledigt='%s' -> status=%d\n", 
			idUser, start, end, kz, showErledigtRaw, statusInt)

		res, err := queries.ListAktionen(c, db.ListAktionenParams{
			IDUser:    idUser,
			StartDate: start,
			EndDate:   end,
			Kz:        kz,
			Status:    int64(statusInt),
		})
		if err != nil {
			fmt.Printf("[API] Error ListAktionen: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		type AktionenResponse struct {
			ID               int64  `json:"id"`
			AktionenKz       string `json:"aktionen_kz"`
			IDUser           int64  `json:"id_user"`
			Aktionsdatum     string `json:"aktionsdatum"`
			Bezeichnung      string `json:"bezeichnung"`
			IntervallTage    int64  `json:"intervall_tage"`
			AnzahlIntervalle int64  `json:"anzahl_intervalle"`
			Erledigt         int64  `json:"erledigt"`
			IDUserErledigt   int64  `json:"id_user_erledigt"`
			Username         string `json:"username"`
			UsernameErledigt string `json:"username_erledigt"`
			ErledigtAm       string `json:"erledigt_am"`
			Bemerkung        string `json:"bemerkung"`
		}

		formatted := make([]AktionenResponse, len(res))
		for i, r := range res {
			kzStr := ""
			if s, ok := r.AktionenKz.(string); ok {
				kzStr = s
			} else if b, ok := r.AktionenKz.([]byte); ok {
				kzStr = string(b)
			}

			formatted[i] = AktionenResponse{
				ID:               r.ID,
				AktionenKz:       kzStr,
				IDUser:           r.IDUser.Int64,
				Aktionsdatum:     r.Aktionsdatum.String,
				Bezeichnung:      r.Bezeichnung.String,
				IntervallTage:    r.IntervallTage.Int64,
				AnzahlIntervalle: r.AnzahlIntervalle.Int64,
				Erledigt:         r.Erledigt.Int64,
				IDUserErledigt:   r.IDUserErledigt.Int64,
				Username:         r.Username.String,
				UsernameErledigt: r.UsernameErledigt.String,
				ErledigtAm:       r.ErledigtAm.String,
				Bemerkung:        r.Bemerkung.String,
			}
		}

		fmt.Printf("[API] ListAktionen returned %d rows\n", len(res))
		c.JSON(http.StatusOK, formatted)
	})

	r.POST("/api/aktionen", func(c *gin.Context) {
		var req struct {
			AktionenKz       string `json:"aktionen_kz"`
			IDUser           int64  `json:"id_user"`
			Aktionsdatum     string `json:"aktionsdatum"`
			Bezeichnung      string `json:"bezeichnung"`
			IntervallTage    int64  `json:"intervall_tage"`
			AnzahlIntervalle int64  `json:"anzahl_intervalle"`
			Erledigt         int64  `json:"erledigt"`
			IDUserErledigt   int64  `json:"id_user_erledigt"`
			ErledigtAm       string `json:"erledigt_am"`
			Bemerkung        string `json:"bemerkung"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Erledigt == 1 && req.ErledigtAm == "" {
			req.ErledigtAm = time.Now().Format("2006-01-02 15:04:05")
		}

		res, err := queries.CreateAktion(c, db.CreateAktionParams{
			AktionenKz:       toNullString(req.AktionenKz),
			IDUser:           toNullInt64(req.IDUser),
			Aktionsdatum:     toNullString(req.Aktionsdatum),
			Bezeichnung:      toNullString(req.Bezeichnung),
			IntervallTage:    toNullInt64(req.IntervallTage),
			AnzahlIntervalle: toNullInt64(req.AnzahlIntervalle),
			Erledigt:         toNullInt64(req.Erledigt),
			IDUserErledigt:   toNullInt64(req.IDUserErledigt),
			ErledigtAm:       toNullString(req.ErledigtAm),
			Bemerkung:        toNullString(req.Bemerkung),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.PUT("/api/aktionen/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		var req struct {
			AktionenKz       string `json:"aktionen_kz"`
			IDUser           int64  `json:"id_user"`
			Aktionsdatum     string `json:"aktionsdatum"`
			Bezeichnung      string `json:"bezeichnung"`
			IntervallTage    int64  `json:"intervall_tage"`
			AnzahlIntervalle int64  `json:"anzahl_intervalle"`
			Erledigt         int64  `json:"erledigt"`
			IDUserErledigt   int64  `json:"id_user_erledigt"`
			ErledigtAm       string `json:"erledigt_am"`
			Bemerkung        string `json:"bemerkung"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Erledigt == 1 && req.ErledigtAm == "" {
			req.ErledigtAm = time.Now().Format("2006-01-02 15:04:05")
		} else if req.Erledigt == 0 {
			req.ErledigtAm = ""
		}

		res, err := queries.UpdateAktion(c, db.UpdateAktionParams{
			ID:               id,
			AktionenKz:       toNullString(req.AktionenKz),
			IDUser:           toNullInt64(req.IDUser),
			Aktionsdatum:     toNullString(req.Aktionsdatum),
			Bezeichnung:      toNullString(req.Bezeichnung),
			IntervallTage:    toNullInt64(req.IntervallTage),
			AnzahlIntervalle: toNullInt64(req.AnzahlIntervalle),
			Erledigt:         toNullInt64(req.Erledigt),
			IDUserErledigt:   toNullInt64(req.IDUserErledigt),
			ErledigtAm:       toNullString(req.ErledigtAm),
			Bemerkung:        toNullString(req.Bemerkung),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.DELETE("/api/aktionen/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}
		if err := queries.DeleteAktion(c, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	return r
}

// --- AES ENCRYPTION HELPERS ---

func encryptAES(plainText, key string) (string, error) {
	byteKey := make([]byte, 16)
	copy(byteKey, []byte(key))
	block, err := aes.NewCipher(byteKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptAES(cipherText, key string) (string, error) {
	byteKey := make([]byte, 16)
	copy(byteKey, []byte(key))
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(byteKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func decodeBase64(s string) []byte {
	if s == "" {
		return nil
	}
	if idx := strings.Index(s, ","); idx != -1 {
		s = s[idx+1:]
	}
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil
	}
	return data
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func toFloat64(i interface{}) (float64, bool) {
	if i == nil {
		return 0, false
	}
	switch v := i.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	case []byte:
		f, err := strconv.ParseFloat(string(v), 64)
		return f, err == nil
	case sql.NullFloat64:
		if v.Valid {
			return v.Float64, true
		}
		return 0, false
	case sql.NullInt64:
		if v.Valid {
			return float64(v.Int64), true
		}
		return 0, false
	}
	return 0, false
}

func formatValue(fieldName string, val interface{}, metadata map[string]int64) string {
	if val == nil {
		return ""
	}
	formatFloat := func(f float64, decimals int) string {
		s := fmt.Sprintf("%.*f", decimals, f)
		parts := strings.Split(s, ".")
		intPart := parts[0]
		isNegative := false
		if strings.HasPrefix(intPart, "-") {
			isNegative = true
			intPart = intPart[1:]
		}
		var res []string
		for len(intPart) > 3 {
			res = append([]string{intPart[len(intPart)-3:]}, res...)
			intPart = intPart[:len(intPart)-3]
		}
		if len(intPart) > 0 {
			res = append([]string{intPart}, res...)
		}
		finalInt := strings.Join(res, ".")
		if isNegative {
			finalInt = "-" + finalInt
		}
		if len(parts) > 1 {
			return finalInt + "," + parts[1]
		}
		return finalInt
	}
	f, ok := toFloat64(val)
	if !ok {
		return fmt.Sprintf("%v", val)
	}

	upperName := strings.ToUpper(fieldName)
	if meta, ok := metadata[upperName]; ok && meta == 0 {
		return formatFloat(f, 2)
	}
	currencyPatterns := []string{"PREIS", "NETTO", "BRUTTO", "KOSTEN", "BETRAG", "ERLOES", "UMSATZ", "KOSTENTAG", "EINSTALLKOSTEN", "ABSCHREIBUNG"}
	for _, p := range currencyPatterns {
		if strings.Contains(upperName, p) {
			return formatFloat(f, 2)
		}
	}
	// Fallback: Wenn Ganzzahl, dann ohne Dezimalstellen (z.B. Silonummer)
	if f == float64(int64(f)) {
		return formatFloat(f, 0)
	}
	return formatFloat(f, 2)
}

func isRightAligned(fieldName string, val interface{}) bool {
	if val == nil {
		return false
	}
	upperName := strings.ToUpper(fieldName)
	// Datumserkennung
	if strings.Contains(upperName, "DATUM") || strings.Contains(upperName, "ZEIT") {
		return true
	}
	// Numerische Erkennung
	_, ok := toFloat64(val)
	return ok
}

func getSortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func executeQueryToMaps(ctx context.Context, conn *sql.DB, sqlStr string, args []interface{}) ([]map[string]interface{}, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	rows, err := conn.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var results []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		for i := range vals {
			vals[i] = new(interface{})
		}
		if err := rows.Scan(vals...); err != nil {
			continue
		}
		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			upperCol := strings.ToUpper(colName)
			val := *(vals[i].(*interface{}))
			if b, ok := val.([]byte); ok {
				rowMap[upperCol] = string(b)
			} else {
				rowMap[upperCol] = val
			}
		}
		results = append(results, rowMap)
	}
	return results, nil
}

func processTemplateReport(templateContent string, firma map[string]interface{}, fparams map[string]interface{}, master map[string]interface{}, details []map[string]interface{}, metadata map[string]int64, report db.DynamischeSql, translations map[string]string, sumRows []map[string]interface{}) string {
	res := templateContent
	for k, v := range translations {
		res = strings.ReplaceAll(res, "§"+k+"§", v)
	}
	res = strings.ReplaceAll(res, "§REPORT_TITLE§", report.Beschreibung)
	res = strings.ReplaceAll(res, "§REPORT_HEADER§", "Bericht: "+report.Beschreibung)
	now := time.Now()
	res = strings.ReplaceAll(res, "§#DATE§", now.Format("02.01.2006"))
	for k, v := range firma {
		val := formatValue(k, v, metadata)
		res = strings.ReplaceAll(res, "<Firma."+k+">", val)
		res = strings.ReplaceAll(res, "§FIRMA."+strings.ToUpper(k)+"§", val)
	}
	for k, v := range master {
		val := formatValue(k, v, metadata)
		res = strings.ReplaceAll(res, "<Master."+k+">", val)
		res = strings.ReplaceAll(res, "§MASTER."+strings.ToUpper(k)+"§", val)
	}
	reDetail := regexp.MustCompile("(?s)<!-- BEGIN Detail -->(.*?)<!-- END Detail -->")
	res = reDetail.ReplaceAllStringFunc(res, func(fullBlock string) string {
		match := reDetail.FindStringSubmatch(fullBlock)
		if len(match) < 2 {
			return fullBlock
		}
		var combinedRowHtml strings.Builder
		for _, row := range details {
			rowHtml := match[1]
			for k, v := range row {
				val := formatValue(k, v, metadata)
				rowHtml = strings.ReplaceAll(rowHtml, "<Detail."+k+">", val)
				rowHtml = strings.ReplaceAll(rowHtml, "§DETAIL."+strings.ToUpper(k)+"§", val)
			}
			combinedRowHtml.WriteString(rowHtml)
		}
		return combinedRowHtml.String()
	})

	// Summen-Block
	reSum := regexp.MustCompile("(?s)<!-- BEGIN Summe -->(.*?)<!-- END Summe -->")
	res = reSum.ReplaceAllStringFunc(res, func(fullBlock string) string {
		match := reSum.FindStringSubmatch(fullBlock)
		if len(match) < 2 {
			return fullBlock
		}
		var combinedSumHtml strings.Builder
		for _, row := range sumRows {
			rowHtml := match[1]
			for k, v := range row {
				val := formatValue(k, v, metadata)
				rowHtml = strings.ReplaceAll(rowHtml, "<Summe."+k+">", val)
				rowHtml = strings.ReplaceAll(rowHtml, "§SUMME."+strings.ToUpper(k)+"§", val)
			}
			combinedSumHtml.WriteString(rowHtml)
		}
		return combinedSumHtml.String()
	})
	return res
}

func toCamelCase(s string) string {
	if s == "" {
		return ""
	}
	parts := regexp.MustCompile("[_ ]").Split(strings.ToLower(s), -1)
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func syncFieldCatalog(dbConn *sql.DB, sqlStr string) string {
	if sqlStr == "" {
		return ""
	}
	re := regexp.MustCompile(`(?i)\s+AS\s+([a-zA-Z0-9_]+)`)
	matches := re.FindAllStringSubmatch(sqlStr, -1)
	for _, match := range matches {
		alias := match[1]
		upperAlias := strings.ToUpper(alias)
		camelAlias := toCamelCase(alias)
		var fieldID int
		err := dbConn.QueryRow("SELECT ID FROM FELD_KATALOG WHERE upper(FELDNAME) = ?", upperAlias).Scan(&fieldID)
		if err == sql.ErrNoRows {
			res, err := dbConn.Exec("INSERT INTO FELD_KATALOG (FELDNAME, NAMEINDB) VALUES (?, 0)", upperAlias)
			if err == nil {
				newID, _ := res.LastInsertId()
				dbConn.Exec("INSERT OR IGNORE INTO TRANSLATEFELDNAMEN (ID_FELD_KATALOG, SPRACHE_KZ, BETREFF) VALUES (?, 'de', ?)", newID, camelAlias)
			}
		}
	}
	return sqlStr
}

func generateDefaultMasterDetailHtml(report db.DynamischeSql, firma, fparams map[string]interface{}, masterRows []map[string]interface{}, mCols []string, allDetails []map[string]interface{}, dCols []string, metadata map[string]int64, translations map[string]string, sumRows []map[string]interface{}, sCols []string, masterSums map[interface{}][]map[string]interface{}) string {
	log.Printf("[REPORTS] generateDefaultMasterDetailHtml: MasterRows=%d, SumRows=%d, MasterSums=%d", len(masterRows), len(sumRows), len(masterSums))
	getSumVal := func(sRow map[string]interface{}, col string) interface{} {
		if v, ok := sRow[col]; ok {
			return v
		}
		for k, v := range sRow {
			if strings.EqualFold(k, col) {
				return v
			}
		}
		clean := func(s string) string {
			return strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
					return r
				}
				return -1
			}, s)
		}
		colClean := clean(col)
		for k, v := range sRow {
			if clean(k) == colClean && colClean != "" {
				return v
			}
		}
		return nil
	}

	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html><html><head><meta charset='UTF-8'><style>")
	sb.WriteString("body { font-family: sans-serif; margin: 0; padding: 1cm; color: #333; } ")
	sb.WriteString(".header { border-bottom: 2px solid #1976D2; padding-bottom: 10px; margin-bottom: 20px; display: flex; justify-content: space-between; } ")
	sb.WriteString(".master-block { background-color: #f9f9f9; border: 1px solid #ccc; padding: 15px; margin-bottom: 20px; } ")
	sb.WriteString(".detail-table { width: 100%; border-collapse: collapse; font-size: 11px; } ")
	sb.WriteString(".detail-table th { background-color: #eee; border: 1px solid #ccc; padding: 8px; text-align: left; } ")
	sb.WriteString(".detail-table td { border: 1px solid #ccc; padding: 6px; } ")
	sb.WriteString(".text-right { text-align: right; } ")
	sb.WriteString("@media print { .no-print { display: none !important; } } ")
	sb.WriteString("</style></head><body style='padding-top: 50px;'>")
	sb.WriteString("<div class='no-print' style='position:fixed; top:0; left:0; right:0; background:#333; color:white; padding:10px; display:flex; gap:10px; z-index:9999; box-shadow: 0 2px 5px rgba(0,0,0,0.3); font-family:sans-serif;'>")
	sb.WriteString("<button onclick='window.print()' style='background:#2196F3; border:none; color:white; padding:8px 15px; border-radius:4px; cursor:pointer; font-weight:bold;'>Drucken</button>")
	sb.WriteString("<button onclick='window.close()' style='background:#f44336; border:none; color:white; padding:8px 15px; border-radius:4px; cursor:pointer; font-weight:bold;'>Beenden (Schließen)</button>")
	sb.WriteString("<div style='margin-left:auto; display:flex; align-items:center; opacity:0.8; font-size:13px;'>Report-Navigator</div>")
	sb.WriteString("</div>")
	sb.WriteString("<div class='header'><h1>" + report.Beschreibung + "</h1></div>")
	for _, master := range masterRows {
		sb.WriteString("<div class='master-block'><h3>Stammdaten</h3>")
		// Master-Spalten in der mCols Reihenfolge ausgeben
		colsToUse := mCols
		if len(colsToUse) == 0 {
			colsToUse = getSortedKeys(master)
		}
		for _, col := range colsToUse {
			v := master[col]
			if !strings.HasPrefix(col, "_") {
				alignClass := ""
				if isRightAligned(col, v) {
					alignClass = " class='text-right'"
				}
				sb.WriteString(fmt.Sprintf("<div%s><strong>%s:</strong> %s</div>", alignClass, col, formatValue(col, v, metadata)))
			}
		}
		sb.WriteString("</div>")
		mID := master["ID"]
		if mID == nil {
			mID = master["id"]
		}
		var masterDetails []map[string]interface{}
		for _, d := range allDetails {
			if d["_MASTER_ID_"] == mID {
				masterDetails = append(masterDetails, d)
			}
		}
		if len(masterDetails) > 0 {
			// Spaltennamen nutzen (bereits in dCols SQL Reihenfolge)
			cols := dCols
			if len(cols) == 0 {
				for k := range masterDetails[0] {
					if !strings.HasPrefix(k, "_") {
						cols = append(cols, k)
					}
				}
				sort.Strings(cols)
			}

			sb.WriteString("<table class='detail-table'><thead><tr>")
			for _, col := range cols {
				alignClass := ""
				if isRightAligned(col, masterDetails[0][col]) {
					alignClass = " class='text-right'"
				}
				sb.WriteString("<th" + alignClass + ">" + col + "</th>")
			}
			sb.WriteString("</tr></thead><tbody>")
			for _, dRow := range masterDetails {
				sb.WriteString("<tr>")
				for _, col := range cols {
					val := dRow[col]
					alignClass := ""
					if isRightAligned(col, val) {
						alignClass = " class='text-right'"
					}
					sb.WriteString("<td" + alignClass + ">" + formatValue(col, val, metadata) + "</td>")
				}
				sb.WriteString("</tr>")
			}
			// --- Sub-Summen direkt in die Detail-Tabelle einfügen für perfekte Ausrichtung ---
			if mSums, ok := masterSums[mID]; ok && len(mSums) > 0 {
				for _, sRow := range mSums {
					sb.WriteString("<tr style='background:#f9f9f9; font-weight:bold; border-top: 2px solid #999;'>")
					for _, col := range dCols {
						val := getSumVal(sRow, col)
						alignClass := ""
						if isRightAligned(col, val) {
							alignClass = " class='text-right'"
						}
						// Wir nutzen formatValue, aber für Summen evtl. fett markieren
						formatted := formatValue(col, val, metadata)
						sb.WriteString("<td" + alignClass + ">" + formatted + "</td>")
					}
					sb.WriteString("</tr>")
				}
			}
			sb.WriteString("</tbody></table>")
		}
	}

	if len(sumRows) > 0 {
		log.Printf("[REPORTS] Rendering SumRows in Master-Detail: %d rows", len(sumRows))
		sb.WriteString("<div style='margin-top: 30px; border-top: 3px solid #333; padding-top:10px;'><h3>Gesamtsummen</h3>")
		sb.WriteString("<table class='detail-table'><thead><tr>")
		// Spalten aus sCols nutzen (SQL Reihenfolge)
		colsToUse := sCols
		if len(colsToUse) == 0 {
			for k := range sumRows[0] {
				if !strings.HasPrefix(k, "_") {
					colsToUse = append(colsToUse, k)
				}
			}
			sort.Strings(colsToUse)
		}
		for _, col := range colsToUse {
			sb.WriteString("<th>" + col + "</th>")
		}
		sb.WriteString("</tr></thead><tbody>")
		for _, sRow := range sumRows {
			sb.WriteString("<tr style='background:#f1f1f1; font-weight:bold;'>")
			for _, col := range colsToUse {
				val := sRow[col]
				alignClass := ""
				if isRightAligned(col, val) {
					alignClass = " class='text-right'"
				}
				sb.WriteString("<td" + alignClass + ">" + formatValue(col, val, metadata) + "</td>")
			}
			sb.WriteString("</tr>")
		}
		sb.WriteString("</tbody></table></div>")
	}

	sb.WriteString("</body></html>")
	return sb.String()
}

func generateDefaultListHtml(report db.DynamischeSql, firma, firmenparams map[string]interface{}, rows []map[string]interface{}, cols []string, metadata map[string]int64, sumRows []map[string]interface{}, sCols []string) string {
	// Hilfsfunktion für Case-Insensitive Zugriff auf Summenzeilen
	getSumVal := func(sRow map[string]interface{}, col string) interface{} {
		if v, ok := sRow[col]; ok {
			return v
		}
		for k, v := range sRow {
			if strings.EqualFold(k, col) {
				return v
			}
		}
		// Robust fallback
		clean := func(s string) string {
			return strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
					return r
				}
				return -1
			}, s)
		}
		colClean := clean(col)
		for k, v := range sRow {
			if clean(k) == colClean && colClean != "" {
				return v
			}
		}
		return nil
	}

	var sb strings.Builder
	orientation := "portrait"
	if report.PageOrientation == "L" {
		orientation = "landscape"
	}
	totalPages := 1
	if report.RowsPerPage > 0 && len(rows) > 0 {
		totalPages = (len(rows) + int(report.RowsPerPage) - 1) / int(report.RowsPerPage)
	}

	sb.WriteString("<html><head><style>@page { size: " + orientation + "; margin: 1cm; } table{width:100%;border-collapse:collapse;font-family:sans-serif;font-size:11px; margin-bottom: 20px;} th,td{border:1px solid #ccc;padding:6px;text-align:left;} th{background-color:#eee; font-weight: bold;} .text-right{text-align:right;} .sum-row { background-color: #f1f1f1; font-weight: bold; } @media print { .no-print { display: none !important; } } </style>")
	sb.WriteString("<script>")
	sb.WriteString("function scrollToPage(n) { const el = document.getElementById('report-page-'+n); if(el) el.scrollIntoView({behavior:'smooth'}); }")
	sb.WriteString("</script>")
	sb.WriteString("</head><body style='padding-top: 60px;'>")
	sb.WriteString("<div class='no-print' style='position:fixed; top:0; left:0; right:0; background:#444; color:white; padding:10px; display:flex; gap:10px; z-index:9999; box-shadow: 0 2px 8px rgba(0,0,0,0.4); font-family:sans-serif; align-items:center;'>")
	sb.WriteString("<button onclick='window.print()' style='background:#2196F3; border:none; color:white; padding:8px 15px; border-radius:4px; cursor:pointer; font-weight:bold;'>Drucken</button>")
	sb.WriteString("<button onclick='window.close()' style='background:#e91e63; border:none; color:white; padding:8px 15px; border-radius:4px; cursor:pointer; font-weight:bold;'>Verlassen</button>")

	if totalPages > 1 {
		sb.WriteString("<div style='display:flex; gap:5px; margin-left:20px; align-items:center;'>")
		sb.WriteString("<button onclick='scrollToPage(1)' style='background:#555; border:1px solid #777; color:white; padding:5px 10px; border-radius:4px; cursor:pointer;'>&lt;&lt;</button>")
		sb.WriteString("<button id='prevBtn' style='background:#555; border:1px solid #777; color:white; padding:5px 10px; border-radius:4px; cursor:pointer;'>&lt;</button>")
		sb.WriteString("<span style='margin:0 10px; font-size:14px; min-width:80px; text-align:center;'>Seite / " + fmt.Sprintf("%d", totalPages) + "</span>")
		sb.WriteString("<button id='nextBtn' style='background:#555; border:1px solid #777; color:white; padding:5px 10px; border-radius:4px; cursor:pointer;'>&gt;</button>")
		sb.WriteString("<button onclick='scrollToPage(" + fmt.Sprintf("%d", totalPages) + ")' style='background:#555; border:1px solid #777; color:white; padding:5px 10px; border-radius:4px; cursor:pointer;'>&gt;&gt;</button>")
		sb.WriteString("</div>")
		sb.WriteString("<script>")
		sb.WriteString("let currentPage = 1; const maxPages = " + fmt.Sprintf("%d", totalPages) + "; ")
		sb.WriteString("function updatePage(inc) { currentPage = Math.max(1, Math.min(maxPages, currentPage + inc)); scrollToPage(currentPage); }")
		sb.WriteString("document.getElementById('prevBtn').onclick = () => updatePage(-1);")
		sb.WriteString("document.getElementById('nextBtn').onclick = () => updatePage(1);")
		sb.WriteString("</script>")
	}

	sb.WriteString("<div style='margin-left:auto; display:flex; align-items:center; opacity:0.8; font-weight:bold; font-family: sans-serif;'>" + report.Beschreibung + "</div>")
	sb.WriteString("</div>")
	sb.WriteString("<div id='report-page-1'><h1>" + report.Beschreibung + "</h1><table>")
	log.Printf("[REPORTS] generateDefaultListHtml: Rows=%d, SumRows=%d, sCols=%v", len(rows), len(sumRows), sCols)
	if len(rows) > 0 {
		if len(cols) == 0 {
			log.Printf("[REPORTS] WARNUNG: generateDefaultListHtml erhielt leere cols. Extrahiere aus rows[0].")
			for k := range rows[0] {
				if !strings.HasPrefix(k, "_") {
					cols = append(cols, k)
				}
			}
			sort.Strings(cols)
		}

		sb.WriteString("<thead><tr>")
		for _, col := range cols {
			alignClass := ""
			if isRightAligned(col, rows[0][col]) {
				alignClass = " class='text-right'"
			}
			sb.WriteString("<th" + alignClass + ">" + col + "</th>")
		}
		sb.WriteString("</tr></thead><tbody>")
		for i, row := range rows {
			if report.RowsPerPage > 0 && i > 0 && i%int(report.RowsPerPage) == 0 {
				pNum := (i / int(report.RowsPerPage)) + 1
				sb.WriteString("</tbody></table><div style='page-break-after: always;'></div>")
				sb.WriteString(fmt.Sprintf("<div id='report-page-%d'></div>", pNum))
				sb.WriteString("<table><thead><tr>")
				for _, col := range cols {
					alignClass := ""
					if isRightAligned(col, rows[0][col]) {
						alignClass = " class='text-right'"
					}
					sb.WriteString("<th" + alignClass + ">" + col + "</th>")
				}
				sb.WriteString("</tr></thead><tbody>")
			}

			sb.WriteString("<tr>")
			for _, col := range cols {
				val := row[col]
				alignClass := ""
				if isRightAligned(col, val) {
					alignClass = " class='text-right'"
				}
				sb.WriteString("<td" + alignClass + ">" + formatValue(col, val, metadata) + "</td>")
			}
			sb.WriteString("</tr>")
		}
		sb.WriteString("</tbody></table>")

		if len(sumRows) > 0 {
			log.Printf("[REPORTS] Rendering SumRows in List: %d rows, sCols provided: %v", len(sumRows), sCols)
			sb.WriteString("<div style='margin-top: 30px; border-top: 2px solid #333; padding-top: 15px;'>")
			sb.WriteString("<h3 style='font-size: 14px; margin-bottom: 10px; font-family: sans-serif; color: #1976D2;'>Gesamtsummen</h3>")
			sb.WriteString("<table style='width: auto; min-width: 40%'><thead><tr>")

			sumColsToUse := sCols
			if len(sumColsToUse) == 0 {
				sumColsToUse = cols
			}

			for _, col := range sumColsToUse {
				alignClass := ""
				if isRightAligned(col, sumRows[0][col]) {
					alignClass = " class='text-right'"
				}
				sb.WriteString("<th" + alignClass + ">" + col + "</th>")
			}
			sb.WriteString("</tr></thead><tbody>")
			for _, sRow := range sumRows {
				sb.WriteString("<tr style='background-color: #f1f1f1; font-weight: bold;'>")
				for _, col := range sumColsToUse {
					val := getSumVal(sRow, col)
					alignClass := ""
					if isRightAligned(col, val) {
						alignClass = " class='text-right'"
					}
					if val == nil {
						sb.WriteString("<td></td>")
					} else {
						sb.WriteString("<td" + alignClass + ">" + formatValue(col, val, metadata) + "</td>")
					}
				}
				sb.WriteString("</tr>")
			}
			sb.WriteString("</tbody></table></div>")
		}
	}
	sb.WriteString("<br/><p style='font-size:10px; color:gray; text-align:right; font-family: sans-serif;'>Erstellt am: " + time.Now().Format("02.01.2006 15:04") + "</p></body></html>")
	return sb.String()
}
func migrateDB(database *wailsdb.DB) {
	db := database.SQL
	log.Println("Prüfe Datenbank-Schema...")

	// 1. VERKAUF Tabelle sicherstellen
	verkaufTable := `CREATE TABLE IF NOT EXISTS VERKAUF (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		ID_EILAGERBUCHUNG INTEGER NOT NULL DEFAULT 0 UNIQUE,
		ID_BUCHUNG INTEGER NOT NULL DEFAULT 0,
		BUCHUNGSDATUM TEXT NOT NULL DEFAULT '0001-01-01',
		MENGESMALL INTEGER NOT NULL DEFAULT 0,
		MENGEMEDIUM INTEGER NOT NULL DEFAULT 0,
		MENGELARGE INTEGER NOT NULL DEFAULT 0,
		MENGEXL INTEGER NOT NULL DEFAULT 0,
		PREISSMALL REAL NOT NULL DEFAULT 0.0,
		PREISMEDIUM REAL NOT NULL DEFAULT 0.0,
		PREISLARGE REAL NOT NULL DEFAULT 0.0,
		PREISXL REAL NOT NULL DEFAULT 0.0,
		GESAMTPREIS REAL NOT NULL DEFAULT 0.0,
		BIO BOOLEAN NOT NULL DEFAULT 0,
		VERBUCHT BOOLEAN NOT NULL DEFAULT 0,
		CHARGE TEXT NOT NULL DEFAULT '',
		RABATTPROZENT REAL NOT NULL DEFAULT 0.0
	)`
	if database.Engine == "mysql" {
		verkaufTable = `CREATE TABLE IF NOT EXISTS VERKAUF (
			ID INTEGER PRIMARY KEY AUTO_INCREMENT,
			ID_EILAGERBUCHUNG INTEGER NOT NULL DEFAULT 0 UNIQUE,
			ID_BUCHUNG INTEGER NOT NULL DEFAULT 0,
			BUCHUNGSDATUM VARCHAR(25) NOT NULL DEFAULT '0001-01-01',
			MENGESMALL INTEGER NOT NULL DEFAULT 0,
			MENGEMEDIUM INTEGER NOT NULL DEFAULT 0,
			MENGELARGE INTEGER NOT NULL DEFAULT 0,
			MENGEXL INTEGER NOT NULL DEFAULT 0,
			PREISSMALL DECIMAL(10, 2) NOT NULL DEFAULT 0.0,
			PREISMEDIUM DECIMAL(10, 2) NOT NULL DEFAULT 0.0,
			PREISLARGE DECIMAL(10, 2) NOT NULL DEFAULT 0.0,
			PREISXL DECIMAL(10, 2) NOT NULL DEFAULT 0.0,
			GESAMTPREIS DECIMAL(10, 2) NOT NULL DEFAULT 0.0,
			BIO BOOLEAN NOT NULL DEFAULT 0,
			VERBUCHT BOOLEAN NOT NULL DEFAULT 0,
			CHARGE VARCHAR(255) NOT NULL DEFAULT '',
			RABATTPROZENT DECIMAL(5, 2) NOT NULL DEFAULT 0.0
		)`
	}
	_, err := db.Exec(verkaufTable)
	if err != nil {
		log.Printf("Fehler beim Erstellen der Tabelle VERKAUF: %v", err)
	}

	// 1b. USER_STATE Tabelle sicherstellen
	userStateTable := `CREATE TABLE IF NOT EXISTS USER_STATE (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		USERNAME TEXT NOT NULL,
		KEY TEXT NOT NULL,
		VALUE TEXT NOT NULL,
		UNIQUE(USERNAME, KEY)
	)`
	if database.Engine == "mysql" {
		userStateTable = `CREATE TABLE IF NOT EXISTS USER_STATE (
			ID INTEGER PRIMARY KEY AUTO_INCREMENT,
			USERNAME VARCHAR(255) NOT NULL,
			` + "`KEY`" + ` VARCHAR(255) NOT NULL,
			VALUE TEXT NOT NULL,
			UNIQUE(USERNAME, ` + "`KEY`" + `)
		)`
	}
	_, err = db.Exec(userStateTable)
	if err != nil {
		log.Printf("Fehler beim Erstellen der Tabelle USER_STATE: %v", err)
	}

	// 1bb. SYSTEMSETTINGS Tabelle sicherstellen
	sysSettingsTable := `CREATE TABLE IF NOT EXISTS SYSTEMSETTINGS (
		NAME TEXT PRIMARY KEY,
		VALUE TEXT
	)`
	if database.Engine == "mysql" {
		sysSettingsTable = `CREATE TABLE IF NOT EXISTS SYSTEMSETTINGS (
			NAME VARCHAR(255) PRIMARY KEY,
			VALUE TEXT
		)`
	}
	_, err = db.Exec(sysSettingsTable)
	if err != nil {
		log.Printf("Fehler beim Erstellen der Tabelle SYSTEMSETTINGS: %v", err)
	}

	// 1c. SILO Tabelle sicherstellen
	siloTable := `CREATE TABLE IF NOT EXISTS SILO (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		SILONUMMER INTEGER NOT NULL DEFAULT 0,
		PERSONENNUMMER INTEGER NOT NULL DEFAULT 0,
		ID_LIEFERANT INTEGER NOT NULL DEFAULT 0,
		BEZEICHNUNG TEXT NOT NULL DEFAULT '',
		INVENTURDATUMALT TEXT NOT NULL DEFAULT '0001-01-01',
		INVENTURDATUMNEU TEXT NOT NULL DEFAULT '0001-01-01',
		MAXFUELLMENGE INTEGER NOT NULL DEFAULT 0,
		MINFUELLMENGE INTEGER NOT NULL DEFAULT 0,
		INVENTURFUELLMENGE INTEGER NOT NULL DEFAULT 0,
		AW INTEGER NOT NULL DEFAULT 0
	)`
	if database.Engine == "mysql" {
		siloTable = `CREATE TABLE IF NOT EXISTS SILO (
			ID INTEGER PRIMARY KEY AUTO_INCREMENT,
			SILONUMMER INTEGER NOT NULL DEFAULT 0,
			PERSONENNUMMER INTEGER NOT NULL DEFAULT 0,
			ID_LIEFERANT INTEGER NOT NULL DEFAULT 0,
			BEZEICHNUNG VARCHAR(30) NOT NULL DEFAULT '',
			INVENTURDATUMALT VARCHAR(25) NOT NULL DEFAULT '0001-01-01',
			INVENTURDATUMNEU VARCHAR(25) NOT NULL DEFAULT '0001-01-01',
			MAXFUELLMENGE INTEGER NOT NULL DEFAULT 0,
			MINFUELLMENGE INTEGER NOT NULL DEFAULT 0,
			INVENTURFUELLMENGE INTEGER NOT NULL DEFAULT 0,
			AW INTEGER NOT NULL DEFAULT 0
		)`
	}
	_, err = db.Exec(siloTable)
	if err != nil {
		log.Printf("Fehler beim Erstellen der Tabelle SILO: %v", err)
	}

	// 1d. PERSON Tabelle sicherstellen
	personTable := `CREATE TABLE IF NOT EXISTS PERSON (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		PERSONENNUMMER INTEGER NOT NULL DEFAULT 0,
		NAME TEXT NOT NULL DEFAULT '',
		FIRMA TEXT NOT NULL DEFAULT '',
		STRASSE TEXT NOT NULL DEFAULT '',
		PLZ TEXT NOT NULL DEFAULT '',
		ORT TEXT NOT NULL DEFAULT '',
		TELEFON TEXT NOT NULL DEFAULT '',
		EMAIL TEXT NOT NULL DEFAULT '',
		KZ TEXT NOT NULL DEFAULT ''
	)`
	if database.Engine == "mysql" {
		personTable = `CREATE TABLE IF NOT EXISTS PERSON (
			ID INTEGER PRIMARY KEY AUTO_INCREMENT,
			PERSONENNUMMER INTEGER NOT NULL DEFAULT 0,
			NAME VARCHAR(191) NOT NULL DEFAULT '',
			FIRMA VARCHAR(191) NOT NULL DEFAULT '',
			STRASSE VARCHAR(191) NOT NULL DEFAULT '',
			PLZ VARCHAR(20) NOT NULL DEFAULT '',
			ORT VARCHAR(191) NOT NULL DEFAULT '',
			TELEFON VARCHAR(50) NOT NULL DEFAULT '',
			EMAIL VARCHAR(191) NOT NULL DEFAULT '',
			KZ VARCHAR(10) NOT NULL DEFAULT ''
		)`
	}
	_, err = db.Exec(personTable)

	// 1e. FIRMENPARAMETER Tabelle sicherstellen
	firmenTable := `CREATE TABLE IF NOT EXISTS FIRMENPARAMETER (
		ID INTEGER PRIMARY KEY AUTO_INCREMENT,
		ID_HERDEN INTEGER NOT NULL DEFAULT 0,
		KZ CHAR(1) NOT NULL DEFAULT 'x'
	)`
	if database.Engine == "mysql" {
		firmenTable = `CREATE TABLE IF NOT EXISTS FIRMENPARAMETER (
			ID INTEGER PRIMARY KEY AUTO_INCREMENT,
			ID_HERDEN INTEGER NOT NULL DEFAULT 0,
			KZ CHAR(1) NOT NULL DEFAULT 'x',
			JUMBOS INTEGER NOT NULL DEFAULT 0,
			KLASSENERFASSEN INTEGER NOT NULL DEFAULT 0,
			KLASSEAERFASSEN INTEGER NOT NULL DEFAULT 0,
			KLASSEAERRECHNEN INTEGER NOT NULL DEFAULT 0,
			KLASSEAVERMITTELN INTEGER NOT NULL DEFAULT 0,
			ERFASSESCHMUTZEI INTEGER NOT NULL DEFAULT 0,
			ERFASSEKNICKEI INTEGER NOT NULL DEFAULT 0,
			ERFASSEBRUCHEI INTEGER NOT NULL DEFAULT 0,
			ERFASSEVOLLEI INTEGER NOT NULL DEFAULT 0,
			MASSVOLLEI INTEGER NOT NULL DEFAULT 0,
			AUFTEILUNGGEWICHT INTEGER NOT NULL DEFAULT 0,
			KONTROLLWIEGUNG INTEGER NOT NULL DEFAULT 0,
			ANZAHLKONTROLLW INTEGER NOT NULL DEFAULT 0,
			VERPACKUNGKG DECIMAL(10,3) NOT NULL DEFAULT 0.000,
			AUFTEILUNGALTER INTEGER NOT NULL DEFAULT 0,
			ERFASSEVOLLEIKG INTEGER NOT NULL DEFAULT 0,
			LAUFZEITWOCHEN INTEGER NOT NULL DEFAULT 0,
			ZEITSTEMPEL VARCHAR(25) NOT NULL DEFAULT '',
			SCHLACHTERLOESHENNE DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			PRODUKTIONSDAUER INTEGER NOT NULL DEFAULT 0,
			ID_TABELLEGEWICHT INTEGER NOT NULL DEFAULT 0,
			ID_TABELLEALTER INTEGER NOT NULL DEFAULT 0,
			LEGEBEGINN_LW INTEGER NOT NULL DEFAULT 0,
			VERLUSTEBEIBUCHUNG INTEGER NOT NULL DEFAULT 0,
			LAGERBUCHUNGBEIBUCHUNG INTEGER NOT NULL DEFAULT 0,
			MAXTAGEVERMITTELN INTEGER NOT NULL DEFAULT 0,
			CHARGEJUMBOS INTEGER NOT NULL DEFAULT 0,
			CHARGEXL INTEGER NOT NULL DEFAULT 0,
			CHARGEMEDIUM INTEGER NOT NULL DEFAULT 0,
			CHARGESMALL INTEGER NOT NULL DEFAULT 0,
			CHARGELARGE INTEGER NOT NULL DEFAULT 0,
			CHARGEVOLLEI INTEGER NOT NULL DEFAULT 0,
			CHARGEPREFIXFIRMA VARCHAR(50) NOT NULL DEFAULT '',
			CHARGEPREFIXHERDENNUMMER INTEGER NOT NULL DEFAULT 0,
			CHARGEDATUM INTEGER NOT NULL DEFAULT 0,
			CHARGELAGERNUMMER INTEGER NOT NULL DEFAULT 0,
			CHARGETRENNUNG VARCHAR(10) NOT NULL DEFAULT '',
			BEIVERMITTELNDATUMAKTUELL INTEGER NOT NULL DEFAULT 0,
			PSEUDOLAGER INTEGER NOT NULL DEFAULT 0,
			BIO INTEGER NOT NULL DEFAULT 0,
			HALTUNGSTYP VARCHAR(10) NOT NULL DEFAULT '',
			BIOAUFSCHLAG DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			AW INTEGER NOT NULL DEFAULT 0
		)`
	}
	_, err = db.Exec(firmenTable)

	// 1f. AKTIONEN Tabelle sicherstellen
	aktionenTable := `CREATE TABLE IF NOT EXISTS AKTIONEN (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		AKTIONEN_KZ CHAR (1),
		ID_USER INTEGER DEFAULT (0),
		AKTIONSDATUM TEXT DEFAULT ('0001-01-01'),
		BEZEICHNUNG TEXT,
		INTERVALL_TAGE INTEGER DEFAULT (0),
		ANZAHL_INTERVALLE INTEGER DEFAULT (0),
		ERLEDIGT INTEGER DEFAULT (0),
		ID_USER_ERLEDIGT INTEGER DEFAULT (0),
		ERLEDIGT_AM TEXT DEFAULT ''
	)`
	if database.Engine == "mysql" {
		aktionenTable = `CREATE TABLE IF NOT EXISTS AKTIONEN (
			ID INTEGER PRIMARY KEY AUTO_INCREMENT,
			AKTIONEN_KZ CHAR (1),
			ID_USER INTEGER DEFAULT (0),
			AKTIONSDATUM VARCHAR(25) DEFAULT ('0001-01-01'),
			BEZEICHNUNG TEXT,
			INTERVALL_TAGE INTEGER DEFAULT (0),
			ANZAHL_INTERVALLE INTEGER DEFAULT (0),
			ERLEDIGT INTEGER DEFAULT (0),
			ID_USER_ERLEDIGT INTEGER DEFAULT (0),
			ERLEDIGT_AM VARCHAR(25) DEFAULT ''
		)`
	}
	_, err = db.Exec(aktionenTable)
	if err != nil {
		log.Printf("Fehler beim Erstellen der Tabelle AKTIONEN: %v", err)
	}

	// 2. Fehlende Spalten hinzufügen
	cols := []struct {
		table    string
		col      string
		def      string
		mysqlDef string
	}{
		{"VERKAUF", "CHARGE", "TEXT NOT NULL DEFAULT ''", "VARCHAR(255) NOT NULL DEFAULT ''"},
		{"VERKAUF", "RABATTPROZENT", "REAL NOT NULL DEFAULT 0.0", "DECIMAL(5, 2) NOT NULL DEFAULT 0.0"},
		{"EILAGERBUCHUNG", "SCHMUTZ", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"EILAGERBUCHUNG", "KNICKEIER", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"EILAGERBUCHUNG", "BRUCHEIER", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"EILAGERBUCHUNG", "BUCHUNGSTYP", "TEXT NOT NULL DEFAULT 'E'", "VARCHAR(10) NOT NULL DEFAULT 'E'"},
		{"EILAGERBUCHUNG", "CHARGE", "TEXT NOT NULL DEFAULT ''", "VARCHAR(255) NOT NULL DEFAULT ''"},
		{"EILAGERBUCHUNG", "KZ_LAGER", "TEXT NOT NULL DEFAULT ''", "VARCHAR(10) NOT NULL DEFAULT ''"},
		{"EILAGERBUCHUNG", "ID_FREMDEBUCHUNG", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"EILAGERBUCHUNG", "VERKAUF", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"EILAGERBUCHUNG", "ID_LAGERPLATZ", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"TIERBEWEGUNGEN", "HERDENNUMMER", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"TIERBEWEGUNGEN", "ID_BUCHUNG", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"TIERBEWEGUNGEN", "BEWEGUNGEN", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"TIERBEWEGUNGEN", "ID_HERDEN_VON", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"TIERBEWEGUNGEN", "ID_HERDEN_NACH", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"AKTIONEN", "ANZAHL_INTERVALLE", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"AKTIONEN", "ID_USER_ERLEDIGT", "INTEGER NOT NULL DEFAULT 0", "INTEGER NOT NULL DEFAULT 0"},
		{"AKTIONEN", "ERLEDIGT_AM", "TEXT NOT NULL DEFAULT ''", "VARCHAR(25) NOT NULL DEFAULT ''"},
	}

	for _, c := range cols {
		exists := false
		if database.Engine == "sqlite" {
			rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", c.table))
			if err == nil {
				for rows.Next() {
					var cid int
					var name, dtype string
					var notnull, pk int
					var dflt_value interface{}
					if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err == nil {
						if strings.EqualFold(name, c.col) {
							exists = true
							break
						}
					}
				}
				rows.Close()
			}
		} else {
			// MySQL Check
			err := db.QueryRow("SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_NAME = ? AND COLUMN_NAME = ? AND TABLE_SCHEMA = DATABASE()", c.table, c.col).Scan(&exists)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("Fehler beim Prüfen der Spalte %s.%s: %v", c.table, c.col, err)
			}
			if err == nil {
				exists = true
			}
		}

		if !exists {
			log.Printf("Migration: Füge Spalte %s zu Tabelle %s hinzu...", c.col, c.table)
			def := c.def
			if database.Engine == "mysql" {
				def = c.mysqlDef
			}
			query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", c.table, c.col, def)
			if _, err := db.Exec(query); err != nil {
				log.Printf("FEHLER bei Migration von %s.%s: %v", c.table, c.col, err)
			}
		}
	}
	log.Println("Schema-Prüfung abgeschlossen.")

	// 3. Basis-Daten synchronisieren, falls MariaDB leer ist
	if database.Engine == "mysql" {
		syncDataFromSQLite(database)
	}

	// 4. Backfill SILONR for bookings that have 0
	log.Println("Migration: Repariere SILONR = 0 in BUCHUNG...")
	repairRes, err := db.Exec(`
		UPDATE BUCHUNG
		SET SILONR = COALESCE((
			SELECT S.SILONUMMER
			FROM HERDEN H
			JOIN SILO S ON H.ID_SILO = S.ID
			WHERE H.ID = BUCHUNG.ID_HERDEN
		), 0)
		WHERE SILONR = 0`)
	if err != nil {
		log.Printf("Fehler beim Reparieren von SILONR in BUCHUNG: %v", err)
	} else {
		if rows, err := repairRes.RowsAffected(); err == nil && rows > 0 {
			log.Printf("Migration: %d Buchungen erfolgreich mit korrekter SILONR aktualisiert.", rows)
		}
	}
	migrateTranslations(database)
}

func migrateTranslations(database *wailsdb.DB) {
	db := database.SQL
	log.Println("[MIGRATION] Checking en/it translations in UEBERSETZUNGEN and TRANSLATEFELDNAMEN...")

	fieldTranslations := map[string][2]string{
		"ID_HERDEN":               {"Herd ID", "ID Gregge"},
		"BUCHUNGSDATUM":           {"Booking Date", "Data Registrazione"},
		"GEWICHTPROBE":            {"Weight Sample", "Campione Peso"},
		"KONTROLLGEWICHT":         {"Control Weight", "Peso di Controllo"},
		"KLASSEA":                 {"Class A", "Classe A"},
		"VERLUSTE":                {"Losses", "Perdite"},
		"EIMASSE":                 {"Egg Mass", "Massa delle Uova"},
		"SCHMUTZ":                 {"Dirty", "Sporco"},
		"KNICKEIER":               {"Cracked Eggs", "Uova Incrinate"},
		"VOLLEI":                  {"Whole Egg", "Uovo Intero"},
		"BRUCHEIER":               {"Broken Eggs", "Uova Rotte"},
		"TIERBESTAND":             {"Animal Stock", "Numero Animali"},
		"ID_EITABELLE":            {"Egg Table ID", "ID Tabella Uova"},
		"ID_DGEWICHTTAB":          {"Weight Table ID", "ID Tabella Peso"},
		"FUTTERKTAG":              {"Feed/Day (g)", "Mangime/Giorno (g)"},
		"SILONR":                  {"Silo No.", "Silo N."},
		"KL6":                     {"Jumbo", "Jumbo"},
		"VERMITTELTAM":            {"Brokered On", "Intermediato il"},
		"SMALL":                   {"Size S", "Dimensione S"},
		"LARGE":                   {"Size L", "Dimensione L"},
		"MEDIUM":                  {"Size M", "Dimensione M"},
		"XL":                      {"Size XL", "Dimensione XL"},
		"ZEITSTEMPEL":             {"Timestamp", "Timestamp"},
		"DGEWICHTEI":              {"Avg Egg Weight", "Peso Medio Uova"},
		"AW":                      {"AW", "AW"},
		"VERMITTELT":              {"Brokered", "Intermediato"},
		"FUTTERVERBRAUCHTIER":     {"Feed Consumption/Animal", "Consumo Mangime/Animale"},
		"LAGERNUMMER":             {"Storage Number", "Numero di Magazzino"},
		"KZ":                      {"Code", "Codice"},
		"BEZEICHNUNG":             {"Designation", "Designazione"},
		"LETZTE_BUCHUNG":          {"Last Booking", "Ultima Registrazione"},
		"JUMBOS":                  {"Jumbos", "Jumbo"},
		"VOLLEIKG":                {"Whole Egg (kg)", "Uovo Intero (kg)"},
		"KLASSE6":                 {"Class 6", "Classe 6"},
		"KLASSE7":                 {"Class 7", "Classe 7"},
		"ID_FREMDESLAGER":         {"External Warehouse ID", "ID Magazzino Esterno"},
		"ID_BUCHUNG":              {"Booking ID", "ID Registrazione"},
		"ID_EILAGER":              {"Egg Storage ID", "ID Deposito Uova"},
		"SCHMUTZ_KG":              {"Dirty (kg)", "Sporco (kg)"},
		"KNICKEIER_KG":            {"Cracked (kg)", "Incrinato (kg)"},
		"BRUCHEIER_KG":            {"Broken (kg)", "Rotto (kg)"},
		"BUCHUNGSTYP":             {"Booking Type", "Tipo di Registrazione"},
		"CHARGE":                  {"Batch", "Lotto"},
		"KZ_LAGER":                {"Warehouse Code", "Codice Magazzino"},
		"ID_FREMDEBUCHUNG":        {"External Booking ID", "ID Registrazione Esterna"},
		"VERKAUF":                 {"Sale", "Vendita"},
		"ID_HERDEN_NEU":           {"New Herd ID", "ID Nuovo Gregge"},
		"JUMBOSERFASSEN":          {"Record Jumbos", "Registra Jumbo"},
		"KLASSENERFASSEN":         {"Record Classes", "Registra Classi"},
		"KLASSEAERFASSEN":         {"Record Class A", "Registra Classe A"},
		"KLASSEAERRECHNEN":        {"Calculate Class A", "Calcola Classe A"},
		"KLASSEAVERMITTELN":       {"Broker Class A", "Intermedia Classe A"},
		"ERFASSESCHMUTZEI":        {"Record Dirty Eggs", "Registra Uova Sporche"},
		"ERFASSEKNICKEI":          {"Record Cracked Eggs", "Registra Uova Incrinate"},
		"ERFASSEBRUCHEI":          {"Record Broken Eggs", "Registra Uova Rotte"},
		"ERFASSEVOLLEI":           {"Record Whole Eggs", "Registra Uova Intere"},
		"MASSVOLLEI":              {"Measure Whole Eggs", "Misura Uova Intere"},
		"AUFTEILUNGGEWICHT":       {"Division by Weight", "Divisione per Peso"},
		"KONTROLLWIEGUNG":         {"Control Weighing", "Pesatura di Controllo"},
		"ANZAHLKONTROLLW":         {"Number of Control Weighings", "Numero Pesature di Controllo"},
		"VERPACKUNGKG":            {"Packaging (kg)", "Imballaggio (kg)"},
		"AUFTEILUNGALTER":         {"Division by Age", "Divisione per Età"},
		"ERFASSEVOLLEIKG":         {"Record Whole Eggs (kg)", "Registra Uova Intere (kg)"},
		"LAUFZEITWOCHEN":          {"Run Time Weeks", "Settimane di Durata"},
		"SCHLACHTERLOESHENNE":     {"Hen Slaughter Proceeds", "Ricavo Macellazione Gallina"},
		"PRODUKTIONSDAUER":        {"Production Duration", "Durata Produzione"},
		"ID_TABELLEGEWICHT":       {"Weight Table ID", "ID Tabella Peso"},
		"ID_TABELLEALTER":         {"Age Table ID", "ID Tabella Età"},
		"LEGEBEGINN_LW":           {"Start of Laying LW", "Inizio Deposizione LW"},
		"VERLUSTEBEIBUCHUNG":      {"Losses on Booking", "Perdite alla Registrazione"},
		"LAGERBUCHUNGBEIBUCHUNG":  {"Stock Booking on Booking", "Registrazione Magazzino alla Registrazione"},
		"MAXTAGEVERMITTELN":       {"Max Days to Broker", "Giorni Max per Intermediare"},
		"CHARGEJUMBOS":            {"Batch Jumbos", "Lotto Jumbo"},
		"CHARGEXL":                {"Batch XL", "Lotto XL"},
		"CHARGEMEDIUM":            {"Batch Medium", "Lotto Medium"},
		"CHARGESMALL":             {"Batch Small", "Lotto Small"},
		"CHARGELARGE":             {"Batch Large", "Lotto Large"},
		"CHARGEVOLLEI":            {"Batch Whole Egg", "Lotto Uovo Intero"},
		"CHARGEPREFIXFIRMA":       {"Batch Prefix Company", "Prefisso Lotto Azienda"},
		"CHARGEPREFIXHERDENNUMMER": {"Batch Prefix Herd No.", "Prefisso Lotto N. Gregge"},
		"CHARGEDATUM":             {"Batch Date", "Data Lotto"},
		"CHARGELAGERNUMMER":       {"Batch Storage No.", "Prefisso Lotto N. Magazzino"},
		"CHARGETRENNUNG":          {"Batch Separator", "Separatore Lotto"},
		"BEIVERMITTELNDATUMAKTUELL": {"Use Current Date on Brokering", "Usa Data Corrente per Intermediazione"},
		"PSEUDOLAGER":             {"Pseudo Storage", "Deposito Pseudo"},
		"BIO":                     {"Bio / Organic", "Biologico"},
		"HALTUNGSTYP":             {"Housing Type", "Tipo di Allevamento"},
		"BIOAUFSCHLAG":            {"Bio Surcharge", "Sovrapprezzo Biologico"},
		"FUTTERINVENTUR":          {"Feed Inventory", "Inventario Mangime"},
		"ID_SILO":                 {"Silo ID", "ID Silo"},
		"SILONUMMER":              {"Silo Number", "Numero Silo"},
		"HERDENR":                 {"Herd No.", "N. Gregge"},
		"ID_PERSON":               {"Person ID", "ID Persona"},
		"LIEFERDATUM":             {"Delivery Date", "Data Consegna"},
		"LIEFERMENGE":             {"Delivery Quantity", "Quantità Consegnata"},
		"PREISDT":                 {"Price/dt", "Prezzo/dt"},
		"RABATTPROZ":              {"Discount %", "Sconto %"},
		"NETTO":                   {"Net", "Netto"},
		"BRUTTO":                  {"Gross", "Lordo"},
		"MWSTPROZ":                {"VAT %", "IVA %"},
		"MWSTKZ":                  {"VAT Code", "Codice IVA"},
		"DATUM":                   {"Date", "Data"},
		"ID_FUTTERSORTEN":         {"Feed Type ID", "ID Tipo Mangime"},
		"ID_STALL":                {"Stall ID", "ID Capannone"},
		"ID_EILAGER_NEU":          {"New Egg Storage ID", "ID Nuovo Deposito Uova"},
		"ID_ZUECHTER":             {"Breeder ID", "ID Allevatore"},
		"ID_RASSE":                {"Breed ID", "ID Razza"},
		"ANFANGSKOSTEN":           {"Initial Cost", "Costo Iniziale"},
		"ANFANGSBESTAND":          {"Initial Stock", "Popolazione Iniziale"},
		"EINSTALLDATUM":           {"Housing Date", "Data Accasamento"},
		"LEGEDATUM":               {"Laying Date", "Data Deposizione"},
		"EINSTALLKOSTEN":          {"Housing Costs", "Costo Accasamento"},
		"AKTIV":                   {"Active", "Attivo"},
		"ALLEBUCHUNGENMITDATUM":   {"All Bookings with Date", "Tutte le Registrazioni con Data"},
		"PREIS":                   {"Price", "Prezzo"},
		"KLASSE":                  {"Class", "Classe"},
		"VON":                     {"From", "Da"},
		"BIS":                     {"To", "A"},
		"BEMERKUNG":               {"Remark", "Nota"},
		"ID_KOSTENTABKOPF":        {"Cost Table Header ID", "ID Intestazione Tabella Costi"},
		"KOSTENTYP":               {"Cost Type", "Tipo di Costo"},
		"TAGE":                    {"Days", "Giorni"},
		"SCHLACHTERLOES":          {"Slaughter Proceeds", "Ricavo Macellazione"},
		"PRODDAUERGEPLANT":        {"Planned Prod Duration", "Durata Prod Pianificata"},
		"GEBAEUDEWERT":            {"Building Value", "Valore Edificio"},
		"ABSCHREIBUNG_G":          {"Depreciation Building", "Ammortamento Edificio"},
		"GERAETEWERT":             {"Equipment Value", "Valore Attrezzatura"},
		"ABSCHREIBUNG_R":          {"Depreciation Equipment", "Ammortamento Attrezzatura"},
		"ALTERINWOCHEN":           {"Age (Weeks)", "Età (Settimane)"},
		"EIZAHLKUM":               {"Cumulative Egg Count", "Conteggio Uova Cumulativo"},
		"LEGERATEAH":              {"Laying Rate AH", "Tasso Deposizione AH"},
		"LEGERATEDH":              {"Laying Rate DH", "Tasso Deposizione DH"},
		"EIGEWICHTWO":             {"Egg Weight Week", "Peso Uovo Settimana"},
		"EIGEWICHTKUM":            {"Cumulative Egg Weight", "Peso Uovo Cumulativo"},
		"EIMASSEWO":               {"Egg Mass Week", "Massa Uova Settimana"},
		"EIMASSEKUM":              {"Cumulative Egg Mass", "Massa Uova Cumulativo"},
		"PROZENT":                 {"Percent", "Percentuale"},
		"KONTO":                   {"Account", "Conto"},
		"ID_TEXTE":                {"Text ID", "ID Testo"},
		"ID_ANREDE":               {"Salutation ID", "ID Appellativo"},
		"POSTFACH":                {"PO Box", "Casella Postale"},
		"STRASSE":                 {"Street", "Via"},
		"PLZ":                     {"Zip Code", "CAP"},
		"TELEFON":                 {"Phone", "Telefono"},
		"MOBILTELEPHON":           {"Mobile", "Cellulare"},
		"EMAIL":                   {"Email", "E-mail"},
		"EMAIL2":                  {"Secondary Email", "E-mail Secondaria"},
		"HOMEPAGE":                {"Homepage", "Sito Web"},
		"STALLNUMMER":             {"Stall Number", "Numero Capannone"},
		"TABELLENTYP":             {"Table Type", "Tipo Tabella"},
		"ANLAGEDATUM":             {"Creation Date", "Data Creazione"},
		"TEXT_TYP_KZ":             {"Text Type Code", "Codice Tipo Testo"},
		"SYSTEM_KZ":               {"System Code", "Codice Sistema"},
		"BETREFF":                 {"Subject", "Oggetto"},
		"INHALT":                  {"Content", "Contenuto"},
		"HERDENNUMMER_VON":        {"From Herd No.", "Dal N. Gregge"},
		"HERDENNUMMER_NACH":       {"To Herd No.", "Al N. Gregge"},
		"GRUND":                   {"Reason", "Motivo"},
		"EIGEWICHT":               {"Egg Weight", "Peso Uovo"},
		"KLASSE1":                 {"Class 1", "Classe 1"},
		"KLASSE2":                 {"Class 2", "Classe 2"},
		"KLASSE3":                 {"Class 3", "Classe 3"},
		"KLASSE4":                 {"Class 4", "Classe 4"},
		"KLASSE5":                 {"Class 5", "Classe 5"},
		"EIERKLASSE":              {"Egg Class", "Classe Uova"},
		"GEWICHT_VON":             {"Weight From", "Peso Da"},
		"GEWICHT_BIS":             {"Weight To", "Peso A"},
		"PREIS_VON":               {"Price From", "Prezzo Da"},
		"PREIS_BIS":               {"Price To", "Prezzo A"},
		"BESCHREIBUNG":            {"Description", "Descrizione"},
		"KATEGORIE_KZ":            {"Category Code", "Codice Categoria"},
		"GRUPPEN_KZ":              {"Group Code", "Codice Gruppo"},
		"TEMPLATE_NAME":           {"Template Name", "Nome Modello"},
		"GROUP_FIELD":             {"Group Field", "Campo Gruppo"},
		"ROWS_PER_PAGE":           {"Rows Per Page", "Righe Per Pagina"},
		"PAGE_ORIENTATION":        {"Page Orientation", "Orientamento Pagina"},
		"SHOW_MASTER_GRID":        {"Show Master Grid", "Mostra Griglia Principale"},
		"SHOW_DETAIL_GRID":        {"Show Detail Grid", "Mostra Griglia Dettaglio"},
		"ROOT_KZ":                 {"Root Code", "Codice Root"},
		"SUMMENZEILE":             {"Summary Row", "Riga Somma"},
		"IST_SUMMENZEILE":         {"Is Summary Row", "È Riga Somma"},
		"FELDNAME":                {"Field Name", "Nome Campo"},
		"NAMEINDB":                {"Name in DB", "Nome nel DB"},
		"USERNAME":                {"Username", "Nome utente"},
		"KLARNAME":                {"Real Name", "Nome reale"},
		"ID_BENUTZER_PROFILE":     {"User Profile ID", "ID Profilo Utente"},
		"TVNAME":                  {"TV Name", "Nome TV"},
		"AKTIONEN_KZ":             {"Action Code", "Codice Azione"},
		"AKTIONSDATUM":            {"Action Date", "Data Azione"},
		"INTERVALL_TAGE":          {"Interval (Days)", "Intervallo (Giorni)"},
		"ANZAHL_INTERVALLE":       {"Number of Intervals", "Numero di Intervalli"},
		"ERLEDIGT_AM":             {"Completed On", "Completato il"},
		"INVENTURDATUMNEU":         {"New Inventory Date", "Nuova Data Inventario"},
		"INVENTURDATUMALT":         {"Old Inventory Date", "Vecchia Data Inventario"},
		"CHARGESMALL_OLD":         {"Batch Small (Old)", "Lotto Small (Vecchio)"},
		"F_BACKUP_ERSTELLEN":       {"Create Backup Permission", "Permesso Crea Backup"},
		"PARAM_DEF":               {"Default Parameter", "Parametro Predefinito"},
		"LIEFERANTNUMMER":         {"Supplier Number", "Numero Fornitore"},
		"TYP_KZ":                  {"Type Code", "Codice Tipo"},
		"LINK_LOGIC":              {"Link Logic", "Logica di Collegamento"},
		"ID_TEXT":                 {"Text ID", "ID Testo"},
		"PERSONENNUMMER":           {"Person Number", "Numero Persona"},
		"ID_FELD_KATALOG":         {"Field Catalog ID", "ID Catalogo Campi"},
		"F_PERSONEN_VERWALTEN":     {"Manage Persons Permission", "Permesso Gestisci Persone"},
		"F_TEXTE_VERWALTEN":       {"Manage Texts Permission", "Permesso Gestisci Testi"},
		"F_KOSTEN_VERWALTEN":       {"Manage Costs Permission", "Permesso Gestisci Costi"},
		"F_PARAMETER_EDITIEREN":   {"Edit Parameters Permission", "Permesso Modifica Parametri"},
		"F_BUCHUNGEN_ERFASSEN":     {"Record Bookings Permission", "Permesso Registra Movimenti"},
		"F_SYSTEM_VERWALTUNG":     {"System Administration Permission", "Permesso Amministrazione Sistema"},
		"F_AUSWERTUNGEN_ANZEIGEN": {"Show Reports Permission", "Permesso Mostra Report"},
		"SQLSTATEMENT":            {"SQL Statement", "Istruzione SQL"},
		"DETAIL_SQL":              {"Detail SQL", "SQL Dettaglio"},
	}

	texteTranslations := map[string][2]string{
		"Lüftung wurde vorübergehend vom Elektriz": {
			"Ventilation was temporarily cut off from the electricity grid",
			"La ventilazione è stata temporaneamente interrotta dalla rete elettrica",
		},
		"Impfung mit Hitcher gegen Newcastle": {
			"Vaccination with Hitcher against Newcastle disease",
			"Vaccinazione con Hitcher contro la malattia di Newcastle",
		},
		"Wasserversorgung war ausgefallen": {
			"Water supply had failed",
			"L'approvvigionamento idrico era interrotto",
		},
		"Eiersammeln ausgefallen wegen Fiertag": {
			"Egg collection cancelled due to public holiday",
			"Raccolta delle uova annullata a causa di un giorno festivo",
		},
		"Statistisches Landesamt": {
			"State Statistical Office",
			"Ufficio statistico statale",
		},
		"Bauart des Stalles": {
			"Barn construction type",
			"Tipo di costruzione del capannone",
		},
		"Lüftungs - System": {
			"Ventilation system",
			"Sistema di ventilazione",
		},
		"Zuechter": {
			"Breeder",
			"Allevatore",
		},
		"Kunde": {
			"Customer",
			"Cliente",
		},
		"Lieferant": {
			"Supplier",
			"Fornitore",
		},
		"Vertreter": {
			"Representative",
			"Rappresentante",
		},
		"Bodenhaltung": {
			"Barn rearing",
			"Allevamento a terra",
		},
		"Käfighaltung": {
			"Cage rearing",
			"Allevamento in gabbia",
		},
		"Freiland Haltung": {
			"Free range rearing",
			"Allevamento all'aperto",
		},
		"Herr": {
			"Mr.",
			"Sig.",
		},
		"Frau": {
			"Mrs./Ms.",
			"Sig.ra",
		},
		"Sehr geehrter Herr": {
			"Dear Mr.",
			"Gentile Signore",
		},
		"Sehr geehrte Frau": {
			"Dear Mrs./Ms.",
			"Gentile Signora",
		},
		"Sehr geehrte Damen und Herren": {
			"Dear Sir or Madam",
			"Gentili Signore e Signori",
		},
		"zusätzlich eingestallt": {
			"Additionally housed",
			"Accasato in aggiunta",
		},
		"Zukauf/zugestallt": {
			"Purchased/Housed",
			"Acquistato/Accasato",
		},
		"wurde von einer in eine andere Herde übertragen": {
			"was transferred from one herd to another",
			"è stato trasferito da un gregge all'altro",
		},
		"Umgebucht": {
			"Rebooked",
			"Stornato/Trasferito",
		},
		"Kanibalismus": {
			"Cannibalism",
			"Cannibalismo",
		},
		"Stromausfall bedingter Verlust durch Klimaausfall": {
			"Loss due to climate system failure caused by power outage",
			"Perdita dovuta a guasto del sistema di climatizzazione per blackout",
		},
		"Klimaanlage ausgefallen": {
			"Air conditioning failed",
			"Aria condizionata guasta",
		},
		"War nicht mehr festzustellen": {
			"Could no longer be determined",
			"Non è stato più possibile stabilirlo",
		},
		"Allgemeine Verluste": {
			"General losses",
			"Perdite generali",
		},
		"Allgemeiner Verbrauch": {
			"General consumption",
			"Consumo generale",
		},
		"Private Adressen Freunde und Bekannte": {
			"Private addresses of friends and acquaintances",
			"Indirizzi privati di amici e conoscenti",
		},
		"Bekannte und Freunde": {
			"Friends and acquaintances",
			"Amici e conoscenti",
		},
		"Bezieht sich auf den Anwender des Programms und kann nur über die Firmenverwaltung geändert werden": {
			"Refers to the program user and can only be changed via company administration",
			"Si riferisce all'utente del programma e può essere modificato solo tramite l'amministrazione aziendale",
		},
		"Anwender": {
			"User",
			"Utente",
		},
		"Bezieht sich auf die Firmenverwaltung sperrt den Firmensatz": {
			"Refers to company administration, locks the company record",
			"Si riferisce all'amministrazione aziendale, blocca il record dell'azienda",
		},
		"Firmenstamm": {
			"Company details",
			"Dati aziendali",
		},
		"Intensive Auslaufhaltung": {
			"Intensive free range housing",
			"Allevamento all'aperto intensivo",
		},
		"Aufteilung nach Alter": {
			"Division by age",
			"Divisione per età",
		},
		"Aufteilung nach Gewicht": {
			"Division by weight",
			"Divisione per peso",
		},
		"Kommt aus den Legeleistungen": {
			"Comes from laying performance",
			"Proviene dalle prestazioni di deposizione",
		},
		"Eingänge": {
			"Receipts/Inputs",
			"Entrate/Ingressi",
		},
		"Entnahmen": {
			"Withdrawals",
			"Prelievi",
		},
		"Ausgänge": {
			"Outputs/Issues",
			"Uscite",
		},
		"Umgebucht in anderes Lager": {
			"Transferred to another warehouse",
			"Trasferito in un altro deposito",
		},
		"Umbuchungen": {
			"Transfers/Rebookings",
			"Trasferimenti",
		},
		"Sind nicht Teil des Lagerbestandes": {
			"Are not part of the inventory",
			"Non fanno parte dell'inventario",
		},
		"Buchen in Pseudo Lager": {
			"Book in pseudo warehouse",
			"Registra in pseudo deposito",
		},
		"Eingänge aus anderem Lager": {
			"Receipts from other warehouse",
			"Entrate da un altro deposito",
		},
		"Rückbuchungen": {
			"Reversals/Returns",
			"Storni",
		},
		"Hofladen": {
			"Farm shop",
			"Negozio dell'azienda agricola",
		},
		"Eier könnten nicht eingelagert werden": {
			"Eggs could not be stored",
			"Non è stato possibile immagazzinare le uova",
		},
		"Vernichtet": {
			"Destroyed",
			"Distrutto",
		},
		"Eier wurden zu persönlichen Verbrauch entnommen": {
			"Eggs were taken for personal consumption",
			"Le uova sono state prelevate per il consumo personale",
		},
		"Privat Entnahmen": {
			"Private withdrawals",
			"Prelievi privati",
		},
		"Hier werden die zum Verkauf bestimmten Eier eingelagert. Sie bestimmen auch den Lagerbestand": {
			"Here the eggs intended for sale are stored. They also determine the inventory",
			"Qui vengono immagazzinate le uova destinate alla vendita. Esse determinano anche l'inventario",
		},
		"Standard Eierlager": {
			"Standard egg storage",
			"Deposito uova standard",
		},
		"Bekommt Eier aus Produktion": {
			"Receives eggs from production",
			"Riceve uova dalla produzione",
		},
		"Eier zum Privaten Verbrauch genutzt": {
			"Eggs used for private consumption",
			"Uova destinate al consumo privato",
		},
		"Eigene Verwendung": {
			"Own use",
			"Uso proprio",
		},
		"Eier wurden verfüttert": {
			"Eggs were fed to animals",
			"Le uova sono state destinate all'alimentazione degli animali",
		},
		"Tierfutter": {
			"Animal feed",
			"Mangime per animali",
		},
		"Eierläger zur Verkauf": {
			"Egg storages for sale",
			"Depositi uova per la vendita",
		},
		"Pseudolager für Ei Verkäufe": {
			"Pseudo warehouse for egg sales",
			"Pseudo deposito per le vendite di uova",
		},
		"Verkäufe aus Eilager": {
			"Sales from egg storage",
			"Vendite da deposito uova",
		},
		"Nur für Verkäufe aus Eilager": {
			"Only for sales from egg storage",
			"Solo per vendite da deposito uova",
		},
		"Behörden und Ämter": {
			"Authorities and offices",
			"Autorità e uffici",
		},
		"Behörde": {
			"Authority",
			"Autorità",
		},
		"Tiere zugestallt": {
			"Animals housed",
			"Animali accasati",
		},
		"Zustallung": {
			"Housing",
			"Accasamento",
		},
		"Diese Aktion wird nach dem Intervall in Tagen zur Wiederholung aktiviert": {
			"This action is activated for repetition after the interval in days",
			"Questa azione si attiva per la ripetizione dopo l'intervallo in giorni",
		},
		"Wiederkehrend": {
			"Recurring",
			"Ricorrente",
		},
		"Diese Aktion führt zur Benachrichtigung der/des User(s)": {
			"This action leads to the notification of the user(s)",
			"Questa azione comporta la notifica degli utenti",
		},
		"Benachrichtigung": {
			"Notification",
			"Notifica",
		},
		"Alle Aktionen die sonst angefallen sind": {
			"All other actions that occurred",
			"Tutte le altre azioni che si sono verificate",
		},
		"Allgemeine Aktion": {
			"General action",
			"Azione generale",
		},
		"Firma": {
			"Company",
			"Azienda",
		},
		"Verkauf": {
			"Sale",
			"Vendita",
		},
		"Kofu la 2 opt": {
			"Kofu la 2 opt",
			"Kofu la 2 opt",
		},
		"kofu la 2": {
			"Kofu la 2",
			"Kofu la 2",
		},
	}

	// 1. Process TRANSLATEFELDNAMEN
	rows, err := db.Query("SELECT ID, FELDNAME FROM FELD_KATALOG")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int64
			var feldname string
			if errScan := rows.Scan(&id, &feldname); errScan == nil {
				upperFeld := strings.ToUpper(feldname)
				trans, exists := fieldTranslations[upperFeld]
				enVal := toCamelCase(feldname)
				itVal := enVal
				if exists {
					enVal = trans[0]
					itVal = trans[1]
				}

				// Check and insert 'en'
				var count int
				db.QueryRow("SELECT COUNT(*) FROM TRANSLATEFELDNAMEN WHERE ID_FELD_KATALOG = ? AND SPRACHE_KZ = 'en'", id).Scan(&count)
				if count == 0 {
					db.Exec("INSERT INTO TRANSLATEFELDNAMEN (ID_FELD_KATALOG, SPRACHE_KZ, BETREFF, INHALT) VALUES (?, 'en', ?, ?)", id, enVal, enVal)
				}

				// Check and insert 'it'
				db.QueryRow("SELECT COUNT(*) FROM TRANSLATEFELDNAMEN WHERE ID_FELD_KATALOG = ? AND SPRACHE_KZ = 'it'", id).Scan(&count)
				if count == 0 {
					db.Exec("INSERT INTO TRANSLATEFELDNAMEN (ID_FELD_KATALOG, SPRACHE_KZ, BETREFF, INHALT) VALUES (?, 'it', ?, ?)", id, itVal, itVal)
				}
			}
		}
	}

	// 2. Process UEBERSETZUNGEN
	rowsTexte, err := db.Query("SELECT ID_TEXTE, BETREFF, INHALT FROM UEBERSETZUNGEN WHERE SPRACHE_KZ = 'de'")
	if err == nil {
		defer rowsTexte.Close()
		for rowsTexte.Next() {
			var idText int64
			var betreff, inhalt string
			if errScan := rowsTexte.Scan(&idText, &betreff, &inhalt); errScan == nil {
				enBetreff := betreff
				itBetreff := betreff
				if trans, found := texteTranslations[betreff]; found {
					enBetreff = trans[0]
					itBetreff = trans[1]
				} else if trans, found := texteTranslations[inhalt]; found {
					enBetreff = trans[0]
					itBetreff = trans[1]
				}

				enInhalt := inhalt
				itInhalt := inhalt
				if trans, found := texteTranslations[inhalt]; found {
					enInhalt = trans[0]
					itInhalt = trans[1]
				} else if trans, found := texteTranslations[betreff]; found {
					enInhalt = trans[0]
					itInhalt = trans[1]
				}

				// Check and insert 'en'
				var count int
				db.QueryRow("SELECT COUNT(*) FROM UEBERSETZUNGEN WHERE ID_TEXTE = ? AND SPRACHE_KZ = 'en'", idText).Scan(&count)
				if count == 0 {
					db.Exec("INSERT INTO UEBERSETZUNGEN (ID_TEXTE, SPRACHE_KZ, BETREFF, INHALT) VALUES (?, 'en', ?, ?)", idText, enBetreff, enInhalt)
				}

				// Check and insert 'it'
				db.QueryRow("SELECT COUNT(*) FROM UEBERSETZUNGEN WHERE ID_TEXTE = ? AND SPRACHE_KZ = 'it'", idText).Scan(&count)
				if count == 0 {
					db.Exec("INSERT INTO UEBERSETZUNGEN (ID_TEXTE, SPRACHE_KZ, BETREFF, INHALT) VALUES (?, 'it', ?, ?)", idText, itBetreff, itInhalt)
				}
			}
		}
	}
}

func syncDataFromSQLite(database *wailsdb.DB) {
	// SQLite öffnen (Suche nach HuhnLite.db in verschiedenen Pfaden)
	sqlitePath := "HuhnLite.db"
	if _, err := os.Stat(sqlitePath); os.IsNotExist(err) {
		// Pfad relativ zur Executable versuchen
		if execPath, err := os.Executable(); err == nil {
			sqlitePath = filepath.Join(filepath.Dir(execPath), "HuhnLite.db")
		}
	}
	if _, err := os.Stat(sqlitePath); os.IsNotExist(err) {
		// Fallback auf AppData Verzeichnis
		if configDir, err := os.UserConfigDir(); err == nil {
			sqlitePath = filepath.Join(configDir, "HuhnLite-Wails", "HuhnLite.db")
		}
	}
	// Letzter Versuch: macOS Pfad (aus altem Code)
	if _, err := os.Stat(sqlitePath); os.IsNotExist(err) {
		sqlitePath = "/Users/wernerhofmann/Projekte/HuhnLite-Wails/HuhnLite.db"
	}

	if _, err := os.Stat(sqlitePath); os.IsNotExist(err) {
		log.Printf("Synchronisation ABGEBROCHEN: SQLite-Datei nicht gefunden (CWD, ExecDir oder AppData)")
		return
	}

	log.Printf("Synchronisation: Öffne SQLite unter %s...", sqlitePath)
	sqliteConn, err := sql.Open("sqlite", sqlitePath)
	if err != nil {
		log.Printf("Konnte SQLite für Synchronisation nicht öffnen: %v", err)
		return
	}
	defer sqliteConn.Close()

	// Rename SYSTEM to SYSTEM_KZ in SQLite if it still exists
	for _, t := range []string{"TEXTE", "TEXT_TYPEN"} {
		var hasSystem bool
		rows, err := sqliteConn.Query(fmt.Sprintf("PRAGMA table_info(%s)", t))
		if err == nil {
			for rows.Next() {
				var cid int
				var name, dtype string
				var notnull, pk int
				var dflt_value interface{}
				if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err == nil && name == "SYSTEM" {
					hasSystem = true
				}
			}
			rows.Close()
		}
		if hasSystem {
			log.Printf("Synchronisation: Benenne Spalte SYSTEM in %s (SQLite) nach SYSTEM_KZ um...", t)
			sqliteConn.Exec(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN `SYSTEM` TO SYSTEM_KZ", t))
		}
	}

	// Reihenfolge beachten wegen Fremdschlüsseln (TEXT_TYPEN vor TEXTE vor UEBERSETZUNGEN)
	tables := []string{
		"TABELLENKOPF", "GEWICHTTABELLE", "EIERPREISE", "RASSE", "STALL", "PERSON", "SILO", "EILAGER", "FUTTERSORTEN", "SYSTEMSETTINGS",
		"MWST", "KLASSIFIZIERUNG", "KOSTENTABKOPF", "SHOWTV", "FELD_KATALOG", "TRANSLATEFELDNAMEN", "DYNAMISCHE_SQL",
		"BENUTZERPROFILE", "BENUTZER", "TEXT_TYPEN", "TEXTE", "UEBERSETZUNGEN", "AKTIONEN",
	}

	for _, table := range tables {
		// Prüfen, ob die Tabelle in MariaDB Daten hat
		var count int
		_ = database.SQL.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if count > 0 {
			continue // Tabelle hat schon Daten, überspringen
		}

		log.Printf("Synchronisation: Tabelle %s ist leer. Kopiere aus SQLite...", table)
		rows, err := sqliteConn.Query(fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			log.Printf("Fehler beim Lesen von %s aus SQLite: %v", table, err)
			continue
		}

		cols, _ := rows.Columns()
		placeholders := make([]string, len(cols))
		for i := range placeholders {
			placeholders[i] = "?"
		}

		insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ","), strings.Join(placeholders, ","))
		if table == "USER_STATE" {
			insertQuery = fmt.Sprintf("INSERT INTO %s (USERNAME, `KEY`, VALUE) VALUES (?, ?, ?)", table)
		}

		count = 0
		for rows.Next() {
			vals := make([]interface{}, len(cols))
			valPtrs := make([]interface{}, len(cols))
			for i := range vals {
				valPtrs[i] = &vals[i]
			}
			if err := rows.Scan(valPtrs...); err == nil {
				// MariaDB Fixes für Typen
				for i, v := range vals {
					if b, ok := v.([]byte); ok {
						vals[i] = string(b)
					}
				}
				_, err = database.SQL.Exec(insertQuery, vals...)
				if err != nil {
					log.Printf("FEHLER beim Kopieren in %s: %v (Query: %s)", table, err, insertQuery)
				} else {
					count++
				}
			}
		}
		rows.Close()
		log.Printf("Synchronisation: %d Zeilen in %s kopiert.", count, table)

		// AUTO_INCREMENT immer sicherstellen, damit neue Inserts nicht mit importierten IDs kollidieren
		var maxId int
		err = database.SQL.QueryRow(fmt.Sprintf("SELECT MAX(ID) FROM %s", table)).Scan(&maxId)
		if err == nil && maxId > 0 {
			_, _ = database.SQL.Exec(fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = %d", table, maxId+1))
		}
	}
}
