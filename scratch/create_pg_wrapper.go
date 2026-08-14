package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	content, err := os.ReadFile("backend/db/mysql_wrapper.go")
	if err != nil {
		fmt.Printf("Error reading mysql_wrapper.go: %v\n", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	inHelperBlock := false
	for _, line := range lines {
		if strings.HasPrefix(line, "// Hilfsfunktionen für Typ-Konvertierung") {
			inHelperBlock = true
			continue
		}
		if inHelperBlock && strings.HasPrefix(line, "// --- Buchung (Leistung) Methoden ---") {
			inHelperBlock = false
		}

		if inHelperBlock {
			continue
		}

		newLines = append(newLines, line)
	}

	s := strings.Join(newLines, "\n")

	// Replace package imports and struct types
	s = strings.ReplaceAll(s, "huhnlite-wails/backend/db/repo_mysql", "huhnlite-wails/backend/db/repo_postgres")
	s = strings.ReplaceAll(s, "repo_mysql", "repo_postgres")
	s = strings.ReplaceAll(s, "MySQLWrapper", "PostgresWrapper")
	s = strings.ReplaceAll(s, "NewMySQLWrapper", "NewPostgresWrapper")
	s = strings.ReplaceAll(s, "mysql:   mysql,", "pg:      pg,")
	s = strings.ReplaceAll(s, "mysql:   mysql.WithTx(tx),", "pg:      pg.WithTx(tx),")
	s = strings.ReplaceAll(s, "mysql:", "pg:")
	s = strings.ReplaceAll(s, "mysql *repo_postgres.Queries", "pg *repo_postgres.Queries")
	s = strings.ReplaceAll(s, "w.mysql", "w.pg")
	s = strings.ReplaceAll(s, "[MARIADB-DEBUG]", "[POSTGRES-DEBUG]")

	// Converter function renames to avoid package db collisions
	s = strings.ReplaceAll(s, "convertEilagerbuchung", "convertPgEilagerbuchung")
	s = strings.ReplaceAll(s, "convertEilager", "convertPgEilager")
	s = strings.ReplaceAll(s, "convertEierpreise", "convertPgEierpreise")
	s = strings.ReplaceAll(s, "convertFirmenparameter", "convertPgFirmenparameter")
	s = strings.ReplaceAll(s, "convertPerson", "convertPgPerson")
	s = strings.ReplaceAll(s, "convertVerkauf", "convertPgVerkauf")
	s = strings.ReplaceAll(s, "convertListAktionenRow", "convertPgListAktionenRow")
	s = strings.ReplaceAll(s, "convertAktion", "convertPgAktion")

	// Replace LastInsertId for w.pg.Create... calls
	s = strings.ReplaceAll(s, "id, _ := res.LastInsertId()", "id := int64(res.ID)")

	// Safe type conversion mappings
	s = strings.ReplaceAll(s, "Jumbos:    v.Jumbos,", "Jumbos:    toInt64(v.Jumbos),")
	s = strings.ReplaceAll(s, "Xl:        v.Xl,", "Xl:        toInt64(v.Xl),")
	s = strings.ReplaceAll(s, "Large:     v.Large,", "Large:     toInt64(v.Large),")
	s = strings.ReplaceAll(s, "Medium:    v.Medium,", "Medium:    toInt64(v.Medium),")
	s = strings.ReplaceAll(s, "Small:     v.Small,", "Small:     toInt64(v.Small),")
	s = strings.ReplaceAll(s, "Volleikg:  v.Volleikg,", "Volleikg:  toFloat(v.Volleikg),")
	s = strings.ReplaceAll(s, "Schmutz:   v.Schmutz,", "Schmutz:   toInt64(v.Schmutz),")
	s = strings.ReplaceAll(s, "Knickeier: v.Knickeier,", "Knickeier: toInt64(v.Knickeier),")
	s = strings.ReplaceAll(s, "Brucheier: v.Brucheier,", "Brucheier: toInt64(v.Brucheier),")

	s = strings.ReplaceAll(s, "Jumbos:   v.Jumbos,", "Jumbos:   toInt64(v.Jumbos),")
	s = strings.ReplaceAll(s, "Xl:       v.Xl,", "Xl:       toInt64(v.Xl),")
	s = strings.ReplaceAll(s, "Large:    v.Large,", "Large:    toInt64(v.Large),")
	s = strings.ReplaceAll(s, "Medium:   v.Medium,", "Medium:   toInt64(v.Medium),")
	s = strings.ReplaceAll(s, "Small:    v.Small,", "Small:    toInt64(v.Small),")
	s = strings.ReplaceAll(s, "Volleikg: v.Volleikg,", "Volleikg: toFloat(v.Volleikg),")

	// Specific struct field differences for Postgres
	s = strings.ReplaceAll(s, "Foto:           sql.NullString{String: string(arg.Foto), Valid: len(arg.Foto) > 0},", "Foto:           arg.Foto,")
	s = strings.ReplaceAll(s, "Foto:           []byte(v.Foto.String),", "Foto:           v.Foto,")
	s = strings.ReplaceAll(s, "ROUND:          arg.ROUND,", "Round:          arg.ROUND,")
	s = strings.ReplaceAll(s, "Kosten:         toNullString(arg.Kosten),", "Kosten:         toNullFloat64(arg.Kosten),")
	s = strings.ReplaceAll(s, "Bio:              arg.Bio,", "Bio:              toInt16(arg.Bio),")
	s = strings.ReplaceAll(s, "Verbucht:         arg.Verbucht,", "Verbucht:         toInt16(arg.Verbucht),")
	s = strings.ReplaceAll(s, "Bio:              v.Bio,", "Bio:              v.Bio != 0,")
	s = strings.ReplaceAll(s, "Verbucht:         v.Verbucht,", "Verbucht:         v.Verbucht != 0,")

	// UpdateAktion return 2 values in Postgres
	s = strings.ReplaceAll(s, "err := w.pg.UpdateAktion(", "_, err := w.pg.UpdateAktion(")

	// convertPgListAktionenRow toNullInt64 mapping
	s = strings.ReplaceAll(s, "IDUser:           v.IDUser,", "IDUser:           toNullInt64(v.IDUser),")
	s = strings.ReplaceAll(s, "IntervallTage:    v.IntervallTage,", "IntervallTage:    toNullInt64(v.IntervallTage),")
	s = strings.ReplaceAll(s, "AnzahlIntervalle: v.AnzahlIntervalle,", "AnzahlIntervalle: toNullInt64(v.AnzahlIntervalle),")
	s = strings.ReplaceAll(s, "Erledigt:         v.Erledigt,", "Erledigt:         toNullInt64(v.Erledigt),")
	s = strings.ReplaceAll(s, "IDUserErledigt:   v.IDUserErledigt,", "IDUserErledigt:   toNullInt64(v.IDUserErledigt),")

	// Replace CreateTabellenkopf, CreateTextTyp, CreateText
	s = strings.ReplaceAll(s, "func (w *PostgresWrapper) CreateTabellenkopf(ctx context.Context, arg repo.CreateTabellenkopfParams) (repo.Tabellenkopf, error) {\n\tlog.Printf(\"[DB] CreateTabellenkopf called for Typ: %s, Nr: %d\", arg.Tabellentyp, arg.Tabellennummer)\n\tquery := `INSERT INTO TABELLENKOPF (TABELLENTYP, TABELLENNUMMER, BEZEICHNUNG, ANLAGEDATUM, DATUM) VALUES (?, ?, ?, ?, ?)`\n\tres, err := w.db.ExecContext(ctx, query, arg.Tabellentyp, arg.Tabellennummer, arg.Bezeichnung, arg.Anlagedatum, arg.Datum)\n\tif err != nil {\n\t\tlog.Printf(\"[DB] CreateTabellenkopf Error: %v\", err)\n\t\treturn repo.Tabellenkopf{}, err\n\t}\n\tid := int64(res.ID)\n\treturn repo.Tabellenkopf{\n\t\tID:             id,\n\t\tTabellentyp:    toString(arg.Tabellentyp),\n\t\tTabellennummer: toInt64(arg.Tabellennummer),\n\t\tBezeichnung:    arg.Bezeichnung,\n\t\tAnlagedatum:    arg.Anlagedatum,\n\t\tDatum:          arg.Datum,\n\t}, nil\n}", "func (w *PostgresWrapper) CreateTabellenkopf(ctx context.Context, arg repo.CreateTabellenkopfParams) (repo.Tabellenkopf, error) {\n\tres, err := w.pg.CreateTabellenkopf(ctx, repo_postgres.CreateTabellenkopfParams{\n\t\tTabellentyp:    toString(arg.Tabellentyp),\n\t\tTabellennummer: int32(arg.Tabellennummer),\n\t\tBezeichnung:    arg.Bezeichnung,\n\t\tAnlagedatum:    arg.Anlagedatum,\n\t\tDatum:          arg.Datum,\n\t})\n\tif err != nil {\n\t\treturn repo.Tabellenkopf{}, err\n\t}\n\treturn repo.Tabellenkopf{\n\t\tID:             int64(res.ID),\n\t\tTabellentyp:    res.Tabellentyp,\n\t\tTabellennummer: int64(res.Tabellennummer),\n\t\tBezeichnung:    res.Bezeichnung,\n\t\tAnlagedatum:    res.Anlagedatum,\n\t\tDatum:          res.Datum,\n\t}, nil\n}")

	s = strings.ReplaceAll(s, "func (w *PostgresWrapper) CreateTextTyp(ctx context.Context, arg repo.CreateTextTypParams) (repo.TextTypen, error) {\n\tlog.Printf(\"[DB] CreateTextTyp called: %s\", arg.Kz)\n\tquery := \"INSERT INTO TEXT_TYPEN (KZ, BEZEICHNUNG, `SYSTEM_KZ`, `STATUS`) VALUES (?, ?, ?, ?)\"\n\tres, err := w.db.ExecContext(ctx, query, arg.Kz, arg.Bezeichnung, arg.SystemKz, arg.Status)\n\tif err != nil {\n\t\tlog.Printf(\"[DB] CreateTextTyp Error: %v\", err)\n\t\treturn repo.TextTypen{}, err\n\t}\n\tid := int64(res.ID)\n\treturn repo.TextTypen{\n\t\tID:          id,\n\t\tKz:          arg.Kz,\n\t\tBezeichnung: arg.Bezeichnung,\n\t\tSystemKz:    arg.SystemKz,\n\t\tStatus:      arg.Status,\n\t}, nil\n}", "func (w *PostgresWrapper) CreateTextTyp(ctx context.Context, arg repo.CreateTextTypParams) (repo.TextTypen, error) {\n\tres, err := w.pg.CreateTextTyp(ctx, repo_postgres.CreateTextTypParams{\n\t\tKz:          arg.Kz,\n\t\tBezeichnung: arg.Bezeichnung,\n\t\tSystemKz:    int32(arg.SystemKz),\n\t\tStatus:      int32(arg.Status),\n\t})\n\tif err != nil {\n\t\treturn repo.TextTypen{}, err\n\t}\n\treturn repo.TextTypen{\n\t\tID:          int64(res.ID),\n\t\tKz:          res.Kz,\n\t\tBezeichnung: res.Bezeichnung,\n\t\tSystemKz:    int64(res.SystemKz),\n\t\tStatus:      int64(res.Status),\n\t}, nil\n}")

	s = strings.ReplaceAll(s, "func (w *PostgresWrapper) CreateText(ctx context.Context, arg repo.CreateTextParams) (repo.Texte, error) {\n\tlog.Printf(\"[DB] CreateText called for Typ: %s\", arg.TextTypKz)\n\tquery := \"INSERT INTO TEXTE (TEXT_TYP_KZ, KZ, `SYSTEM_KZ`, `STATUS`) VALUES (?, ?, ?, ?)\"\n\tres, err := w.db.ExecContext(ctx, query, arg.TextTypKz, arg.Kz, arg.SystemKz, arg.Status)\n\tif err != nil {\n\t\tlog.Printf(\"[DB] CreateText Error: %v\", err)\n\t\treturn repo.Texte{}, err\n\t}\n\tid := int64(res.ID)\n\treturn repo.Texte{\n\t\tID:        id,\n\t\tTextTypKz: arg.TextTypKz,\n\t\tKz:        arg.Kz,\n\t\tSystemKz:  arg.SystemKz,\n\t\tStatus:    arg.Status,\n\t}, nil\n}", "func (w *PostgresWrapper) CreateText(ctx context.Context, arg repo.CreateTextParams) (repo.Texte, error) {\n\tres, err := w.pg.CreateText(ctx, repo_postgres.CreateTextParams{\n\t\tTextTypKz: arg.TextTypKz,\n\t\tKz:        arg.Kz,\n\t\tSystemKz:  int32(arg.SystemKz),\n\t\tStatus:    int32(arg.Status),\n\t})\n\tif err != nil {\n\t\treturn repo.Texte{}, err\n\t}\n\treturn repo.Texte{\n\t\tID:        int64(res.ID),\n\t\tTextTypKz: res.TextTypKz,\n\t\tKz:        res.Kz,\n\t\tSystemKz:  int64(res.SystemKz),\n\t\tStatus:    int64(res.Status),\n\t}, nil\n}")

	err = os.WriteFile("backend/db/postgres_wrapper.go", []byte(s), 0644)
	if err != nil {
		fmt.Printf("Error writing postgres_wrapper.go: %v\n", err)
		return
	}

	fmt.Println("Successfully re-generated backend/db/postgres_wrapper.go!")
}
