package db

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"huhnlite-wails/backend/db/repo"
	"huhnlite-wails/backend/db/repo_mysql"
	"log"
	"strconv"
	"strings"
)

// MySQLWrapper implementiert das Querier-Interface, nutzt aber intern das MySQL-Repo
type MySQLWrapper struct {
	*repo.Queries
	mysql *repo_mysql.Queries
	db    *sql.DB
}

func NewMySQLWrapper(sqlite *repo.Queries, mysql *repo_mysql.Queries, db *sql.DB) *MySQLWrapper {
	return &MySQLWrapper{
		Queries: sqlite,
		mysql:   mysql,
		db:      db,
	}
}

func (w *MySQLWrapper) WithTx(tx *sql.Tx) *MySQLWrapper {
	return &MySQLWrapper{
		Queries: w.Queries.WithTx(tx),
		mysql:   w.mysql.WithTx(tx),
		db:      w.db,
	}
}

// Hilfsfunktionen für Typ-Konvertierung
func DecodeBase64IfNeeded(s string) string {
	// Aggressives Dekodieren: Solange es wie Base64 aussieht und wir es dekodieren können, tun wir es.
	// Das löst das Problem der doppelten Kodierung (ZUE9PQ== -> eA== -> x).
	current := s
	for len(current) >= 4 && (strings.HasSuffix(current, "=") || len(current)%4 == 0) {
		data, err := base64.StdEncoding.DecodeString(current)
		if err != nil || len(data) == 0 {
			break
		}
		current = string(data)
		// Wenn wir bei einem einzelnen Zeichen angekommen sind, stop.
		if len(current) <= 1 {
			break
		}
	}
	return current
}

func SanitizeKZ(s string) string {
	val := DecodeBase64IfNeeded(s)
	if len(val) > 1 {
		return val[:1]
	}
	if val == "" {
		return "x"
	}
	return val
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	}
	return fmt.Sprintf("%v", v)
}

func toFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int64:
		return float64(t)
	case int32:
		return float64(t)
	case int:
		return float64(t)
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	case []byte:
		f, _ := strconv.ParseFloat(string(t), 64)
		return f
	}
	return 0
}
func toInt32(v interface{}) int32 {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case int32:
		return t
	case int64:
		return int32(t)
	case int:
		return int32(t)
	case float64:
		return int32(t)
	case string:
		var i int32
		fmt.Sscanf(t, "%d", &i)
		return i
	case []byte:
		var i int32
		fmt.Sscanf(string(t), "%d", &i)
		return i
	}
	return 0
}

func toInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case int64:
		return t
	case int32:
		return int64(t)
	case int:
		return int64(t)
	case float64:
		return int64(t)
	case string:
		var i int64
		fmt.Sscanf(t, "%d", &i)
		return i
	case []byte:
		var i int64
		fmt.Sscanf(string(t), "%d", &i)
		return i
	}
	return 0
}
func toNullFloat64(v interface{}) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{Valid: false}
	}
	switch t := v.(type) {
	case float64:
		return sql.NullFloat64{Float64: t, Valid: true}
	case float32:
		return sql.NullFloat64{Float64: float64(t), Valid: true}
	case int64:
		return sql.NullFloat64{Float64: float64(t), Valid: true}
	case int32:
		return sql.NullFloat64{Float64: float64(t), Valid: true}
	case int:
		return sql.NullFloat64{Float64: float64(t), Valid: true}
	case []byte:
		f, _ := strconv.ParseFloat(string(t), 64)
		return sql.NullFloat64{Float64: f, Valid: true}
	}
	return sql.NullFloat64{Valid: false}
}

func toNullInt64(v interface{}) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{Valid: false}
	}
	switch t := v.(type) {
	case int64:
		return sql.NullInt64{Int64: t, Valid: true}
	case int32:
		return sql.NullInt64{Int64: int64(t), Valid: true}
	case int:
		return sql.NullInt64{Int64: int64(t), Valid: true}
	case float64:
		return sql.NullInt64{Int64: int64(t), Valid: true}
	case []byte:
		var i int64
		fmt.Sscanf(string(t), "%lld", &i)
		return sql.NullInt64{Int64: i, Valid: true}
	}
	return sql.NullInt64{Valid: false}
}
func toNullInt32(v interface{}) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{Valid: false}
	}
	switch t := v.(type) {
	case int32:
		return sql.NullInt32{Int32: t, Valid: true}
	case int64:
		return sql.NullInt32{Int32: int32(t), Valid: true}
	case int:
		return sql.NullInt32{Int32: int32(t), Valid: true}
	case sql.NullInt64:
		return sql.NullInt32{Int32: int32(t.Int64), Valid: t.Valid}
	case []byte:
		var i int32
		fmt.Sscanf(string(t), "%d", &i)
		return sql.NullInt32{Int32: i, Valid: true}
	}
	return sql.NullInt32{Valid: false}
}

func toNullString(v interface{}) sql.NullString {
	if v == nil {
		return sql.NullString{Valid: false}
	}
	switch t := v.(type) {
	case string:
		return sql.NullString{String: t, Valid: true}
	case sql.NullString:
		return t
	case []byte:
		return sql.NullString{String: string(t), Valid: true}
	case float64:
		return sql.NullString{String: fmt.Sprintf("%.2f", t), Valid: true}
	case sql.NullFloat64:
		if !t.Valid {
			return sql.NullString{Valid: false}
		}
		return sql.NullString{String: fmt.Sprintf("%.2f", t.Float64), Valid: true}
	}
	return sql.NullString{String: fmt.Sprint(v), Valid: true}
}

// --- Buchung (Leistung) Methoden ---
func (w *MySQLWrapper) ListBuchungen(ctx context.Context) ([]repo.ListBuchungenRow, error) {
	query := `
		SELECT B.ID, B.ID_HERDEN, B.LW, B.HERDENNUMMER, B.BUCHUNGSDATUM, B.GEWICHTPROBE, B.KONTROLLGEWICHT, 
		       B.KLASSEA, B.VERLUSTE, B.EIMASSE, B.SCHMUTZ, B.KNICKEIER, B.VOLLEI, B.BRUCHEIER, 
		       B.TIERBESTAND, B.ID_EITABELLE, B.ID_DGEWICHTTAB, B.FUTTERKTAG, B.SILONR, B.KL6, 
		       B.VERMITTELTAM, B.SMALL, B.LARGE, B.MEDIUM, B.XL, B.ZEITSTEMPEL, B.DGEWICHTEI, B.AW, B.VERMITTELT,
		       H.HERDENNUMMER AS HERDEN_NUMMER_REL, H.BEZEICHNUNG AS HERDEN_BEZEICHNUNG_REL, 
		       H.ID_EILAGER AS HERDEN_ID_EILAGER, H.AKTIV AS HERDEN_AKTIV_REL
		FROM BUCHUNG B
		LEFT JOIN HERDEN H ON B.ID_HERDEN = H.ID
		WHERE B.VERMITTELT != 'S'
		ORDER BY B.BUCHUNGSDATUM DESC
	`
	rows, err := w.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []repo.ListBuchungenRow
	for rows.Next() {
		var i repo.ListBuchungenRow
		var id, idh, lw, hnr, ka, v, s, k, vo, b, tb, idet, idgt, sn, sm, l, m, xl, aw, hnr_r, h_ide, h_akt int32
		var gp, kg, em, fk, dge interface{}
		var bdat, kl6, vam, zst, h_bez_r sql.NullString
		if err := rows.Scan(
			&id, &idh, &lw, &hnr, &bdat, &gp, &kg,
			&ka, &v, &em, &s, &k, &vo, &b,
			&tb, &idet, &idgt, &fk, &sn, &kl6,
			&vam, &sm, &l, &m, &xl, &zst, &dge, &aw, &i.Vermittelt,
			&hnr_r, &h_bez_r, &h_ide, &h_akt,
		); err != nil {
			return nil, err
		}
		i.ID = int64(id)
		i.IDHerden = int64(idh)
		i.Lw = int64(lw)
		i.Herdennummer = int64(hnr)
		i.Buchungsdatum = bdat.String
		i.Gewichtprobe = int64(toFloat(gp))
		i.Kontrollgewicht = toFloat(kg)
		i.Klassea = int64(ka)
		i.Verluste = int64(v)
		i.Eimasse = toFloat(em)
		i.Schmutz = int64(s)
		i.Knickeier = int64(k)
		i.Vollei = float64(vo)
		i.Brucheier = int64(b)
		i.Tierbestand = int64(tb)
		i.IDEitabelle = int64(idet)
		i.IDDgewichttab = int64(idgt)
		i.Futterktag = int64(toFloat(fk))
		i.Silonr = int64(sn)
		i.Kl6 = int64(toInt64(kl6.String))
		i.Vermitteltam = vam.String
		i.Small = int64(sm)
		i.Large = int64(l)
		i.Medium = int64(m)
		i.Xl = int64(xl)
		i.Zeitstempel = zst.String
		i.Dgewichtei = toFloat(dge)
		i.Aw = int64(aw)
		i.HerdenNummerRel = sql.NullInt64{Int64: int64(hnr_r), Valid: true}
		i.HerdenBezeichnungRel = h_bez_r
		i.HerdenIDEilager = sql.NullInt64{Int64: int64(h_ide), Valid: true}
		i.HerdenAktivRel = sql.NullInt64{Int64: int64(h_akt), Valid: true}
		items = append(items, i)
	}
	return items, nil
}

func (w *MySQLWrapper) GetBuchung(ctx context.Context, id int64) (repo.Buchung, error) {
	v, err := w.mysql.GetBuchung(ctx, int32(id))
	if err != nil {
		return repo.Buchung{}, err
	}
	return repo.Buchung{
		ID:              int64(v.ID),
		IDHerden:        int64(v.IDHerden),
		Lw:              int64(v.Lw),
		Herdennummer:    int64(v.Herdennummer),
		Buchungsdatum:   v.Buchungsdatum,
		Gewichtprobe:    int64(v.Gewichtprobe),
		Kontrollgewicht: v.Kontrollgewicht,
		Klassea:         int64(v.Klassea),
		Verluste:        int64(v.Verluste),
		Eimasse:         v.Eimasse,
		Schmutz:         int64(v.Schmutz),
		Knickeier:       int64(v.Knickeier),
		Vollei:          v.Vollei,
		Brucheier:       int64(v.Brucheier),
		Tierbestand:     int64(v.Tierbestand),
		IDEitabelle:     int64(v.IDEitabelle),
		IDDgewichttab:   int64(v.IDDgewichttab),
		Futterktag:      int64(v.Futterktag),
		Silonr:          int64(v.Silonr),
		Kl6:             int64(v.Kl6),
		Vermitteltam:    v.Vermitteltam,
		Small:           int64(v.Small),
		Large:           int64(v.Large),
		Medium:          int64(v.Medium),
		Xl:              int64(v.Xl),
		Zeitstempel:     v.Zeitstempel,
		Dgewichtei:      v.Dgewichtei,
		Aw:              int64(v.Aw),
		Vermittelt:      v.Vermittelt,
	}, nil
}

func (w *MySQLWrapper) CreateBuchung(ctx context.Context, arg repo.CreateBuchungParams) (repo.Buchung, error) {
	res, err := w.mysql.CreateBuchung(ctx, repo_mysql.CreateBuchungParams{
		IDHerden:        int32(arg.IDHerden),
		Lw:              int32(arg.Lw),
		Herdennummer:    int32(arg.Herdennummer),
		Buchungsdatum:   arg.Buchungsdatum,
		Gewichtprobe:    int32(arg.Gewichtprobe),
		Kontrollgewicht: arg.Kontrollgewicht,
		Klassea:         int32(arg.Klassea),
		Verluste:        int32(arg.Verluste),
		Eimasse:         arg.Eimasse,
		Schmutz:         int32(arg.Schmutz),
		Knickeier:       int32(arg.Knickeier),
		Vollei:          arg.Vollei,
		Brucheier:       int32(arg.Brucheier),
		Tierbestand:     int32(arg.Tierbestand),
		IDEitabelle:     int32(arg.IDEitabelle),
		IDDgewichttab:   int32(arg.IDDgewichttab),
		Futterktag:      int32(arg.Futterktag),
		Silonr:          int32(arg.Silonr),
		Kl6:             int32(arg.Kl6),
		Vermitteltam:    arg.Vermitteltam,
		Small:           int32(arg.Small),
		Large:           int32(arg.Large),
		Medium:          int32(arg.Medium),
		Xl:              int32(arg.Xl),
		Dgewichtei:      arg.Dgewichtei,
		Zeitstempel:     arg.Zeitstempel,
		Aw:              int32(arg.Aw),
		Vermittelt:      toString(arg.Vermittelt),
	})
	if err != nil {
		return repo.Buchung{}, err
	}
	id, _ := res.LastInsertId()
	return w.GetBuchung(ctx, id)
}

func (w *MySQLWrapper) UpdateBuchung(ctx context.Context, arg repo.UpdateBuchungParams) (repo.Buchung, error) {
	_, err := w.mysql.UpdateBuchung(ctx, repo_mysql.UpdateBuchungParams{
		Lw:              int32(arg.Lw),
		Herdennummer:    int32(arg.Herdennummer),
		Buchungsdatum:   arg.Buchungsdatum,
		Gewichtprobe:    int32(arg.Gewichtprobe),
		Kontrollgewicht: arg.Kontrollgewicht,
		Klassea:         int32(arg.Klassea),
		Verluste:        int32(arg.Verluste),
		Eimasse:         arg.Eimasse,
		Schmutz:         int32(arg.Schmutz),
		Knickeier:       int32(arg.Knickeier),
		Vollei:          arg.Vollei,
		Brucheier:       int32(arg.Brucheier),
		Tierbestand:     int32(arg.Tierbestand),
		IDEitabelle:     int32(arg.IDEitabelle),
		IDDgewichttab:   int32(arg.IDDgewichttab),
		Futterktag:      int32(arg.Futterktag),
		Silonr:          int32(arg.Silonr),
		Kl6:             int32(arg.Kl6),
		Vermitteltam:    arg.Vermitteltam,
		Small:           int32(arg.Small),
		Large:           int32(arg.Large),
		Medium:          int32(arg.Medium),
		Xl:              int32(arg.Xl),
		Dgewichtei:      arg.Dgewichtei,
		Zeitstempel:     arg.Zeitstempel,
		Aw:              int32(arg.Aw),
		Vermittelt:      toString(arg.Vermittelt),
		ID:              int32(arg.ID),
	})
	if err != nil {
		return repo.Buchung{}, err
	}
	return w.GetBuchung(ctx, arg.ID)
}

// --- Herde Methoden ---

func (w *MySQLWrapper) ListHerden(ctx context.Context) ([]repo.ListHerdenRow, error) {
	res, err := w.mysql.ListHerden(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.ListHerdenRow, len(res))
	for i, v := range res {
		items[i] = repo.ListHerdenRow{
			ID:                    int64(v.ID),
			IDSilo:                int64(v.IDSilo),
			IDStall:               int64(v.IDStall),
			IDEilager:             int64(v.IDEilager),
			IDGewichttab:          int64(v.IDGewichttab),
			IDZuechter:            int64(v.IDZuechter),
			IDRasse:               int64(v.IDRasse),
			Herdennummer:          int64(v.Herdennummer),
			Bezeichnung:           v.Bezeichnung,
			Anfangskosten:         int64(v.Anfangskosten),
			Anfangsbestand:        int64(v.Anfangsbestand),
			Einstalldatum:         v.Einstalldatum,
			Legedatum:             v.Legedatum,
			Einstallkosten:        v.Einstallkosten,
			Datum:                 v.Datum,
			Zeitstempel:           v.Zeitstempel,
			Aktiv:                 int64(v.Aktiv),
			Aw:                    int64(v.Aw),
			Allebuchungenmitdatum: int64(v.Allebuchungenmitdatum),
			StallBezeichnung:      v.StallBezeichnung,
		}
	}
	return items, nil
}

func (w *MySQLWrapper) GetHerde(ctx context.Context, id int64) (repo.Herden, error) {
	v, err := w.mysql.GetHerde(ctx, int32(id))
	if err != nil {
		return repo.Herden{}, err
	}
	return repo.Herden{
		ID:                    int64(v.ID),
		IDSilo:                int64(v.IDSilo),
		IDStall:               int64(v.IDStall),
		IDEilager:             int64(v.IDEilager),
		IDRasse:               int64(v.IDRasse),
		IDZuechter:            int64(v.IDZuechter),
		Herdennummer:          int64(v.Herdennummer),
		Bezeichnung:           v.Bezeichnung,
		Anfangsbestand:        int64(v.Anfangsbestand),
		Einstalldatum:         v.Einstalldatum,
		Legedatum:             v.Legedatum,
		Einstallkosten:        v.Einstallkosten,
		Aktiv:                 int64(v.Aktiv),
		Allebuchungenmitdatum: int64(v.Allebuchungenmitdatum),
	}, nil
}

func (w *MySQLWrapper) CreateHerde(ctx context.Context, arg repo.CreateHerdeParams) (repo.Herden, error) {
	res, err := w.mysql.CreateHerde(ctx, repo_mysql.CreateHerdeParams{
		IDSilo:                int32(arg.IDSilo),
		IDStall:               int32(arg.IDStall),
		IDEilager:             int32(arg.IDEilager),
		IDRasse:               int32(arg.IDRasse),
		IDZuechter:            int32(arg.IDZuechter),
		Herdennummer:          int32(arg.Herdennummer),
		Bezeichnung:           arg.Bezeichnung,
		Anfangsbestand:        int32(arg.Anfangsbestand),
		Einstalldatum:         arg.Einstalldatum,
		Legedatum:             arg.Legedatum,
		Einstallkosten:        arg.Einstallkosten,
		Aktiv:                 int32(arg.Aktiv),
		Allebuchungenmitdatum: int32(arg.Allebuchungenmitdatum),
	})
	if err != nil {
		return repo.Herden{}, err
	}
	id, _ := res.LastInsertId()
	return w.GetHerde(ctx, id)
}

func (w *MySQLWrapper) UpdateHerde(ctx context.Context, arg repo.UpdateHerdeParams) (repo.Herden, error) {
	_, err := w.mysql.UpdateHerde(ctx, repo_mysql.UpdateHerdeParams{
		IDSilo:                int32(arg.IDSilo),
		IDStall:               int32(arg.IDStall),
		IDEilager:             int32(arg.IDEilager),
		IDRasse:               int32(arg.IDRasse),
		IDZuechter:            int32(arg.IDZuechter),
		Herdennummer:          int32(arg.Herdennummer),
		Bezeichnung:           arg.Bezeichnung,
		Anfangsbestand:        int32(arg.Anfangsbestand),
		Einstalldatum:         arg.Einstalldatum,
		Legedatum:             arg.Legedatum,
		Einstallkosten:        arg.Einstallkosten,
		Aktiv:                 int32(arg.Aktiv),
		Allebuchungenmitdatum: int32(arg.Allebuchungenmitdatum),
		ID:                    int32(arg.ID),
	})
	if err != nil {
		return repo.Herden{}, err
	}
	return w.GetHerde(ctx, arg.ID)
}

// --- Statistik / Grafik Methoden ---

func (w *MySQLWrapper) GetEggBookingYears(ctx context.Context, arg repo.GetEggBookingYearsParams) ([]interface{}, error) {
	res, err := w.mysql.GetEggBookingYears(ctx, repo_mysql.GetEggBookingYearsParams{
		IDHerden:   toInt32(arg.IDHerden),
		OnlyActive: arg.OnlyActive,
	})
	if err != nil {
		return nil, err
	}
	items := make([]interface{}, len(res))
	for i, v := range res {
		items[i] = v
	}
	return items, nil
}

func (w *MySQLWrapper) GetEggStatsByHerde(ctx context.Context, arg repo.GetEggStatsByHerdeParams) (repo.GetEggStatsByHerdeRow, error) {
	res, err := w.mysql.GetEggStatsByHerde(ctx, repo_mysql.GetEggStatsByHerdeParams{
		ID:       toInt32(arg.ID),
		IDHerden: toInt32(arg.IDHerden),
	})
	if err != nil {
		return repo.GetEggStatsByHerdeRow{}, err
	}
	return repo.GetEggStatsByHerdeRow{
		SumKlasseA:  res.SumKlasseA,
		SumSmall:    res.SumSmall,
		SumMedium:   res.SumMedium,
		SumLarge:    res.SumLarge,
		SumXl:       res.SumXl,
		SumVerluste: int64(res.SumVerluste),
	}, nil
}

func (w *MySQLWrapper) GetEggStatsByHerdeFiltered(ctx context.Context, arg repo.GetEggStatsByHerdeFilteredParams) (repo.GetEggStatsByHerdeFilteredRow, error) {
	res, err := w.mysql.GetEggStatsByHerdeFiltered(ctx, repo_mysql.GetEggStatsByHerdeFilteredParams{
		IDHerden:   toInt32(arg.IDHerden),
		OnlyActive: arg.OnlyActive,
		Year:       toString(arg.Year),
		Quarter:    toString(arg.Quarter),
		Month:      toString(arg.Month),
	})
	if err != nil {
		return repo.GetEggStatsByHerdeFilteredRow{}, err
	}
	return repo.GetEggStatsByHerdeFilteredRow{
		SumKlasseA:  res.SumKlasseA,
		SumSmall:    res.SumSmall,
		SumMedium:   res.SumMedium,
		SumLarge:    res.SumLarge,
		SumXl:       res.SumXl,
		SumVerluste: res.SumVerluste,
	}, nil
}

func (w *MySQLWrapper) GetEggStatsWeeklyByHerde(ctx context.Context, idHerden int64) ([]repo.GetEggStatsWeeklyByHerdeRow, error) {
	res, err := w.mysql.GetEggStatsWeeklyByHerde(ctx, int32(idHerden))
	if err != nil {
		return nil, err
	}
	items := make([]repo.GetEggStatsWeeklyByHerdeRow, len(res))
	for i, v := range res {
		items[i] = repo.GetEggStatsWeeklyByHerdeRow{
			Lebenswoche:  int64(v.Lebenswoche),
			LetztesDatum: v.LetztesDatum,
			SumKlasseA:   v.SumKlasseA,
			SumSmall:     v.SumSmall,
			SumMedium:    v.SumMedium,
			SumLarge:     v.SumLarge,
			SumXl:        v.SumXl,
			SumVerluste:  v.SumVerluste,
		}
	}
	return items, nil
}

// --- Eilager Methoden ---

func (w *MySQLWrapper) ListEilager(ctx context.Context) ([]repo.ListEilagerRow, error) {
	log.Printf("[DB] ListEilager called")
	query := `SELECT id, lagernummer, kz, bezeichnung, letzte_buchung, jumbos, xl, large, medium, small, volleikg, aw, klasse6, klasse7 FROM eilager ORDER BY lagernummer`
	rows, err := w.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("[DB] ListEilager Query Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []repo.ListEilagerRow
	for rows.Next() {
		var i repo.ListEilagerRow
		var id, ln, j, xl, l, m, s, aw, k6, k7 interface{}
		var kz, bez, lb sql.NullString
		var v interface{}
		if err := rows.Scan(
			&id, &ln, &kz, &bez, &lb,
			&j, &xl, &l, &m, &s, &v,
			&aw, &k6, &k7,
		); err != nil {
			log.Printf("[DB] ListEilager Scan Error: %v", err)
			return nil, err
		}
		i.ID = toInt64(id)
		i.Lagernummer = toInt64(ln)
		i.Kz = kz.String
		i.Bezeichnung = bez.String
		i.LetzteBuchung = lb.String
		i.Jumbos = toInt64(j)
		i.Xl = toInt64(xl)
		i.Large = toInt64(l)
		i.Medium = toInt64(m)
		i.Small = toInt64(s)
		i.Volleikg = toFloat(v)
		i.Aw = toInt64(aw)
		i.Klasse6 = toString(k6)
		i.Klasse7 = toString(k7)
		items = append(items, i)
	}
	log.Printf("[DB] ListEilager found %d records", len(items))
	return items, nil
}

func (w *MySQLWrapper) GetEilager(ctx context.Context, id int64) (repo.Eilager, error) {
	v, err := w.mysql.GetEilager(ctx, int32(id))
	if err != nil {
		return repo.Eilager{}, err
	}
	return convertEilager(v), nil
}

func (w *MySQLWrapper) CreateEilager(ctx context.Context, arg repo.CreateEilagerParams) (repo.Eilager, error) {
	res, err := w.mysql.CreateEilager(ctx, repo_mysql.CreateEilagerParams{
		Lagernummer:   int32(arg.Lagernummer),
		Kz:            toString(arg.Kz),
		Bezeichnung:   arg.Bezeichnung,
		LetzteBuchung: arg.LetzteBuchung,
	})
	if err != nil {
		return repo.Eilager{}, err
	}
	id, _ := res.LastInsertId()
	return w.GetEilager(ctx, id)
}

func (w *MySQLWrapper) UpdateEilager(ctx context.Context, arg repo.UpdateEilagerParams) (repo.Eilager, error) {
	_, err := w.mysql.UpdateEilager(ctx, repo_mysql.UpdateEilagerParams{
		Lagernummer:   int32(arg.Lagernummer),
		Kz:            toString(arg.Kz),
		Bezeichnung:   arg.Bezeichnung,
		LetzteBuchung: arg.LetzteBuchung,
		ID:            int32(arg.ID),
	})

	if err != nil {
		return repo.Eilager{}, err
	}
	return w.GetEilager(ctx, arg.ID)
}

func (w *MySQLWrapper) DeleteEilager(ctx context.Context, id int64) error {
	return w.mysql.DeleteEilager(ctx, int32(id))
}

func (w *MySQLWrapper) GetBestandsuebersicht(ctx context.Context, arg repo.GetBestandsuebersichtParams) ([]repo.GetBestandsuebersichtRow, error) {
	log.Printf("[DB] GetBestandsuebersicht called with IDEilager: %v", arg.IDEilager)
	query := `
		SELECT CHARGE,
		       LAGERPLATZ_BEZEICHNUNG,
		       LAGERPLATZ_ID,
		       EILAGER_KZ,
		       EILAGER_ID,
		       EILAGER_BEZEICHNUNG,
		       SUM(JUMBOS)   AS JUMBOS,
		       SUM(XL)       AS XL,
		       SUM(LARGE)    AS LARGE,
		       SUM(MEDIUM)   AS MEDIUM,
		       SUM(SMALL)    AS SMALL,
		       SUM(VOLLEIKG) AS VOLLEIKG
		FROM (
		         -- Positive Buchungen (Einlagerung / Ziel einer Umbuchung)
		         SELECT EB.CHARGE,
		                LP.BEZEICHNUNG AS LAGERPLATZ_BEZEICHNUNG,
		                LP.ID          AS LAGERPLATZ_ID,
		                E.KZ           AS EILAGER_KZ,
		                EB.ID_EILAGER  AS EILAGER_ID,
		                E.BEZEICHNUNG  AS EILAGER_BEZEICHNUNG,
		                EB.JUMBOS,
		                EB.XL,
		                EB.LARGE,
		                EB.MEDIUM,
		                EB.SMALL,
		                EB.VOLLEIKG
		         FROM eilagerbuchung EB
		                  LEFT JOIN eilager E ON EB.ID_EILAGER = E.ID
		                  LEFT JOIN lagerplatz LP ON EB.ID_LAGERPLATZ = LP.ID
		         WHERE (? = 0 OR EB.ID_EILAGER = ?)
		
		         UNION ALL
		
		         -- Negative Buchungen (Abgang / Quelle einer Umbuchung)
		         SELECT EB.CHARGE,
		                ''                 AS LAGERPLATZ_BEZEICHNUNG,
		                0                  AS LAGERPLATZ_ID,
		                E.KZ               AS EILAGER_KZ,
		                EB.ID_FREMDESLAGER AS EILAGER_ID,
		                E.BEZEICHNUNG      AS EILAGER_BEZEICHNUNG,
		                -EB.JUMBOS,
		                -EB.XL,
		                -EB.LARGE,
		                -EB.MEDIUM,
		                -EB.SMALL,
		                -EB.VOLLEIKG
		         FROM eilagerbuchung EB
		                  LEFT JOIN eilager E ON EB.ID_FREMDESLAGER = E.ID
		         WHERE EB.ID_FREMDESLAGER != 0 AND (? = 0 OR EB.ID_FREMDESLAGER = ?)
		) t
		GROUP BY CHARGE, LAGERPLATZ_ID, EILAGER_ID, EILAGER_KZ, EILAGER_BEZEICHNUNG
		ORDER BY EILAGER_BEZEICHNUNG, CHARGE, LAGERPLATZ_BEZEICHNUNG
	`

	idInt := toInt64(arg.IDEilager)
	rows, err := w.db.QueryContext(ctx, query, idInt, idInt, idInt, idInt)
	if err != nil {
		log.Printf("[DB] GetBestandsuebersicht Query Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []repo.GetBestandsuebersichtRow
	for rows.Next() {
		var charge, lpBez, eiKz, eiBez sql.NullString
		var lpid, eid interface{}
		var j, xl, l, m, s, v interface{}

		if err := rows.Scan(
			&charge, &lpBez, &lpid, &eiKz, &eid, &eiBez,
			&j, &xl, &l, &m, &s, &v,
		); err != nil {
			log.Printf("[DB] GetBestandsuebersicht Scan Error: %v", err)
			return nil, err
		}

		items = append(items, repo.GetBestandsuebersichtRow{
			Charge:                charge.String,
			LagerplatzBezeichnung: lpBez.String,
			LagerplatzID:          toNullInt64(lpid),
			EilagerKz:             eiKz.String,
			EilagerID:             toInt64(eid),
			EilagerBezeichnung:    eiBez,
			Jumbos:                toNullFloat64(j),
			Xl:                    toNullFloat64(xl),
			Large:                 toNullFloat64(l),
			Medium:                toNullFloat64(m),
			Small:                 toNullFloat64(s),
			Volleikg:              toNullFloat64(v),
		})
	}
	log.Printf("[DB] GetBestandsuebersicht found %d records", len(items))
	return items, nil
}

func (w *MySQLWrapper) ListEilagerBuchungenByKZ(ctx context.Context, kz interface{}) ([]repo.Eilagerbuchung, error) {
	kzStr := toString(kz)
	log.Printf("[DB] ListEilagerBuchungenByKZ called for KZ: %s", kzStr)
	query := `SELECT id, id_fremdeslager, id_buchung, id_eilager, buchungsdatum, jumbos, xl, large, medium, small, volleikg, schmutz, knickeier, brucheier, buchungstyp, charge, kz_lager, id_fremdebuchung, verkauf FROM eilagerbuchung WHERE kz_lager = ? ORDER BY buchungsdatum DESC`
	rows, err := w.db.QueryContext(ctx, query, kzStr)
	if err != nil {
		log.Printf("[DB] ListEilagerBuchungenByKZ Query Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []repo.Eilagerbuchung
	for rows.Next() {
		var id, idf, idb, ide, j, xl, l, m, s, sm, ke, be, idfb, vk interface{}
		var v interface{}
		var bdat, btyp, char, kzl sql.NullString
		if err := rows.Scan(
			&id, &idf, &idb, &ide, &bdat, &j, &xl, &l, &m, &s, &v, &sm, &ke, &be, &btyp, &char, &kzl, &idfb, &vk,
		); err != nil {
			log.Printf("[DB] ListEilagerBuchungenByKZ Scan Error: %v", err)
			return nil, err
		}
		items = append(items, repo.Eilagerbuchung{
			ID:              toInt64(id),
			IDFremdeslager:  toInt64(idf),
			IDBuchung:       toInt64(idb),
			IDEilager:       toInt64(ide),
			Buchungsdatum:   bdat.String,
			Jumbos:          toInt64(j),
			Xl:              toInt64(xl),
			Large:           toInt64(l),
			Medium:          toInt64(m),
			Small:           toInt64(s),
			Volleikg:        toFloat(v),
			Schmutz:         toInt64(sm),
			Knickeier:       toInt64(ke),
			Brucheier:       toInt64(be),
			Buchungstyp:     btyp.String,
			Charge:          char.String,
			KzLager:         kzl.String,
			IDFremdebuchung: toInt64(idfb),
			Verkauf:         toInt64(vk),
		})
	}
	log.Printf("[DB] ListEilagerBuchungenByKZ found %d records", len(items))
	return items, nil
}

func (w *MySQLWrapper) ListEilagerBuchungenByLager(ctx context.Context, idEilager int64) ([]repo.Eilagerbuchung, error) {
	log.Printf("[DB] ListEilagerBuchungenByLager called for ID: %d", idEilager)
	query := `SELECT id, id_fremdeslager, id_buchung, id_eilager, buchungsdatum, jumbos, xl, large, medium, small, volleikg, schmutz, knickeier, brucheier, buchungstyp, charge, kz_lager, id_fremdebuchung, verkauf FROM eilagerbuchung WHERE id_eilager = ? ORDER BY buchungsdatum DESC`
	rows, err := w.db.QueryContext(ctx, query, idEilager)
	if err != nil {
		log.Printf("[DB] ListEilagerBuchungenByLager Query Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []repo.Eilagerbuchung
	for rows.Next() {
		var id, idf, idb, ide, j, xl, l, m, s, sm, ke, be, idfb, vk interface{}
		var v interface{}
		var bdat, btyp, char, kz sql.NullString
		if err := rows.Scan(
			&id, &idf, &idb, &ide, &bdat, &j, &xl, &l, &m, &s, &v, &sm, &ke, &be, &btyp, &char, &kz, &idfb, &vk,
		); err != nil {
			log.Printf("[DB] ListEilagerBuchungenByLager Scan Error: %v", err)
			return nil, err
		}
		items = append(items, repo.Eilagerbuchung{
			ID:              toInt64(id),
			IDFremdeslager:  toInt64(idf),
			IDBuchung:       toInt64(idb),
			IDEilager:       toInt64(ide),
			Buchungsdatum:   bdat.String,
			Jumbos:          toInt64(j),
			Xl:              toInt64(xl),
			Large:           toInt64(l),
			Medium:          toInt64(m),
			Small:           toInt64(s),
			Volleikg:        toFloat(v),
			Schmutz:         toInt64(sm),
			Knickeier:       toInt64(ke),
			Brucheier:       toInt64(be),
			Buchungstyp:     btyp.String,
			Charge:          char.String,
			KzLager:         kz.String,
			IDFremdebuchung: toInt64(idfb),
			Verkauf:         toInt64(vk),
		})
	}
	log.Printf("[DB] ListEilagerBuchungenByLager found %d records", len(items))
	return items, nil
}

func (w *MySQLWrapper) GetEilagerSumByBuchungID(ctx context.Context, idBuchung int64) (repo.GetEilagerSumByBuchungIDRow, error) {
	v, err := w.mysql.GetEilagerSumByBuchungID(ctx, int32(idBuchung))
	if err != nil {
		return repo.GetEilagerSumByBuchungIDRow{}, err
	}
	return repo.GetEilagerSumByBuchungIDRow{
		Jumbos:    v.Jumbos,
		Xl:        v.Xl,
		Large:     v.Large,
		Medium:    v.Medium,
		Small:     v.Small,
		Volleikg:  v.Volleikg,
		Schmutz:   v.Schmutz,
		Knickeier: v.Knickeier,
		Brucheier: v.Brucheier,
	}, nil
}

func (w *MySQLWrapper) GetEilagerSumBySource(ctx context.Context, arg repo.GetEilagerSumBySourceParams) (repo.GetEilagerSumBySourceRow, error) {
	v, err := w.mysql.GetEilagerSumBySource(ctx, repo_mysql.GetEilagerSumBySourceParams{
		IDBuchung:      toInt32(arg.IDBuchung),
		IDFremdeslager: toInt32(arg.IDFremdeslager),
	})
	if err != nil {
		return repo.GetEilagerSumBySourceRow{}, err
	}
	return repo.GetEilagerSumBySourceRow{
		Jumbos:   v.Jumbos,
		Xl:       v.Xl,
		Large:    v.Large,
		Medium:   v.Medium,
		Small:    v.Small,
		Volleikg: v.Volleikg,
	}, nil
}

func convertEilager(v repo_mysql.Eilager) repo.Eilager {
	return repo.Eilager{
		ID:            int64(v.ID),
		Lagernummer:   int64(v.Lagernummer),
		Kz:            v.Kz,
		Bezeichnung:   v.Bezeichnung,
		LetzteBuchung: v.LetzteBuchung,
		Jumbos:        int64(v.Jumbos),
		Xl:            int64(v.Xl),
		Large:         int64(v.Large),
		Medium:        int64(v.Medium),
		Small:         int64(v.Small),
		Volleikg:      v.Volleikg,
		Klasse6:       v.Klasse6,
		Klasse7:       v.Klasse7,
		Aw:            int64(v.Aw),
	}
}

func convertEilagerbuchung(v repo_mysql.Eilagerbuchung) repo.Eilagerbuchung {
	return repo.Eilagerbuchung{
		ID:              int64(v.ID),
		IDFremdeslager:  int64(v.IDFremdeslager),
		IDBuchung:       int64(v.IDBuchung),
		IDEilager:       int64(v.IDEilager),
		Buchungsdatum:   v.Buchungsdatum,
		Jumbos:          int64(v.Jumbos),
		Xl:              int64(v.Xl),
		Large:           int64(v.Large),
		Medium:          int64(v.Medium),
		Small:           int64(v.Small),
		Volleikg:        v.Volleikg,
		Schmutz:         int64(v.Schmutz),
		Knickeier:       int64(v.Knickeier),
		Brucheier:       int64(v.Brucheier),
		Buchungstyp:     v.Buchungstyp,
		Charge:          v.Charge,
		KzLager:         v.KzLager,
		IDFremdebuchung: int64(v.IDFremdebuchung),
		Verkauf:         int64(v.Verkauf),
	}
}

func (w *MySQLWrapper) GetEierpreis(ctx context.Context, id int64) (repo.Eierpreise, error) {
	v, err := w.mysql.GetEierpreis(ctx, int32(id))
	if err != nil {
		return repo.Eierpreise{}, err
	}
	return convertEierpreise(v), nil
}

func convertEierpreise(v repo_mysql.Eierpreise) repo.Eierpreise {
	return repo.Eierpreise{
		ID:            int64(v.ID),
		KzHaltungstyp: v.KzHaltungstyp,
		Eierklasse:    v.Eierklasse,
		GewichtVon:    v.GewichtVon,
		GewichtBis:    v.GewichtBis,
		PreisVon:      v.PreisVon,
		PreisBis:      v.PreisBis,
	}
}

// --- Silo Methoden ---

func (w *MySQLWrapper) ListSilos(ctx context.Context) ([]repo.Silo, error) {
	res, err := w.mysql.ListSilos(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.Silo, len(res))
	for i, v := range res {
		items[i] = repo.Silo{
			ID:                 int64(v.ID),
			Silonummer:         int64(v.Silonummer),
			Personennummer:     int64(v.Personennummer),
			IDLieferant:        int64(v.IDLieferant),
			Bezeichnung:        v.Bezeichnung,
			Inventurdatumalt:   v.Inventurdatumalt,
			Inventurdatumneu:   v.Inventurdatumneu,
			Maxfuellmenge:      int64(v.Maxfuellmenge),
			Minfuellmenge:      int64(v.Minfuellmenge),
			Inventurfuellmenge: int64(v.Inventurfuellmenge),
			Aw:                 int64(v.Aw),
		}
	}
	return items, nil
}

func (w *MySQLWrapper) CreateSilo(ctx context.Context, arg repo.CreateSiloParams) (repo.Silo, error) {
	res, err := w.mysql.CreateSilo(ctx, repo_mysql.CreateSiloParams{
		Silonummer:         int32(arg.Silonummer),
		Personennummer:     int32(arg.Personennummer),
		Bezeichnung:        arg.Bezeichnung,
		Inventurdatumalt:   arg.Inventurdatumalt,
		Inventurdatumneu:   arg.Inventurdatumneu,
		Maxfuellmenge:      int32(arg.Maxfuellmenge),
		Minfuellmenge:      int32(arg.Minfuellmenge),
		Inventurfuellmenge: int32(arg.Inventurfuellmenge),
		IDLieferant:        int32(arg.IDLieferant),
	})
	if err != nil {
		return repo.Silo{}, err
	}
	id, _ := res.LastInsertId()
	return w.GetSilo(ctx, id)
}

func (w *MySQLWrapper) GetSilo(ctx context.Context, id int64) (repo.Silo, error) {
	v, err := w.mysql.GetSilo(ctx, int32(id))
	if err != nil {
		return repo.Silo{}, err
	}
	return repo.Silo{
		ID:                 int64(v.ID),
		Silonummer:         int64(v.Silonummer),
		Personennummer:     int64(v.Personennummer),
		IDLieferant:        int64(v.IDLieferant),
		Bezeichnung:        v.Bezeichnung,
		Inventurdatumalt:   v.Inventurdatumalt,
		Inventurdatumneu:   v.Inventurdatumneu,
		Maxfuellmenge:      int64(v.Maxfuellmenge),
		Minfuellmenge:      int64(v.Minfuellmenge),
		Inventurfuellmenge: int64(v.Inventurfuellmenge),
		Aw:                 int64(v.Aw),
	}, nil
}

// --- Firmenparameter ---

func (w *MySQLWrapper) GetGlobalFirmenparameter(ctx context.Context) (repo.Firmenparameter, error) {
	v, err := w.mysql.GetGlobalFirmenparameter(ctx)
	if err != nil {
		return repo.Firmenparameter{}, err
	}
	return convertFirmenparameter(v), nil
}

func (w *MySQLWrapper) UpdateFirmenparameter(ctx context.Context, arg repo.UpdateFirmenparameterParams) (repo.Firmenparameter, error) {
	_, err := w.mysql.UpdateFirmenparameter(ctx, repo_mysql.UpdateFirmenparameterParams{
		IDHerden:                  int32(arg.IDHerden),
		Kz:                        toString(arg.Kz),
		Jumbos:                    int32(arg.Jumbos),
		Klassenerfassen:           int32(arg.Klassenerfassen),
		Klasseaerfassen:           int32(arg.Klasseaerfassen),
		Klasseaerrechnen:          int32(arg.Klasseaerrechnen),
		Klasseavermitteln:         int32(arg.Klasseavermitteln),
		Erfasseschmutzei:          int32(arg.Erfasseschmutzei),
		Erfasseknickei:            int32(arg.Erfasseknickei),
		Erfassebruchei:            int32(arg.Erfassebruchei),
		Erfassevollei:             int32(arg.Erfassevollei),
		Massvollei:                int32(arg.Massvollei),
		Aufteilunggewicht:         int32(arg.Aufteilunggewicht),
		Kontrollwiegung:           int32(arg.Kontrollwiegung),
		Anzahlkontrollw:           int32(arg.Anzahlkontrollw),
		Verpackungkg:              arg.Verpackungkg,
		Aufteilungalter:           int32(arg.Aufteilungalter),
		Erfassevolleikg:           int32(arg.Erfassevolleikg),
		Laufzeitwochen:            int32(arg.Laufzeitwochen),
		Zeitstempel:               arg.Zeitstempel,
		Schlachterloeshenne:       arg.Schlachterloeshenne,
		Produktionsdauer:          int32(arg.Produktionsdauer),
		IDTabellegewicht:          int32(arg.IDTabellegewicht),
		IDTabellealter:            int32(arg.IDTabellealter),
		LegebeginnLw:              int32(arg.LegebeginnLw),
		Verlustebeibuchung:        int32(arg.Verlustebeibuchung),
		Lagerbuchungbeibuchung:    int32(arg.Lagerbuchungbeibuchung),
		Maxtagevermitteln:         int32(arg.Maxtagevermitteln),
		Chargejumbos:              int32(arg.Chargejumbos),
		Chargexl:                  int32(arg.Chargexl),
		Chargemedium:              int32(arg.Chargemedium),
		Chargesmall:               int32(arg.Chargesmall),
		Chargelarge:               int32(arg.Chargelarge),
		Chargevollei:              int32(arg.Chargevollei),
		Chargeprefixfirma:         toString(arg.Chargeprefixfirma),
		Chargeprefixherdennummer:  int32(arg.Chargeprefixherdennummer),
		Chargedatum:               int32(arg.Chargedatum),
		Chargelagernummer:         int32(arg.Chargelagernummer),
		Chargetrennung:            toString(arg.Chargetrennung),
		Beivermittelndatumaktuell: int32(arg.Beivermittelndatumaktuell),
		Pseudolager:               int32(arg.Pseudolager),
		Bio:                       int32(arg.Bio),
		Haltungstyp:               toString(arg.Haltungstyp),
		Bioaufschlag:              toFloat(arg.Bioaufschlag),
	})
	if err != nil {
		return repo.Firmenparameter{}, err
	}
	return w.GetFirmenparameterByHerde(ctx, arg.IDHerden)
}

func (w *MySQLWrapper) GetFirmenparameterByHerde(ctx context.Context, idHerden int64) (repo.Firmenparameter, error) {
	v, err := w.mysql.GetFirmenparameterByHerde(ctx, int32(idHerden))
	if err != nil {
		return repo.Firmenparameter{}, err
	}
	return convertFirmenparameter(v), nil
}

func convertFirmenparameter(v repo_mysql.Firmenparameter) repo.Firmenparameter {
	return repo.Firmenparameter{
		ID:                        int64(v.ID),
		IDHerden:                  int64(v.IDHerden),
		Kz:                        v.Kz,
		Jumbos:                    int64(v.Jumbos),
		Klassenerfassen:           int64(v.Klassenerfassen),
		Klasseaerfassen:           int64(v.Klasseaerfassen),
		Klasseaerrechnen:          int64(v.Klasseaerrechnen),
		Klasseavermitteln:         int64(v.Klasseavermitteln),
		Erfasseschmutzei:          int64(v.Erfasseschmutzei),
		Erfasseknickei:            int64(v.Erfasseknickei),
		Erfassebruchei:            int64(v.Erfassebruchei),
		Erfassevollei:             int64(v.Erfassevollei),
		Massvollei:                int64(v.Massvollei),
		Aufteilunggewicht:         int64(v.Aufteilunggewicht),
		Kontrollwiegung:           int64(v.Kontrollwiegung),
		Anzahlkontrollw:           int64(v.Anzahlkontrollw),
		Verpackungkg:              v.Verpackungkg,
		Aufteilungalter:           int64(v.Aufteilungalter),
		Erfassevolleikg:           int64(v.Erfassevolleikg),
		Laufzeitwochen:            int64(v.Laufzeitwochen),
		Zeitstempel:               v.Zeitstempel,
		Schlachterloeshenne:       v.Schlachterloeshenne,
		Produktionsdauer:          int64(v.Produktionsdauer),
		IDTabellegewicht:          int64(v.IDTabellegewicht),
		IDTabellealter:            int64(v.IDTabellealter),
		LegebeginnLw:              int64(v.LegebeginnLw),
		Verlustebeibuchung:        int64(v.Verlustebeibuchung),
		Lagerbuchungbeibuchung:    int64(v.Lagerbuchungbeibuchung),
		Maxtagevermitteln:         int64(v.Maxtagevermitteln),
		Chargejumbos:              int64(v.Chargejumbos),
		Chargexl:                  int64(v.Chargexl),
		Chargemedium:              int64(v.Chargemedium),
		Chargesmall:               int64(v.Chargesmall),
		Chargelarge:               int64(v.Chargelarge),
		Chargevollei:              int64(v.Chargevollei),
		Chargeprefixfirma:         v.Chargeprefixfirma,
		Chargeprefixherdennummer:  int64(v.Chargeprefixherdennummer),
		Chargedatum:               int64(v.Chargedatum),
		Chargelagernummer:         int64(v.Chargelagernummer),
		Chargetrennung:            v.Chargetrennung,
		Beivermittelndatumaktuell: int64(v.Beivermittelndatumaktuell),
		Pseudolager:               int64(v.Pseudolager),
		Bio:                       int64(v.Bio),
		Haltungstyp:               v.Haltungstyp,
		Bioaufschlag:              v.Bioaufschlag,
		Aw:                        int64(v.Aw),
	}
}

// --- Person Methoden ---

func (w *MySQLWrapper) ListPersonen(ctx context.Context) ([]repo.Person, error) {
	res, err := w.mysql.ListPersonen(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.Person, len(res))
	for i, v := range res {
		items[i] = convertPerson(v)
	}
	return items, nil
}

func (w *MySQLWrapper) GetPerson(ctx context.Context, id int64) (repo.Person, error) {
	v, err := w.mysql.GetPerson(ctx, int32(id))
	if err != nil {
		return repo.Person{}, err
	}
	return convertPerson(v), nil
}

func (w *MySQLWrapper) CreatePerson(ctx context.Context, arg repo.CreatePersonParams) (repo.Person, error) {
	res, err := w.mysql.CreatePerson(ctx, repo_mysql.CreatePersonParams{
		IDTexte:        int32(arg.IDTexte),
		IDAnrede:       int32(arg.IDAnrede),
		Personennummer: int32(arg.Personennummer),
		Kz:             arg.Kz,
		Postfach:       arg.Postfach,
		Name:           arg.Name,
		Firma:          arg.Firma,
		Strasse:        arg.Strasse,
		Plz:            arg.Plz,
		Ort:            arg.Ort,
		Telefon:        arg.Telefon,
		Mobiltelephon:  arg.Mobiltelephon,
		Email:          arg.Email,
		Email2:         arg.Email2,
		Foto:           sql.NullString{String: string(arg.Foto), Valid: len(arg.Foto) > 0},
		Homepage:       arg.Homepage,
	})
	if err != nil {
		return repo.Person{}, err
	}
	id, _ := res.LastInsertId()
	return w.GetPerson(ctx, id)
}

func (w *MySQLWrapper) ListTabellenkopfByType(ctx context.Context, tabellentyp interface{}) ([]repo.Tabellenkopf, error) {
	log.Printf("[MARIADB-FIX] ListTabellenkopfByType called: %v", tabellentyp)
	rows, err := w.db.QueryContext(ctx, "SELECT id, tabellentyp, tabellennummer, bezeichnung, anlagedatum, datum FROM TABELLENKOPF WHERE tabellentyp = ?", tabellentyp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []repo.Tabellenkopf
	for rows.Next() {
		var i repo.Tabellenkopf
		var ttyp interface{}
		if err := rows.Scan(&i.ID, &ttyp, &i.Tabellennummer, &i.Bezeichnung, &i.Anlagedatum, &i.Datum); err != nil {
			return nil, err
		}
		i.Tabellentyp = toString(ttyp)
		items = append(items, i)
	}
	return items, nil
}

func (w *MySQLWrapper) GetGewichtByTabNumAndWeight(ctx context.Context, arg repo.GetGewichtByTabNumAndWeightParams) (repo.Gewichttabelle, error) {
	v, err := w.mysql.GetGewichtByTabNumAndWeight(ctx, repo_mysql.GetGewichtByTabNumAndWeightParams{
		Tabellennummer: int32(arg.Tabellennummer),
		ROUND:          arg.ROUND,
	})
	if err != nil {
		return repo.Gewichttabelle{}, err
	}
	return repo.Gewichttabelle{
		ID:             int64(v.ID),
		Tabellennummer: int64(v.Tabellennummer),
		Eigewicht:      v.Eigewicht,
		Klasse1:        float64(v.Klasse1),
		Klasse2:        float64(v.Klasse2),
		Klasse3:        float64(v.Klasse3),
		Klasse4:        float64(v.Klasse4),
		Klasse5:        float64(v.Klasse5),
		Klasse6:        float64(v.Klasse6),
		Klasse7:        float64(v.Klasse7),
	}, nil
}

func (w *MySQLWrapper) ListGewichtByTabNum(ctx context.Context, tabNum int64) ([]repo.Gewichttabelle, error) {
	res, err := w.mysql.ListGewichtByTabNum(ctx, int32(tabNum))
	if err != nil {
		return nil, err
	}
	items := make([]repo.Gewichttabelle, len(res))
	for i, v := range res {
		items[i] = repo.Gewichttabelle{
			ID:             int64(v.ID),
			Tabellennummer: int64(v.Tabellennummer),
			Eigewicht:      v.Eigewicht,
			Klasse1:        float64(v.Klasse1),
			Klasse2:        float64(v.Klasse2),
			Klasse3:        float64(v.Klasse3),
			Klasse4:        float64(v.Klasse4),
			Klasse5:        float64(v.Klasse5),
			Klasse6:        float64(v.Klasse6),
			Klasse7:        float64(v.Klasse7),
		}
	}
	return items, nil
}

func (w *MySQLWrapper) ListEierpreise(ctx context.Context) ([]repo.Eierpreise, error) {
	res, err := w.mysql.ListEierpreise(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.Eierpreise, len(res))
	for i, v := range res {
		items[i] = repo.Eierpreise{
			ID:            int64(v.ID),
			KzHaltungstyp: v.KzHaltungstyp,
			Eierklasse:    v.Eierklasse,
			GewichtVon:    v.GewichtVon,
			GewichtBis:    v.GewichtBis,
			PreisVon:      v.PreisVon,
			PreisBis:      v.PreisBis,
		}
	}
	return items, nil
}

func convertPerson(v repo_mysql.Person) repo.Person {
	return repo.Person{
		ID:             int64(v.ID),
		IDTexte:        int64(v.IDTexte),
		IDAnrede:       int64(v.IDAnrede),
		Personennummer: int64(v.Personennummer),
		Kz:             v.Kz,
		Postfach:       v.Postfach,
		Name:           v.Name,
		Firma:          v.Firma,
		Strasse:        v.Strasse,
		Plz:            v.Plz,
		Ort:            v.Ort,
		Telefon:        v.Telefon,
		Mobiltelephon:  v.Mobiltelephon,
		Email:          v.Email,
		Email2:         v.Email2,
		Foto:           []byte(v.Foto.String),
		Homepage:       v.Homepage,
	}
}

// --- Tierbewegungen Methoden ---

func (w *MySQLWrapper) ListTierbewegungen(ctx context.Context, spracheKz string) ([]repo.ListTierbewegungenRow, error) {
	log.Printf("[DB] ListTierbewegungen called with spracheKz: %s", spracheKz)
	query := `
		SELECT t.id,
			   coalesce(t.herdennummer, 0),
			   coalesce(t.id_buchung, 0),
			   coalesce(t.typ, 'x'),
			   coalesce(t.id_texte, 0),
			   coalesce(t.bewegungsdatum, '0001-01-01'),
			   coalesce(t.bewegungen, 0),
			   coalesce(t.id_herden_von, 0),
			   coalesce(t.id_herden_nach, 0),
			   coalesce(t.kosten, 0),
			   txt.text_typ_kz,
			   u.betreff      as grund_text,
			   h.bezeichnung  as herden_bezeichnung,
			   hv.bezeichnung                as herden_von_bezeichnung,
			   hn.bezeichnung                as herden_nach_bezeichnung
		FROM tierbewegungen t
				 LEFT JOIN texte txt on t.id_texte = txt.id
				 LEFT JOIN uebersetzungen u on txt.id = u.id_texte and u.sprache_kz = ?
				 LEFT JOIN herden h on t.herdennummer = h.herdennummer
				 LEFT JOIN herden hv on t.id_herden_von = hv.id
				 LEFT JOIN herden hn on t.id_herden_nach = hn.id
		ORDER BY t.bewegungsdatum DESC
	`
	rows, err := w.db.QueryContext(ctx, query, spracheKz)
	if err != nil {
		log.Printf("[DB] ListTierbewegungen Query Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []repo.ListTierbewegungenRow
	for rows.Next() {
		var i repo.ListTierbewegungenRow
		var id, hnr, idb, idt, hvon, hnach, bew interface{}
		var k interface{}
		var ttkz, gt, hb, hvb, hnb sql.NullString
		var bdat, typ interface{}
		if err := rows.Scan(
			&id, &hnr, &idb, &typ, &idt, &bdat,
			&bew, &hvon, &hnach, &k,
			&ttkz, &gt, &hb, &hvb, &hnb,
		); err != nil {
			log.Printf("[DB] ListTierbewegungen Scan Error: %v", err)
			return nil, err
		}
		i.ID = toInt64(id)
		i.Herdennummer = toInt64(hnr)
		i.IDBuchung = toInt64(idb)
		i.Typ = toString(typ)
		i.IDTexte = toInt64(idt)
		i.Bewegungsdatum = toString(bdat)
		i.Bewegungen = toInt64(bew)
		i.IDHerdenVon = toInt64(hvon)
		i.IDHerdenNach = toInt64(hnach)
		i.Kosten = toFloat(k)
		i.TextTypKz = ttkz
		i.GrundText = gt
		i.HerdenBezeichnung = hb
		i.HerdenVonBezeichnung = hvb
		i.HerdenNachBezeichnung = hnb
		items = append(items, i)
	}
	log.Printf("[DB] ListTierbewegungen found %d records", len(items))
	return items, nil
}

func (w *MySQLWrapper) GetTierbewegung(ctx context.Context, id int64) (repo.Tierbewegungen, error) {
	query := `SELECT id, herdennummer, id_buchung, typ, id_texte, bewegungsdatum, bewegungen, id_herden_von, id_herden_nach, kosten FROM tierbewegungen WHERE id = ?`
	row := w.db.QueryRowContext(ctx, query, id)
	var t repo.Tierbewegungen
	var hnr, idb, idt, hvon, hnach int32
	var kosten float64
	var bdat, typ string
	if err := row.Scan(&t.ID, &hnr, &idb, &typ, &idt, &bdat, &t.Bewegungen, &hvon, &hnach, &kosten); err != nil {
		return repo.Tierbewegungen{}, err
	}
	t.Herdennummer = sql.NullInt64{Int64: int64(hnr), Valid: true}
	t.IDBuchung = sql.NullInt64{Int64: int64(idb), Valid: true}
	t.Typ = typ
	t.IDTexte = sql.NullInt64{Int64: int64(idt), Valid: true}
	t.Bewegungsdatum = sql.NullString{String: bdat, Valid: true}
	t.IDHerdenVon = sql.NullInt64{Int64: int64(hvon), Valid: true}
	t.IDHerdenNach = sql.NullInt64{Int64: int64(hnach), Valid: true}
	t.Kosten = sql.NullFloat64{Float64: kosten, Valid: true}
	return t, nil
}

func (w *MySQLWrapper) CreateTierbewegung(ctx context.Context, arg repo.CreateTierbewegungParams) (repo.Tierbewegungen, error) {
	res, err := w.mysql.CreateTierbewegung(ctx, repo_mysql.CreateTierbewegungParams{
		Herdennummer:   toNullInt32(arg.Herdennummer),
		IDBuchung:      toNullInt32(arg.IDBuchung),
		Typ:            toNullString(arg.Typ),
		IDTexte:        toNullInt32(arg.IDTexte),
		Bewegungsdatum: toNullString(arg.Bewegungsdatum),
		Bewegungen:     toNullInt32(arg.Bewegungen),
		IDHerdenVon:    toNullInt32(arg.IDHerdenVon),
		IDHerdenNach:   toNullInt32(arg.IDHerdenNach),
		Kosten:         toNullString(arg.Kosten),
	})
	if err != nil {
		return repo.Tierbewegungen{}, err
	}
	id, _ := res.LastInsertId()
	return w.GetTierbewegung(ctx, id)
}

func (w *MySQLWrapper) UpdateTierbewegung(ctx context.Context, arg repo.UpdateTierbewegungParams) (repo.Tierbewegungen, error) {
	_, err := w.mysql.UpdateTierbewegung(ctx, repo_mysql.UpdateTierbewegungParams{
		Herdennummer:   toNullInt32(arg.Herdennummer),
		IDBuchung:      toNullInt32(arg.IDBuchung),
		Typ:            toNullString(arg.Typ),
		IDTexte:        toNullInt32(arg.IDTexte),
		Bewegungsdatum: toNullString(arg.Bewegungsdatum),
		Bewegungen:     toNullInt32(arg.Bewegungen),
		IDHerdenVon:    toNullInt32(arg.IDHerdenVon),
		IDHerdenNach:   toNullInt32(arg.IDHerdenNach),
		Kosten:         toNullString(arg.Kosten),
		ID:             int32(arg.ID),
	})
	if err != nil {
		return repo.Tierbewegungen{}, err
	}
	return w.GetTierbewegung(ctx, arg.ID)
}

func (w *MySQLWrapper) DeleteTierbewegung(ctx context.Context, id int64) error {
	return w.mysql.DeleteTierbewegung(ctx, int32(id))
}

// --- Lookup & Liste Methoden ---

func (w *MySQLWrapper) ListHerdenLookup(ctx context.Context) ([]repo.ListHerdenLookupRow, error) {
	res, err := w.mysql.ListHerdenLookup(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.ListHerdenLookupRow, len(res))
	for i, v := range res {
		items[i] = repo.ListHerdenLookupRow{
			ID:           int64(v.ID),
			Herdennummer: int64(v.Herdennummer),
			Bezeichnung:  v.Bezeichnung,
			Aktiv:        int64(v.Aktiv),
		}
	}
	return items, nil
}

func (w *MySQLWrapper) ListRassen(ctx context.Context) ([]repo.Rasse, error) {
	res, err := w.mysql.ListRassen(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.Rasse, len(res))
	for i, v := range res {
		items[i] = repo.Rasse{
			ID:    int64(v.ID),
			Rasse: v.Rasse,
			Aw:    int64(v.Aw),
		}
	}
	return items, nil
}

func (w *MySQLWrapper) ListStaelle(ctx context.Context) ([]repo.Stall, error) {
	res, err := w.mysql.ListStaelle(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.Stall, len(res))
	for i, v := range res {
		items[i] = repo.Stall{
			ID:          int64(v.ID),
			IDAlt:       int64(v.IDAlt),
			Stallnummer: int64(v.Stallnummer),
			Bezeichnung: v.Bezeichnung,
		}
	}
	return items, nil
}

func (w *MySQLWrapper) ListLieferanten(ctx context.Context) ([]repo.Person, error) {
	res, err := w.mysql.ListLieferanten(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.Person, len(res))
	for i, v := range res {
		items[i] = convertPerson(v)
	}
	return items, nil
}

func (w *MySQLWrapper) ListZuechter(ctx context.Context) ([]repo.Person, error) {
	res, err := w.mysql.ListZuechter(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.Person, len(res))
	for i, v := range res {
		items[i] = convertPerson(v)
	}
	return items, nil
}

// --- Texte & Übersetzungen ---

// (Duplicate removed)

func (w *MySQLWrapper) GetTranslatedText(ctx context.Context, arg repo.GetTranslatedTextParams) (repo.GetTranslatedTextRow, error) {
	v, err := w.mysql.GetTranslatedText(ctx, repo_mysql.GetTranslatedTextParams{
		SpracheKz: toString(arg.SpracheKz),
		ID:        int32(arg.ID),
	})
	if err != nil {
		return repo.GetTranslatedTextRow{}, err
	}
	return repo.GetTranslatedTextRow{
		ID:        int64(v.ID),
		TextTypKz: v.TextTypKz,
		Betreff:   v.Betreff,
		Inhalt:    v.Inhalt,
		SpracheKz: v.SpracheKz,
	}, nil
}

func (w *MySQLWrapper) ListFuttersorten(ctx context.Context) ([]repo.Futtersorten, error) {
	res, err := w.mysql.ListFuttersorten(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.Futtersorten, len(res))
	for i, v := range res {
		items[i] = repo.Futtersorten{
			ID:          int64(v.ID),
			Bezeichnung: v.Bezeichnung,
		}
	}
	return items, nil
}

func (w *MySQLWrapper) ListLagerplaetze(ctx context.Context) ([]repo.ListLagerplaetzeRow, error) {
	res, err := w.mysql.ListLagerplaetze(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.ListLagerplaetzeRow, len(res))
	for i, v := range res {
		items[i] = repo.ListLagerplaetzeRow{
			ID:                 int64(v.ID),
			IDEilager:          int64(v.IDEilager),
			Bezeichnung:        v.Bezeichnung,
			Bemerkung:          v.Bemerkung,
			EilagerBezeichnung: v.EilagerBezeichnung,
		}
	}
	return items, nil
}

func (w *MySQLWrapper) ListLagerplaetzeByEilager(ctx context.Context, idEilager int64) ([]repo.ListLagerplaetzeByEilagerRow, error) {
	res, err := w.mysql.ListLagerplaetzeByEilager(ctx, int32(idEilager))
	if err != nil {
		return nil, err
	}
	items := make([]repo.ListLagerplaetzeByEilagerRow, len(res))
	for i, v := range res {
		items[i] = repo.ListLagerplaetzeByEilagerRow{
			ID:          int64(v.ID),
			Bezeichnung: v.Bezeichnung,
		}
	}
	return items, nil
}

func (w *MySQLWrapper) UpdateEierpreis(ctx context.Context, arg repo.UpdateEierpreisParams) (repo.Eierpreise, error) {
	_, err := w.mysql.UpdateEierpreis(ctx, repo_mysql.UpdateEierpreisParams{
		Eierklasse: arg.Eierklasse,
		PreisVon:   toFloat(arg.PreisVon),
		ID:         int32(arg.ID),
	})
	if err != nil {
		return repo.Eierpreise{}, err
	}
	return w.GetEierpreis(ctx, arg.ID)
}

func (w *MySQLWrapper) ListVerkauf(ctx context.Context) ([]repo.Verkauf, error) {
	log.Printf("[DB] ListVerkauf called")
	query := `SELECT id, id_eilagerbuchung, id_buchung, buchungsdatum, mengesmall, mengemedium, mengelarge, mengexl, preissmall, preismedium, preislarge, preisxl, gesamtpreis, bio, verbucht, charge, rabattprozent FROM verkauf ORDER BY buchungsdatum DESC`
	rows, err := w.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("[DB] ListVerkauf Query Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []repo.Verkauf
	for rows.Next() {
		var v repo.Verkauf
		var id, ideb, idb, ms, mm, ml, mxl int32
		var ps, pm, pl, pxl, gp, rp interface{}
		if err := rows.Scan(
			&id, &ideb, &idb, &v.Buchungsdatum,
			&ms, &mm, &ml, &mxl,
			&ps, &pm, &pl, &pxl, &gp,
			&v.Bio, &v.Verbucht, &v.Charge, &rp,
		); err != nil {
			log.Printf("[DB] ListVerkauf Scan Error: %v", err)
			return nil, err
		}
		v.ID = int64(id)
		v.IDEilagerbuchung = int64(ideb)
		v.IDBuchung = int64(idb)
		v.Mengesmall = int64(ms)
		v.Mengemedium = int64(mm)
		v.Mengelarge = int64(ml)
		v.Mengexl = int64(mxl)
		v.Preissmall = toFloat(ps)
		v.Preismedium = toFloat(pm)
		v.Preislarge = toFloat(pl)
		v.Preisxl = toFloat(pxl)
		v.Gesamtpreis = toFloat(gp)
		v.Rabattprozent = toFloat(rp)
		items = append(items, v)
	}
	log.Printf("[DB] ListVerkauf found %d records", len(items))
	return items, nil
}

func (w *MySQLWrapper) GetVerkauf(ctx context.Context, id int64) (repo.Verkauf, error) {
	v, err := w.mysql.GetVerkauf(ctx, int32(id))
	if err != nil {
		return repo.Verkauf{}, err
	}
	return convertVerkauf(v), nil
}

func (w *MySQLWrapper) CreateVerkauf(ctx context.Context, arg repo.CreateVerkaufParams) (repo.Verkauf, error) {
	res, err := w.mysql.CreateVerkauf(ctx, repo_mysql.CreateVerkaufParams{
		IDEilagerbuchung: int32(arg.IDEilagerbuchung),
		IDBuchung:        int32(arg.IDBuchung),
		Buchungsdatum:    arg.Buchungsdatum,
		Mengesmall:       int32(arg.Mengesmall),
		Mengemedium:      int32(arg.Mengemedium),
		Mengelarge:       int32(arg.Mengelarge),
		Mengexl:          int32(arg.Mengexl),
		Preissmall:       arg.Preissmall,
		Preismedium:      arg.Preismedium,
		Preislarge:       arg.Preislarge,
		Preisxl:          arg.Preisxl,
		Gesamtpreis:      arg.Gesamtpreis,
		Bio:              arg.Bio,
		Verbucht:         arg.Verbucht,
		Charge:           toString(arg.Charge),
		Rabattprozent:    toFloat(arg.Rabattprozent),
	})
	if err != nil {
		return repo.Verkauf{}, err
	}
	id, _ := res.LastInsertId()
	return w.GetVerkauf(ctx, id)
}

func (w *MySQLWrapper) UpdateVerkauf(ctx context.Context, arg repo.UpdateVerkaufParams) (repo.Verkauf, error) {
	_, err := w.mysql.UpdateVerkauf(ctx, repo_mysql.UpdateVerkaufParams{
		IDEilagerbuchung: int32(arg.IDEilagerbuchung),
		IDBuchung:        int32(arg.IDBuchung),
		Buchungsdatum:    arg.Buchungsdatum,
		Mengesmall:       int32(arg.Mengesmall),
		Mengemedium:      int32(arg.Mengemedium),
		Mengelarge:       int32(arg.Mengelarge),
		Mengexl:          int32(arg.Mengexl),
		Preissmall:       arg.Preissmall,
		Preismedium:      arg.Preismedium,
		Preislarge:       arg.Preislarge,
		Preisxl:          arg.Preisxl,
		Gesamtpreis:      arg.Gesamtpreis,
		Bio:              arg.Bio,
		Verbucht:         arg.Verbucht,
		Charge:           toString(arg.Charge),
		Rabattprozent:    toFloat(arg.Rabattprozent),
		ID:               int32(arg.ID),
	})
	if err != nil {
		return repo.Verkauf{}, err
	}
	return w.GetVerkauf(ctx, arg.ID)
}

func convertVerkauf(v repo_mysql.Verkauf) repo.Verkauf {
	return repo.Verkauf{
		ID:               int64(v.ID),
		IDEilagerbuchung: int64(v.IDEilagerbuchung),
		IDBuchung:        int64(v.IDBuchung),
		Buchungsdatum:    v.Buchungsdatum,
		Mengesmall:       int64(v.Mengesmall),
		Mengemedium:      int64(v.Mengemedium),
		Mengelarge:       int64(v.Mengelarge),
		Mengexl:          int64(v.Mengexl),
		Preissmall:       v.Preissmall,
		Preismedium:      v.Preismedium,
		Preislarge:       v.Preislarge,
		Preisxl:          v.Preisxl,
		Gesamtpreis:      v.Gesamtpreis,
		Bio:              v.Bio,
		Verbucht:         v.Verbucht,
		Charge:           v.Charge,
		Rabattprozent:    v.Rabattprozent,
	}
}
func (w *MySQLWrapper) ListDynamischeSQL(ctx context.Context) ([]repo.ListDynamischeSQLRow, error) {
	log.Printf("[MARIADB-FIX] ListDynamischeSQL called")
	query := `SELECT id, beschreibung, sqlstatement, kategorie_kz, gruppen_kz, typ_kz, system_kz, template_name, param_def, detail_sql, link_logic, group_field, rows_per_page, page_orientation, show_master_grid, show_detail_grid, sqlstatement_native, detail_sql_native, root_kz, summenzeile, ist_summenzeile FROM DYNAMISCHE_SQL ORDER BY beschreibung`
	rows, err := w.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("[DB] ListDynamischeSQL Query Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []repo.ListDynamischeSQLRow
	for rows.Next() {
		var i repo.ListDynamischeSQLRow
		var kkz, gkz, tkz, rkz interface{}
		if err := rows.Scan(
			&i.ID, &i.Beschreibung, &i.Sqlstatement, &kkz, &gkz, &tkz, &i.SystemKz, &i.TemplateName, &i.ParamDef, &i.DetailSql, &i.LinkLogic, &i.GroupField, &i.RowsPerPage, &i.PageOrientation, &i.ShowMasterGrid, &i.ShowDetailGrid, &i.SqlstatementNative, &i.DetailSqlNative, &rkz, &i.Summenzeile, &i.IstSummenzeile,
		); err != nil {
			log.Printf("[DB] ListDynamischeSQL Scan Error: %v", err)
			return nil, err
		}
		i.KategorieKz = toString(kkz)
		i.GruppenKz = toString(gkz)
		i.TypKz = toString(tkz)
		i.RootKz = toString(rkz)
		items = append(items, i)
	}
	log.Printf("[DB] ListDynamischeSQL found %d records", len(items))
	return items, nil
}

func (w *MySQLWrapper) GetDynamischeSQL(ctx context.Context, id int64) (repo.DynamischeSql, error) {
	log.Printf("[DB] GetDynamischeSQL called for ID: %d", id)
	query := `SELECT id, beschreibung, sqlstatement, kategorie_kz, gruppen_kz, typ_kz, template_name, param_def, detail_sql, link_logic, group_field, rows_per_page, page_orientation, show_master_grid, show_detail_grid, system_kz, sqlstatement_native, detail_sql_native, root_kz, summenzeile, ist_summenzeile FROM DYNAMISCHE_SQL WHERE id = ?`
	var i repo.DynamischeSql
	var kat, grp, typ, sys, rkz interface{}
	err := w.db.QueryRowContext(ctx, query, id).Scan(
		&i.ID, &i.Beschreibung, &i.Sqlstatement, &kat, &grp, &typ, &i.TemplateName, &i.ParamDef, &i.DetailSql, &i.LinkLogic, &i.GroupField, &i.RowsPerPage, &i.PageOrientation, &i.ShowMasterGrid, &i.ShowDetailGrid, &sys, &i.SqlstatementNative, &i.DetailSqlNative, &rkz, &i.Summenzeile, &i.IstSummenzeile,
	)
	if err != nil {
		log.Printf("[DB] GetDynamischeSQL Error: %v", err)
		return i, err
	}
	i.KategorieKz = toString(kat)
	i.GruppenKz = toString(grp)
	i.TypKz = toString(typ)
	i.SystemKz = toString(sys)
	i.RootKz = toString(rkz)
	return i, nil
}

func (w *MySQLWrapper) CreateTabellenkopf(ctx context.Context, arg repo.CreateTabellenkopfParams) (repo.Tabellenkopf, error) {
	log.Printf("[DB] CreateTabellenkopf called for Typ: %s, Nr: %d", arg.Tabellentyp, arg.Tabellennummer)
	query := `INSERT INTO TABELLENKOPF (TABELLENTYP, TABELLENNUMMER, BEZEICHNUNG, ANLAGEDATUM, DATUM) VALUES (?, ?, ?, ?, ?)`
	res, err := w.db.ExecContext(ctx, query, arg.Tabellentyp, arg.Tabellennummer, arg.Bezeichnung, arg.Anlagedatum, arg.Datum)
	if err != nil {
		log.Printf("[DB] CreateTabellenkopf Error: %v", err)
		return repo.Tabellenkopf{}, err
	}
	id, _ := res.LastInsertId()
	return repo.Tabellenkopf{
		ID:             id,
		Tabellentyp:    toString(arg.Tabellentyp),
		Tabellennummer: toInt64(arg.Tabellennummer),
		Bezeichnung:    arg.Bezeichnung,
		Anlagedatum:    arg.Anlagedatum,
		Datum:          arg.Datum,
	}, nil
}

func (w *MySQLWrapper) UpdateTabellenkopf(ctx context.Context, arg repo.UpdateTabellenkopfParams) (repo.Tabellenkopf, error) {
	log.Printf("[DB] UpdateTabellenkopf called for ID: %d", arg.ID)
	query := `UPDATE TABELLENKOPF SET TABELLENNUMMER = ?, BEZEICHNUNG = ?, ANLAGEDATUM = ?, DATUM = ? WHERE ID = ?`
	_, err := w.db.ExecContext(ctx, query, arg.Tabellennummer, arg.Bezeichnung, arg.Anlagedatum, arg.Datum, arg.ID)
	if err != nil {
		log.Printf("[DB] UpdateTabellenkopf Error: %v", err)
		return repo.Tabellenkopf{}, err
	}
	// Reload to get the full object including Tabellentyp which isn't in params
	var t repo.Tabellenkopf
	var ttyp interface{}
	err = w.db.QueryRowContext(ctx, "SELECT ID, TABELLENTYP, TABELLENNUMMER, BEZEICHNUNG, ANLAGEDATUM, DATUM FROM TABELLENKOPF WHERE ID = ?", arg.ID).Scan(
		&t.ID, &ttyp, &t.Tabellennummer, &t.Bezeichnung, &t.Anlagedatum, &t.Datum,
	)
	t.Tabellentyp = toString(ttyp)
	return t, err
}

func (w *MySQLWrapper) ListTextTypen(ctx context.Context) ([]repo.TextTypen, error) {
	log.Printf("[DB] ListTextTypen called")
	rows, err := w.db.QueryContext(ctx, "SELECT ID, KZ, BEZEICHNUNG, `SYSTEM_KZ`, `STATUS` FROM TEXT_TYPEN")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []repo.TextTypen
	for rows.Next() {
		var i repo.TextTypen
		if err := rows.Scan(&i.ID, &i.Kz, &i.Bezeichnung, &i.SystemKz, &i.Status); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (w *MySQLWrapper) CreateTextTyp(ctx context.Context, arg repo.CreateTextTypParams) (repo.TextTypen, error) {
	log.Printf("[DB] CreateTextTyp called: %s", arg.Kz)
	query := "INSERT INTO TEXT_TYPEN (KZ, BEZEICHNUNG, `SYSTEM_KZ`, `STATUS`) VALUES (?, ?, ?, ?)"
	res, err := w.db.ExecContext(ctx, query, arg.Kz, arg.Bezeichnung, arg.SystemKz, arg.Status)
	if err != nil {
		log.Printf("[DB] CreateTextTyp Error: %v", err)
		return repo.TextTypen{}, err
	}
	id, _ := res.LastInsertId()
	return repo.TextTypen{
		ID:          id,
		Kz:          arg.Kz,
		Bezeichnung: arg.Bezeichnung,
		SystemKz:    arg.SystemKz,
		Status:      arg.Status,
	}, nil
}

func (w *MySQLWrapper) UpdateTextTyp(ctx context.Context, arg repo.UpdateTextTypParams) (repo.TextTypen, error) {
	log.Printf("[DB] UpdateTextTyp called: %d", arg.ID)
	query := "UPDATE TEXT_TYPEN SET KZ = ?, BEZEICHNUNG = ?, `SYSTEM_KZ` = ?, `STATUS` = ? WHERE ID = ?"
	_, err := w.db.ExecContext(ctx, query, arg.Kz, arg.Bezeichnung, arg.SystemKz, arg.Status, arg.ID)
	if err != nil {
		log.Printf("[DB] UpdateTextTyp Error: %v", err)
		return repo.TextTypen{}, err
	}
	var t repo.TextTypen
	err = w.db.QueryRowContext(ctx, "SELECT ID, KZ, BEZEICHNUNG, `SYSTEM_KZ`, `STATUS` FROM TEXT_TYPEN WHERE ID = ?", arg.ID).Scan(
		&t.ID, &t.Kz, &t.Bezeichnung, &t.SystemKz, &t.Status,
	)
	return t, err
}

func (w *MySQLWrapper) CreateText(ctx context.Context, arg repo.CreateTextParams) (repo.Texte, error) {
	log.Printf("[DB] CreateText called for Typ: %s", arg.TextTypKz)
	query := "INSERT INTO TEXTE (TEXT_TYP_KZ, KZ, `SYSTEM_KZ`, `STATUS`) VALUES (?, ?, ?, ?)"
	res, err := w.db.ExecContext(ctx, query, arg.TextTypKz, arg.Kz, arg.SystemKz, arg.Status)
	if err != nil {
		log.Printf("[DB] CreateText Error: %v", err)
		return repo.Texte{}, err
	}
	id, _ := res.LastInsertId()
	return repo.Texte{
		ID:        id,
		TextTypKz: arg.TextTypKz,
		Kz:        arg.Kz,
		SystemKz:  arg.SystemKz,
		Status:    arg.Status,
	}, nil
}

func (w *MySQLWrapper) UpdateText(ctx context.Context, arg repo.UpdateTextParams) (repo.Texte, error) {
	log.Printf("[DB] UpdateText called: %d", arg.ID)
	query := "UPDATE TEXTE SET TEXT_TYP_KZ = ?, KZ = ?, `SYSTEM_KZ` = ?, `STATUS` = ? WHERE ID = ?"
	_, err := w.db.ExecContext(ctx, query, arg.TextTypKz, arg.Kz, arg.SystemKz, arg.Status, arg.ID)
	if err != nil {
		log.Printf("[DB] UpdateText Error: %v", err)
		return repo.Texte{}, err
	}
	var t repo.Texte
	var kz interface{}
	err = w.db.QueryRowContext(ctx, "SELECT ID, TEXT_TYP_KZ, KZ, `SYSTEM_KZ`, `STATUS` FROM TEXTE WHERE ID = ?", arg.ID).Scan(
		&t.ID, &t.TextTypKz, &kz, &t.SystemKz, &t.Status,
	)
	t.Kz = toString(kz)
	return t, err
}

func (w *MySQLWrapper) CreateUebersetzung(ctx context.Context, arg repo.CreateUebersetzungParams) (repo.Uebersetzungen, error) {
	log.Printf("[DB] CreateUebersetzung called for ID: %d, Lang: %s", arg.IDTexte, arg.SpracheKz)
	query := "INSERT INTO UEBERSETZUNGEN (ID_TEXTE, SPRACHE_KZ, BETREFF, INHALT) VALUES (?, ?, ?, ?)"
	_, err := w.db.ExecContext(ctx, query, arg.IDTexte, arg.SpracheKz, arg.Betreff, arg.Inhalt)
	if err != nil {
		log.Printf("[DB] CreateUebersetzung Error: %v", err)
		return repo.Uebersetzungen{}, err
	}
	return repo.Uebersetzungen{
		IDTexte:   arg.IDTexte,
		SpracheKz: arg.SpracheKz,
		Betreff:   arg.Betreff,
		Inhalt:    arg.Inhalt,
	}, nil
}

func (w *MySQLWrapper) UpsertUebersetzung(ctx context.Context, arg repo.UpsertUebersetzungParams) (repo.Uebersetzungen, error) {
	log.Printf("[DB] UpsertUebersetzung called for ID: %d, Lang: %s", arg.IDTexte, arg.SpracheKz)
	query := `INSERT INTO UEBERSETZUNGEN (ID_TEXTE, SPRACHE_KZ, BETREFF, INHALT) 
              VALUES (?, ?, ?, ?) 
              ON DUPLICATE KEY UPDATE BETREFF = VALUES(BETREFF), INHALT = VALUES(INHALT)`
	_, err := w.db.ExecContext(ctx, query, arg.IDTexte, arg.SpracheKz, arg.Betreff, arg.Inhalt)
	if err != nil {
		log.Printf("[DB] UpsertUebersetzung Error: %v", err)
		return repo.Uebersetzungen{}, err
	}
	return repo.Uebersetzungen{
		IDTexte:   arg.IDTexte,
		SpracheKz: arg.SpracheKz,
		Betreff:   arg.Betreff,
		Inhalt:    arg.Inhalt,
	}, nil
}

func (w *MySQLWrapper) ListTexte(ctx context.Context, spracheKz string) ([]repo.ListTexteRow, error) {
	log.Printf("[DB] ListTexte called for Lang: %s", spracheKz)
	query := `SELECT T.ID, T.TEXT_TYP_KZ, T.KZ, T.` + "`SYSTEM_KZ`" + `, T.` + "`STATUS`" + `, COALESCE(U.BETREFF, '') AS BETREFF, COALESCE(U.INHALT, '') AS INHALT 
              FROM TEXTE T 
              LEFT JOIN UEBERSETZUNGEN U ON T.ID = U.ID_TEXTE AND U.SPRACHE_KZ = ?`
	rows, err := w.db.QueryContext(ctx, query, spracheKz)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []repo.ListTexteRow
	for rows.Next() {
		var i repo.ListTexteRow
		var kz interface{}
		if err := rows.Scan(&i.ID, &i.TextTypKz, &kz, &i.SystemKz, &i.Status, &i.Betreff, &i.Inhalt); err != nil {
			return nil, err
		}
		i.Kz = SanitizeKZ(toString(kz))
		items = append(items, i)
	}
	return items, nil
}

func (w *MySQLWrapper) ListTexteByType(ctx context.Context, arg repo.ListTexteByTypeParams) ([]repo.ListTexteByTypeRow, error) {
	log.Printf("[DB] ListTexteByType called for Typ: %s, Lang: %s", arg.TextTypKz, arg.SpracheKz)
	query := `SELECT T.ID, T.TEXT_TYP_KZ, T.KZ, T.` + "`SYSTEM_KZ`" + `, T.` + "`STATUS`" + `, COALESCE(U.BETREFF, '') AS BETREFF, COALESCE(U.INHALT, '') AS INHALT 
              FROM TEXTE T 
              LEFT JOIN UEBERSETZUNGEN U ON T.ID = U.ID_TEXTE AND U.SPRACHE_KZ = ? 
              WHERE T.TEXT_TYP_KZ = ?`
	rows, err := w.db.QueryContext(ctx, query, arg.SpracheKz, arg.TextTypKz)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []repo.ListTexteByTypeRow
	for rows.Next() {
		var i repo.ListTexteByTypeRow
		var kz interface{}
		if err := rows.Scan(&i.ID, &i.TextTypKz, &kz, &i.SystemKz, &i.Status, &i.Betreff, &i.Inhalt); err != nil {
			return nil, err
		}
		i.Kz = SanitizeKZ(toString(kz))
		items = append(items, i)
	}
	return items, nil
}

func (w *MySQLWrapper) ListBenutzerProfile(ctx context.Context) ([]repo.Benutzerprofile, error) {
	log.Printf("[MARIADB-FIX] ListBenutzerProfile called")
	rows, err := w.db.QueryContext(ctx, "SELECT id, profil_kz, beschreibung, f_dashboard, f_herden_verwalten, f_einrichtungen_verwalten, f_personen_verwalten, f_buchungen_erfassen, f_auswertungen_anzeigen, f_sql_struktur_verwalten, f_benutzer_profile, f_parameter_editieren, f_kosten_verwalten, f_tabellen_anzeigen, f_texte_verwalten, f_system_verwaltung, f_backup_erstellen FROM BENUTZERPROFILE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []repo.Benutzerprofile
	for rows.Next() {
		var i repo.Benutzerprofile
		var pkz interface{}
		err := rows.Scan(&i.ID, &pkz, &i.Beschreibung, &i.FDashboard, &i.FHerdenVerwalten, &i.FEinrichtungenVerwalten, &i.FPersonenVerwalten, &i.FBuchungenErfassen, &i.FAuswertungenAnzeigen, &i.FSqlStrukturVerwalten, &i.FBenutzerProfile, &i.FParameterEditieren, &i.FKostenVerwalten, &i.FTabellenAnzeigen, &i.FTexteVerwalten, &i.FSystemVerwaltung, &i.FBackupErstellen)
		if err != nil {
			return nil, err
		}
		i.ProfilKz = toString(pkz)
		items = append(items, i)
	}
	return items, nil
}

func (w *MySQLWrapper) GetBenutzerProfilByID(ctx context.Context, id int64) (repo.Benutzerprofile, error) {
	log.Printf("[MARIADB-FIX] GetBenutzerProfilByID called: %d", id)
	var i repo.Benutzerprofile
	var pkz interface{}
	err := w.db.QueryRowContext(ctx, "SELECT id, profil_kz, beschreibung, f_dashboard, f_herden_verwalten, f_einrichtungen_verwalten, f_personen_verwalten, f_buchungen_erfassen, f_auswertungen_anzeigen, f_sql_struktur_verwalten, f_benutzer_profile, f_parameter_editieren, f_kosten_verwalten, f_tabellen_anzeigen, f_texte_verwalten, f_system_verwaltung, f_backup_erstellen FROM BENUTZERPROFILE WHERE id = ?", id).Scan(
		&i.ID, &pkz, &i.Beschreibung, &i.FDashboard, &i.FHerdenVerwalten, &i.FEinrichtungenVerwalten, &i.FPersonenVerwalten, &i.FBuchungenErfassen, &i.FAuswertungenAnzeigen, &i.FSqlStrukturVerwalten, &i.FBenutzerProfile, &i.FParameterEditieren, &i.FKostenVerwalten, &i.FTabellenAnzeigen, &i.FTexteVerwalten, &i.FSystemVerwaltung, &i.FBackupErstellen,
	)
	i.ProfilKz = toString(pkz)
	return i, err
}

func (w *MySQLWrapper) GetBenutzerProfilByKZ(ctx context.Context, pkz interface{}) (repo.Benutzerprofile, error) {
	log.Printf("[MARIADB-FIX] GetBenutzerProfilByKZ called: %v", pkz)
	var i repo.Benutzerprofile
	var pkzResult interface{}
	err := w.db.QueryRowContext(ctx, "SELECT id, profil_kz, beschreibung, f_dashboard, f_herden_verwalten, f_einrichtungen_verwalten, f_personen_verwalten, f_buchungen_erfassen, f_auswertungen_anzeigen, f_sql_struktur_verwalten, f_benutzer_profile, f_parameter_editieren, f_kosten_verwalten, f_tabellen_anzeigen, f_texte_verwalten, f_system_verwaltung, f_backup_erstellen FROM BENUTZERPROFILE WHERE profil_kz = ?", pkz).Scan(
		&i.ID, &pkzResult, &i.Beschreibung, &i.FDashboard, &i.FHerdenVerwalten, &i.FEinrichtungenVerwalten, &i.FPersonenVerwalten, &i.FBuchungenErfassen, &i.FAuswertungenAnzeigen, &i.FSqlStrukturVerwalten, &i.FBenutzerProfile, &i.FParameterEditieren, &i.FKostenVerwalten, &i.FTabellenAnzeigen, &i.FTexteVerwalten, &i.FSystemVerwaltung, &i.FBackupErstellen,
	)
	i.ProfilKz = toString(pkzResult)
	return i, err
}

func (w *MySQLWrapper) ListBenutzer(ctx context.Context) ([]repo.ListBenutzerRow, error) {
	log.Printf("[MARIADB-FIX] ListBenutzer called")
	query := `SELECT B.ID, B.USERNAME, B.PASSWORT, B.KLARNAME, B.ID_BENUTZER_PROFILE, P.PROFIL_KZ 
              FROM BENUTZER B 
              LEFT JOIN BENUTZERPROFILE P ON B.ID_BENUTZER_PROFILE = P.ID`
	rows, err := w.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []repo.ListBenutzerRow
	for rows.Next() {
		var i repo.ListBenutzerRow
		var pkz interface{}
		if err := rows.Scan(&i.ID, &i.Username, &i.Passwort, &i.Klarname, &i.IDBenutzerProfile, &pkz); err != nil {
			return nil, err
		}
		i.ProfilKz = toString(pkz)
		items = append(items, i)
	}
	return items, nil
}

func (w *MySQLWrapper) ListShowTV(ctx context.Context) ([]repo.Showtv, error) {
	log.Printf("[MARIADB-FIX] ListShowTV called")
	rows, err := w.db.QueryContext(ctx, "SELECT id, tvname, showit FROM SHOWTV")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []repo.Showtv
	for rows.Next() {
		var i repo.Showtv
		var name interface{}
		if err := rows.Scan(&i.ID, &name, &i.Showit); err != nil {
			return nil, err
		}
		i.Tvname = toString(name)
		items = append(items, i)
	}
	return items, nil
}
