package db

import (
	"context"
	"database/sql"
	"huhnlite-wails/backend/db/repo"
	"huhnlite-wails/backend/db/repo_postgres"
	"log"
	"strings"
)

// PostgresWrapper implementiert das Querier-Interface, nutzt aber intern das MySQL-Repo
type PostgresWrapper struct {
	*repo.Queries
	pg *repo_postgres.Queries
	db    *sql.DB
	tx    *sql.Tx
}

func NewPostgresWrapper(sqlite *repo.Queries, pg *repo_postgres.Queries, db *sql.DB) *PostgresWrapper {
	return &PostgresWrapper{
		Queries: sqlite,
		pg:      pg,
		db:      db,
	}
}

func (w *PostgresWrapper) WithTx(tx *sql.Tx) *PostgresWrapper {
	return &PostgresWrapper{
		Queries: w.Queries.WithTx(tx),
		pg:   w.pg.WithTx(tx),
		db:      w.db,
		tx:      tx,
	}
}

func (w *PostgresWrapper) queryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	log.Printf("[POSTGRES-DEBUG] QueryRow: %s | Args: %v", strings.TrimSpace(query), args)
	if w.tx != nil {
		return w.tx.QueryRowContext(ctx, query, args...)
	}
	return w.db.QueryRowContext(ctx, query, args...)
}

func (w *PostgresWrapper) query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	log.Printf("[POSTGRES-DEBUG] Query: %s | Args: %v", strings.TrimSpace(query), args)
	var rows *sql.Rows
	var err error
	if w.tx != nil {
		rows, err = w.tx.QueryContext(ctx, query, args...)
	} else {
		rows, err = w.db.QueryContext(ctx, query, args...)
	}
	if err != nil {
		log.Printf("[POSTGRES-DEBUG] Query Error: %v", err)
	}
	return rows, err
}

// --- Buchung (Leistung) Methoden ---
func (w *PostgresWrapper) ListBuchungen(ctx context.Context) ([]repo.ListBuchungenRow, error) {
	query := `
		SELECT B.ID, B.ID_HERDEN, B.LW, B.HERDENNUMMER, B.BUCHUNGSDATUM, B.GEWICHTPROBE, B.KONTROLLGEWICHT, 
		       B.KLASSEA, B.VERLUSTE, B.EIMASSE, B.SCHMUTZ, B.KNICKEIER, B.VOLLEI, B.BRUCHEIER, 
		       B.TIERBESTAND, B.ID_EITABELLE, B.ID_DGEWICHTTAB, B.FUTTERKTAG, B.SILONR, B.KL6, 
		       B.VERMITTELTAM, B.SMALL, B.LARGE, B.MEDIUM, B.XL, B.ZEITSTEMPEL, B.DGEWICHTEI, B.AW, B.VERMITTELT, B.FUTTERVERBRAUCHTIER,
		       H.HERDENNUMMER AS HERDEN_NUMMER_REL, H.BEZEICHNUNG AS HERDEN_BEZEICHNUNG_REL, 
		       H.ID_EILAGER AS HERDEN_ID_EILAGER, H.AKTIV AS HERDEN_AKTIV_REL
		FROM BUCHUNG B
		LEFT JOIN HERDEN H ON B.ID_HERDEN = H.ID
		WHERE B.VERMITTELT != 'S'
		ORDER BY B.BUCHUNGSDATUM DESC
	`
	rows, err := w.query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []repo.ListBuchungenRow
	for rows.Next() {
		var i repo.ListBuchungenRow
		var id, idh, lw, hnr, ka, v, s, k, vo, b, tb, idet, idgt, sn, sm, l, m, xl, aw, fvt int32
		var gp, kg, em, fk, dge interface{}
		var bdat, kl6, vam, zst, h_bez_r sql.NullString
		var hnr_r, h_ide, h_akt interface{}
		if err := rows.Scan(
			&id, &idh, &lw, &hnr, &bdat, &gp, &kg,
			&ka, &v, &em, &s, &k, &vo, &b,
			&tb, &idet, &idgt, &fk, &sn, &kl6,
			&vam, &sm, &l, &m, &xl, &zst, &dge, &aw, &i.Vermittelt, &fvt,
			&hnr_r, &h_bez_r, &h_ide, &h_akt,
		); err != nil {
			log.Printf("[DB] ListBuchungen Scan Error: %v", err)
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
		i.Futterverbrauchtier = int64(fvt)
		i.HerdenNummerRel = sql.NullInt64{Int64: toInt64(hnr_r), Valid: hnr_r != nil}
		i.HerdenBezeichnungRel = h_bez_r
		i.HerdenIDEilager = sql.NullInt64{Int64: toInt64(h_ide), Valid: h_ide != nil}
		i.HerdenAktivRel = sql.NullInt64{Int64: toInt64(h_akt), Valid: h_akt != nil}
		items = append(items, i)
	}
	return items, nil
}

func (w *PostgresWrapper) ListBuchungenWithHerde(ctx context.Context) ([]repo.ListBuchungenWithHerdeRow, error) {
	log.Printf("[DB] ListBuchungenWithHerde called")
	res, err := w.pg.ListBuchungenWithHerde(ctx)
	if err != nil {
		log.Printf("[DB] ListBuchungenWithHerde Query Error: %v", err)
		return nil, err
	}
	items := make([]repo.ListBuchungenWithHerdeRow, len(res))
	for i, v := range res {
		items[i] = repo.ListBuchungenWithHerdeRow{
			ID:                int64(v.ID),
			IDHerden:          int64(v.IDHerden),
			Herdennummer:      int64(v.Herdennummer),
			Buchungsdatum:     v.Buchungsdatum,
			Tierbestand:       int64(v.Tierbestand),
			HerdenBezeichnung: v.HerdenBezeichnung,
		}
	}
	return items, nil
}

func (w *PostgresWrapper) GetBuchung(ctx context.Context, id int64) (repo.Buchung, error) {
	query := `
		SELECT B.ID, B.ID_HERDEN, B.LW, B.HERDENNUMMER, B.BUCHUNGSDATUM, B.GEWICHTPROBE, B.KONTROLLGEWICHT, 
		       B.KLASSEA, B.VERLUSTE, B.EIMASSE, B.SCHMUTZ, B.KNICKEIER, B.VOLLEI, B.BRUCHEIER, 
		       B.TIERBESTAND, B.ID_EITABELLE, B.ID_DGEWICHTTAB, B.FUTTERKTAG, B.SILONR, B.KL6, 
		       B.VERMITTELTAM, B.SMALL, B.LARGE, B.MEDIUM, B.XL, B.ZEITSTEMPEL, B.DGEWICHTEI, B.AW, B.VERMITTELT, B.FUTTERVERBRAUCHTIER
		FROM BUCHUNG B
		WHERE B.ID = $1
	`
	row := w.queryRow(ctx, query, id)
	var idVal, idh, lw, hnr, ka, v, s, k, vo, b, tb, idet, idgt, sn, sm, l, m, xl, aw, fvt int32
	var gp, kg, em, fk, dge interface{}
	var bdat, kl6, vam, zst sql.NullString
	var vermittelt string

	err := row.Scan(
		&idVal, &idh, &lw, &hnr, &bdat, &gp, &kg,
		&ka, &v, &em, &s, &k, &vo, &b,
		&tb, &idet, &idgt, &fk, &sn, &kl6,
		&vam, &sm, &l, &m, &xl, &zst, &dge, &aw, &vermittelt, &fvt,
	)
	if err != nil {
		log.Printf("[DB] GetBuchung Scan Error for ID %d: %v", id, err)
		return repo.Buchung{}, err
	}

	return repo.Buchung{
		ID:              int64(idVal),
		IDHerden:        int64(idh),
		Lw:              int64(lw),
		Herdennummer:    int64(hnr),
		Buchungsdatum:   bdat.String,
		Gewichtprobe:    int64(toFloat(gp)),
		Kontrollgewicht: toFloat(kg),
		Klassea:         int64(ka),
		Verluste:        int64(v),
		Eimasse:         toFloat(em),
		Schmutz:         int64(s),
		Knickeier:       int64(k),
		Vollei:          toFloat(vo),
		Brucheier:       int64(b),
		Tierbestand:     int64(tb),
		IDEitabelle:     int64(idet),
		IDDgewichttab:   int64(idgt),
		Futterktag:      int64(toFloat(fk)),
		Silonr:          int64(sn),
		Kl6:             int64(toInt64(kl6.String)),
		Vermitteltam:    vam.String,
		Small:           int64(sm),
		Large:           int64(l),
		Medium:          int64(m),
		Xl:              int64(xl),
		Zeitstempel:     zst.String,
		Dgewichtei:      toFloat(dge),
		Aw:              int64(aw),
		Vermittelt:      vermittelt,
		Futterverbrauchtier: int64(fvt),
	}, nil
}


func (w *PostgresWrapper) CreateBuchung(ctx context.Context, arg repo.CreateBuchungParams) (repo.Buchung, error) {
	res, err := w.pg.CreateBuchung(ctx, repo_postgres.CreateBuchungParams{
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
		Futterverbrauchtier: int32(arg.Futterverbrauchtier),
	})
	if err != nil {
		return repo.Buchung{}, err
	}
	id := int64(res.ID)
	return w.GetBuchung(ctx, id)
}

func (w *PostgresWrapper) UpdateBuchung(ctx context.Context, arg repo.UpdateBuchungParams) (repo.Buchung, error) {
	_, err := w.pg.UpdateBuchung(ctx, repo_postgres.UpdateBuchungParams{
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
		Futterverbrauchtier: int32(arg.Futterverbrauchtier),
		ID:              int32(arg.ID),
	})
	if err != nil {
		return repo.Buchung{}, err
	}
	return w.GetBuchung(ctx, arg.ID)
}

// --- Herde Methoden ---

func (w *PostgresWrapper) ListHerden(ctx context.Context) ([]repo.ListHerdenRow, error) {
	res, err := w.pg.ListHerden(ctx)
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
			Aktiv:                 int64(v.Aktiv),
			Aw:                    int64(v.Aw),
			Allebuchungenmitdatum: int64(v.Allebuchungenmitdatum),
			StallBezeichnung:      v.StallBezeichnung,
		}
	}
	return items, nil
}

func (w *PostgresWrapper) GetHerde(ctx context.Context, id int64) (repo.Herden, error) {
	v, err := w.pg.GetHerde(ctx, int32(id))
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

func (w *PostgresWrapper) CreateHerde(ctx context.Context, arg repo.CreateHerdeParams) (repo.Herden, error) {
	res, err := w.pg.CreateHerde(ctx, repo_postgres.CreateHerdeParams{
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
	id := int64(res.ID)
	return w.GetHerde(ctx, id)
}

func (w *PostgresWrapper) UpdateHerde(ctx context.Context, arg repo.UpdateHerdeParams) (repo.Herden, error) {
	_, err := w.pg.UpdateHerde(ctx, repo_postgres.UpdateHerdeParams{
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

func (w *PostgresWrapper) GetEggBookingYears(ctx context.Context, arg repo.GetEggBookingYearsParams) ([]interface{}, error) {
	query := `
		SELECT DISTINCT SUBSTRING(BUCHUNGSDATUM FROM 1 FOR 4) as year
		FROM BUCHUNG
		WHERE ($1 = -1 OR ID_HERDEN = $1)
		  AND ($2 = 0 OR EXISTS (SELECT 1 FROM HERDEN H WHERE H.ID = BUCHUNG.ID_HERDEN AND H.AKTIV = 1))
		ORDER BY year DESC
	`
	idHerden := toInt64(arg.IDHerden)
	onlyActive := toInt64(arg.OnlyActive)

	rows, err := w.query(ctx, query, idHerden, onlyActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []interface{}
	for rows.Next() {
		var year sql.NullString
		if err := rows.Scan(&year); err != nil {
			return nil, err
		}
		if year.Valid && year.String != "" {
			items = append(items, year.String)
		}
	}
	return items, nil
}

func (w *PostgresWrapper) GetEggStatsByHerde(ctx context.Context, arg repo.GetEggStatsByHerdeParams) (repo.GetEggStatsByHerdeRow, error) {
	query := `
		SELECT COALESCE(SUM(KLASSEA), 0)                                                               AS SUM_KLASSE_A,
		       COALESCE(SUM(SMALL), 0)                                                                 AS SUM_SMALL,
		       COALESCE(SUM(MEDIUM), 0)                                                                AS SUM_MEDIUM,
		       COALESCE(SUM(LARGE), 0)                                                                 AS SUM_LARGE,
		       COALESCE(SUM(XL), 0)                                                                    AS SUM_XL,
		       COALESCE(SUM(VERLUSTE), 0) +
		       COALESCE((SELECT SUM(T.BEWEGUNGEN)
		                 FROM TIERBEWEGUNGEN T
		                          JOIN HERDEN H ON T.HERDENNUMMER = H.HERDENNUMMER
		                 WHERE H.ID = $1), 0) AS SUM_VERLUSTE
		FROM BUCHUNG
		WHERE BUCHUNG.ID_HERDEN = $2
		  AND VERMITTELT IN ('N', 'V')
	`
	row := w.queryRow(ctx, query, arg.ID, arg.IDHerden)

	var res repo.GetEggStatsByHerdeRow
	var sumA, sumS, sumM, sumL, sumXL, sumV int64
	if err := row.Scan(&sumA, &sumS, &sumM, &sumL, &sumXL, &sumV); err != nil {
		return res, err
	}
	res.SumKlasseA = sumA
	res.SumSmall = sumS
	res.SumMedium = sumM
	res.SumLarge = sumL
	res.SumXl = sumXL
	res.SumVerluste = sumV
	return res, nil
}

func (w *PostgresWrapper) GetEggStatsByHerdeFiltered(ctx context.Context, arg repo.GetEggStatsByHerdeFilteredParams) (repo.GetEggStatsByHerdeFilteredRow, error) {
	query := `
		SELECT COALESCE(SUM(KLASSEA), 0)  AS SUM_KLASSE_A,
		       COALESCE(SUM(SMALL), 0)    AS SUM_SMALL,
		       COALESCE(SUM(MEDIUM), 0)   AS SUM_MEDIUM,
		       COALESCE(SUM(LARGE), 0)    AS SUM_LARGE,
		       COALESCE(SUM(XL), 0)       AS SUM_XL,
		       COALESCE(SUM(VERLUSTE), 0) AS SUM_VERLUSTE
		FROM BUCHUNG
		WHERE ($1 = -1 OR BUCHUNG.ID_HERDEN = $1)
		  AND ($2 = 0 OR EXISTS (SELECT 1 FROM HERDEN H WHERE H.ID = BUCHUNG.ID_HERDEN AND H.AKTIV = 1))
		  AND ($3 = '' OR SUBSTRING(BUCHUNGSDATUM FROM 1 FOR 4) = $3)
		  AND ($4 = 0 OR (CAST(SUBSTRING(BUCHUNGSDATUM FROM 6 FOR 2) AS INTEGER) + 2) / 3 = $4)
		  AND ($5 = 0 OR CAST(SUBSTRING(BUCHUNGSDATUM FROM 6 FOR 2) AS INTEGER) = $5)
		  AND VERMITTELT IN ('N', 'V')
	`
	idHerden := toInt64(arg.IDHerden)
	onlyActive := toInt64(arg.OnlyActive)
	yearStr := toString(arg.Year)
	quarterInt := toInt64(arg.Quarter)
	monthInt := toInt64(arg.Month)

	row := w.queryRow(ctx, query, idHerden, onlyActive, yearStr, quarterInt, monthInt)

	var res repo.GetEggStatsByHerdeFilteredRow
	var sumA, sumS, sumM, sumL, sumXL, sumV int64
	if err := row.Scan(&sumA, &sumS, &sumM, &sumL, &sumXL, &sumV); err != nil {
		return res, err
	}
	res.SumKlasseA = sumA
	res.SumSmall = sumS
	res.SumMedium = sumM
	res.SumLarge = sumL
	res.SumXl = sumXL
	res.SumVerluste = sumV
	return res, nil
}

func (w *PostgresWrapper) GetEggStatsWeeklyByHerde(ctx context.Context, idHerden int64) ([]repo.GetEggStatsWeeklyByHerdeRow, error) {
	query := `
		SELECT LW                         AS LEBENSWOCHE,
		       MAX(BUCHUNGSDATUM)         AS LETZTES_DATUM,
		       COALESCE(SUM(KLASSEA), 0)  AS SUM_KLASSE_A,
		       COALESCE(SUM(SMALL), 0)    AS SUM_SMALL,
		       COALESCE(SUM(MEDIUM), 0)   AS SUM_MEDIUM,
		       COALESCE(SUM(LARGE), 0)    AS SUM_LARGE,
		       COALESCE(SUM(XL), 0)       AS SUM_XL,
		       COALESCE(SUM(VERLUSTE), 0) AS SUM_VERLUSTE
		FROM BUCHUNG
		WHERE ID_HERDEN = $1
		  AND VERMITTELT IN ('N', 'V')
		GROUP BY LW
		ORDER BY LEBENSWOCHE DESC
	`
	rows, err := w.query(ctx, query, idHerden)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []repo.GetEggStatsWeeklyByHerdeRow
	for rows.Next() {
		var i repo.GetEggStatsWeeklyByHerdeRow
		var lw, sumA, sumS, sumM, sumL, sumXL, sumV int64
		var datum sql.NullString
		if err := rows.Scan(&lw, &datum, &sumA, &sumS, &sumM, &sumL, &sumXL, &sumV); err != nil {
			return nil, err
		}
		i.Lebenswoche = lw
		i.LetztesDatum = datum.String
		i.SumKlasseA = sumA
		i.SumSmall = sumS
		i.SumMedium = sumM
		i.SumLarge = sumL
		i.SumXl = sumXL
		i.SumVerluste = sumV
		items = append(items, i)
	}
	return items, nil
}


// --- Eilager Methoden ---

func (w *PostgresWrapper) ListEilager(ctx context.Context) ([]repo.ListEilagerRow, error) {
	log.Printf("[DB] ListEilager called (Postgres)")
	res, err := w.pg.ListEilager(ctx)
	if err != nil {
		log.Printf("[DB] ListEilager Query Error: %v", err)
		return nil, err
	}
	items := make([]repo.ListEilagerRow, len(res))
	for idx, v := range res {
		items[idx] = repo.ListEilagerRow{
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
			Aw:            int64(v.Aw),
			Klasse6:       v.Klasse6,
			Klasse7:       v.Klasse7,
		}
	}
	return items, nil
}

func (w *PostgresWrapper) GetEilager(ctx context.Context, id int64) (repo.Eilager, error) {
	v, err := w.pg.GetEilager(ctx, int32(id))
	if err != nil {
		return repo.Eilager{}, err
	}
	return convertPgEilager(v), nil
}

func (w *PostgresWrapper) CreateEilager(ctx context.Context, arg repo.CreateEilagerParams) (repo.Eilager, error) {
	res, err := w.pg.CreateEilager(ctx, repo_postgres.CreateEilagerParams{
		Lagernummer:   int32(arg.Lagernummer),
		Kz:            toString(arg.Kz),
		Bezeichnung:   arg.Bezeichnung,
		LetzteBuchung: arg.LetzteBuchung,
	})
	if err != nil {
		return repo.Eilager{}, err
	}
	id := int64(res.ID)
	return w.GetEilager(ctx, id)
}

func (w *PostgresWrapper) UpdateEilager(ctx context.Context, arg repo.UpdateEilagerParams) (repo.Eilager, error) {
	_, err := w.pg.UpdateEilager(ctx, repo_postgres.UpdateEilagerParams{
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

func (w *PostgresWrapper) DeleteEilager(ctx context.Context, id int64) error {
	return w.pg.DeleteEilager(ctx, int32(id))
}

func (w *PostgresWrapper) GetBestandsuebersicht(ctx context.Context, arg repo.GetBestandsuebersichtParams) ([]repo.GetBestandsuebersichtRow, error) {
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
		         FROM EILAGERBUCHUNG EB
		                  LEFT JOIN EILAGER E ON EB.ID_EILAGER = E.ID
		                  LEFT JOIN LAGERPLATZ LP ON EB.ID_LAGERPLATZ = LP.ID
		         WHERE ($1 = 0 OR EB.ID_EILAGER = $2)
		
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
		         FROM EILAGERBUCHUNG EB
		                  LEFT JOIN EILAGER E ON EB.ID_FREMDESLAGER = E.ID
		         WHERE EB.ID_FREMDESLAGER != 0 AND ($3 = 0 OR EB.ID_FREMDESLAGER = $4)
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

func (w *PostgresWrapper) ListEilagerBuchungenByKZ(ctx context.Context, kz interface{}) ([]repo.Eilagerbuchung, error) {
	kzStr := toString(kz)
	log.Printf("[DB] ListEilagerBuchungenByKZ called for KZ: %s", kzStr)
	query := `SELECT id, id_fremdeslager, id_buchung, id_eilager, buchungsdatum, jumbos, xl, large, medium, small, volleikg, schmutz, knickeier, brucheier, buchungstyp, charge, kz_lager, id_fremdebuchung, verkauf FROM EILAGERBUCHUNG WHERE kz_lager = $1 ORDER BY buchungsdatum DESC`
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

func (w *PostgresWrapper) ListEilagerBuchungenByLager(ctx context.Context, idEilager int64) ([]repo.Eilagerbuchung, error) {
	log.Printf("[DB] ListEilagerBuchungenByLager called for ID: %d", idEilager)
	query := `SELECT id, id_fremdeslager, id_buchung, id_eilager, buchungsdatum, jumbos, xl, large, medium, small, volleikg, schmutz, knickeier, brucheier, buchungstyp, charge, kz_lager, id_fremdebuchung, verkauf FROM EILAGERBUCHUNG WHERE id_eilager = $1 ORDER BY buchungsdatum DESC`
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

func (w *PostgresWrapper) GetEilagerSumByBuchungID(ctx context.Context, idBuchung int64) (repo.GetEilagerSumByBuchungIDRow, error) {
	v, err := w.pg.GetEilagerSumByBuchungID(ctx, int32(idBuchung))
	if err != nil {
		return repo.GetEilagerSumByBuchungIDRow{}, err
	}
	return repo.GetEilagerSumByBuchungIDRow{
		Jumbos:    toInt64(v.Jumbos),
		Xl:        toInt64(v.Xl),
		Large:     toInt64(v.Large),
		Medium:    toInt64(v.Medium),
		Small:     toInt64(v.Small),
		Volleikg:  toFloat(v.Volleikg),
		Schmutz:   toInt64(v.Schmutz),
		Knickeier: toInt64(v.Knickeier),
		Brucheier: toInt64(v.Brucheier),
	}, nil
}

func (w *PostgresWrapper) GetEilagerSumBySource(ctx context.Context, arg repo.GetEilagerSumBySourceParams) (repo.GetEilagerSumBySourceRow, error) {
	v, err := w.pg.GetEilagerSumBySource(ctx, repo_postgres.GetEilagerSumBySourceParams{
		IDBuchung:      toInt32(arg.IDBuchung),
		IDFremdeslager: toInt32(arg.IDFremdeslager),
	})
	if err != nil {
		return repo.GetEilagerSumBySourceRow{}, err
	}
	return repo.GetEilagerSumBySourceRow{
		Jumbos:   toInt64(v.Jumbos),
		Xl:       toInt64(v.Xl),
		Large:    toInt64(v.Large),
		Medium:   toInt64(v.Medium),
		Small:    toInt64(v.Small),
		Volleikg: toFloat(v.Volleikg),
	}, nil
}

func convertPgEilager(v repo_postgres.Eilager) repo.Eilager {
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

func convertPgEilagerbuchung(v repo_postgres.Eilagerbuchung) repo.Eilagerbuchung {
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

func (w *PostgresWrapper) GetEierpreis(ctx context.Context, id int64) (repo.Eierpreise, error) {
	v, err := w.pg.GetEierpreis(ctx, int32(id))
	if err != nil {
		return repo.Eierpreise{}, err
	}
	return convertPgEierpreise(v), nil
}

func convertPgEierpreise(v repo_postgres.Eierpreise) repo.Eierpreise {
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

func (w *PostgresWrapper) ListSilos(ctx context.Context) ([]repo.Silo, error) {
	res, err := w.pg.ListSilos(ctx)
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

func (w *PostgresWrapper) CreateSilo(ctx context.Context, arg repo.CreateSiloParams) (repo.Silo, error) {
	res, err := w.pg.CreateSilo(ctx, repo_postgres.CreateSiloParams{
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
	id := int64(res.ID)
	return w.GetSilo(ctx, id)
}

func (w *PostgresWrapper) GetSilo(ctx context.Context, id int64) (repo.Silo, error) {
	v, err := w.pg.GetSilo(ctx, int32(id))
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

func (w *PostgresWrapper) GetGlobalFirmenparameter(ctx context.Context) (repo.Firmenparameter, error) {
	v, err := w.pg.GetGlobalFirmenparameter(ctx)
	if err != nil {
		return repo.Firmenparameter{}, err
	}
	return convertPgFirmenparameter(v), nil
}

func (w *PostgresWrapper) UpdateFirmenparameter(ctx context.Context, arg repo.UpdateFirmenparameterParams) (repo.Firmenparameter, error) {
	_, err := w.pg.UpdateFirmenparameter(ctx, repo_postgres.UpdateFirmenparameterParams{
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
		Futterinventur:            int32(arg.Futterinventur),
	})
	if err != nil {
		return repo.Firmenparameter{}, err
	}
	return w.GetFirmenparameterByHerde(ctx, arg.IDHerden)
}

func (w *PostgresWrapper) GetFirmenparameterByHerde(ctx context.Context, idHerden int64) (repo.Firmenparameter, error) {
	v, err := w.pg.GetFirmenparameterByHerde(ctx, int32(idHerden))
	if err != nil {
		return repo.Firmenparameter{}, err
	}
	return convertPgFirmenparameter(v), nil
}

func convertPgFirmenparameter(v repo_postgres.Firmenparameter) repo.Firmenparameter {
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
		Futterinventur:            int64(v.Futterinventur),
	}
}

// --- Person Methoden ---

func (w *PostgresWrapper) ListPersonen(ctx context.Context) ([]repo.Person, error) {
	res, err := w.pg.ListPersonen(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.Person, len(res))
	for i, v := range res {
		items[i] = convertPgPerson(v)
	}
	return items, nil
}

func (w *PostgresWrapper) GetPerson(ctx context.Context, id int64) (repo.Person, error) {
	v, err := w.pg.GetPerson(ctx, int32(id))
	if err != nil {
		return repo.Person{}, err
	}
	return convertPgPerson(v), nil
}

func (w *PostgresWrapper) CreatePerson(ctx context.Context, arg repo.CreatePersonParams) (repo.Person, error) {
	res, err := w.pg.CreatePerson(ctx, repo_postgres.CreatePersonParams{
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
		Foto:           arg.Foto,
		Homepage:       arg.Homepage,
	})
	if err != nil {
		return repo.Person{}, err
	}
	id := int64(res.ID)
	return w.GetPerson(ctx, id)
}

func (w *PostgresWrapper) ListTabellenkopfByType(ctx context.Context, tabellentyp interface{}) ([]repo.Tabellenkopf, error) {
	log.Printf("[MARIADB-FIX] ListTabellenkopfByType called: %v", tabellentyp)
	rows, err := w.db.QueryContext(ctx, "SELECT id, tabellentyp, tabellennummer, bezeichnung, anlagedatum, datum FROM TABELLENKOPF WHERE tabellentyp = $1", tabellentyp)
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

func (w *PostgresWrapper) GetGewichtByTabNumAndWeight(ctx context.Context, arg repo.GetGewichtByTabNumAndWeightParams) (repo.Gewichttabelle, error) {
	v, err := w.pg.GetGewichtByTabNumAndWeight(ctx, repo_postgres.GetGewichtByTabNumAndWeightParams{
		Tabellennummer: int32(arg.Tabellennummer),
		Round:          arg.ROUND,
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

func (w *PostgresWrapper) ListGewichtByTabNum(ctx context.Context, tabNum int64) ([]repo.Gewichttabelle, error) {
	res, err := w.pg.ListGewichtByTabNum(ctx, int32(tabNum))
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

func (w *PostgresWrapper) ListEierpreise(ctx context.Context) ([]repo.Eierpreise, error) {
	res, err := w.pg.ListEierpreise(ctx)
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

func convertPgPerson(v repo_postgres.Person) repo.Person {
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
		Foto:           v.Foto,
		Homepage:       v.Homepage,
	}
}

// --- Tierbewegungen Methoden ---

func (w *PostgresWrapper) ListTierbewegungen(ctx context.Context, spracheKz string) ([]repo.ListTierbewegungenRow, error) {
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
		FROM TIERBEWEGUNGEN t
				 LEFT JOIN TEXTE txt on t.id_texte = txt.id
				 LEFT JOIN UEBERSETZUNGEN u on txt.id = u.id_texte and u.sprache_kz = $1
				 LEFT JOIN HERDEN h on t.herdennummer = h.herdennummer
				 LEFT JOIN HERDEN hv on t.id_herden_von = hv.id
				 LEFT JOIN HERDEN hn on t.id_herden_nach = hn.id
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

func (w *PostgresWrapper) GetTierbewegung(ctx context.Context, id int64) (repo.Tierbewegungen, error) {
	query := `SELECT id, herdennummer, id_buchung, typ, id_texte, bewegungsdatum, bewegungen, id_herden_von, id_herden_nach, kosten FROM TIERBEWEGUNGEN WHERE id = $1`
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

func (w *PostgresWrapper) CreateTierbewegung(ctx context.Context, arg repo.CreateTierbewegungParams) (repo.Tierbewegungen, error) {
	res, err := w.pg.CreateTierbewegung(ctx, repo_postgres.CreateTierbewegungParams{
		Herdennummer:   toNullInt32(arg.Herdennummer),
		IDBuchung:      toNullInt32(arg.IDBuchung),
		Typ:            toNullString(arg.Typ),
		IDTexte:        toNullInt32(arg.IDTexte),
		Bewegungsdatum: toNullString(arg.Bewegungsdatum),
		Bewegungen:     toNullInt32(arg.Bewegungen),
		IDHerdenVon:    toNullInt32(arg.IDHerdenVon),
		IDHerdenNach:   toNullInt32(arg.IDHerdenNach),
		Kosten:         toNullFloat64(arg.Kosten),
	})
	if err != nil {
		return repo.Tierbewegungen{}, err
	}
	id := int64(res.ID)
	return w.GetTierbewegung(ctx, id)
}

func (w *PostgresWrapper) UpdateTierbewegung(ctx context.Context, arg repo.UpdateTierbewegungParams) (repo.Tierbewegungen, error) {
	_, err := w.pg.UpdateTierbewegung(ctx, repo_postgres.UpdateTierbewegungParams{
		Herdennummer:   toNullInt32(arg.Herdennummer),
		IDBuchung:      toNullInt32(arg.IDBuchung),
		Typ:            toNullString(arg.Typ),
		IDTexte:        toNullInt32(arg.IDTexte),
		Bewegungsdatum: toNullString(arg.Bewegungsdatum),
		Bewegungen:     toNullInt32(arg.Bewegungen),
		IDHerdenVon:    toNullInt32(arg.IDHerdenVon),
		IDHerdenNach:   toNullInt32(arg.IDHerdenNach),
		Kosten:         toNullFloat64(arg.Kosten),
		ID:             int32(arg.ID),
	})
	if err != nil {
		return repo.Tierbewegungen{}, err
	}
	return w.GetTierbewegung(ctx, arg.ID)
}

func (w *PostgresWrapper) DeleteTierbewegung(ctx context.Context, id int64) error {
	return w.pg.DeleteTierbewegung(ctx, int32(id))
}

// --- Lookup & Liste Methoden ---

func (w *PostgresWrapper) ListHerdenLookup(ctx context.Context) ([]repo.ListHerdenLookupRow, error) {
	res, err := w.pg.ListHerdenLookup(ctx)
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

func (w *PostgresWrapper) ListRassen(ctx context.Context) ([]repo.Rasse, error) {
	res, err := w.pg.ListRassen(ctx)
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

func (w *PostgresWrapper) ListStaelle(ctx context.Context) ([]repo.Stall, error) {
	res, err := w.pg.ListStaelle(ctx)
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

func (w *PostgresWrapper) ListLieferanten(ctx context.Context) ([]repo.Person, error) {
	res, err := w.pg.ListLieferanten(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.Person, len(res))
	for i, v := range res {
		items[i] = convertPgPerson(v)
	}
	return items, nil
}

func (w *PostgresWrapper) ListZuechter(ctx context.Context) ([]repo.Person, error) {
	res, err := w.pg.ListZuechter(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]repo.Person, len(res))
	for i, v := range res {
		items[i] = convertPgPerson(v)
	}
	return items, nil
}

// --- Texte & Übersetzungen ---

// (Duplicate removed)

func (w *PostgresWrapper) GetTranslatedText(ctx context.Context, arg repo.GetTranslatedTextParams) (repo.GetTranslatedTextRow, error) {
	v, err := w.pg.GetTranslatedText(ctx, repo_postgres.GetTranslatedTextParams{
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

func (w *PostgresWrapper) ListFuttersorten(ctx context.Context) ([]repo.Futtersorten, error) {
	res, err := w.pg.ListFuttersorten(ctx)
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

func (w *PostgresWrapper) ListLagerplaetze(ctx context.Context) ([]repo.ListLagerplaetzeRow, error) {
	res, err := w.pg.ListLagerplaetze(ctx)
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

func (w *PostgresWrapper) ListLagerplaetzeByEilager(ctx context.Context, idEilager int64) ([]repo.ListLagerplaetzeByEilagerRow, error) {
	res, err := w.pg.ListLagerplaetzeByEilager(ctx, int32(idEilager))
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

func (w *PostgresWrapper) UpdateEierpreis(ctx context.Context, arg repo.UpdateEierpreisParams) (repo.Eierpreise, error) {
	_, err := w.pg.UpdateEierpreis(ctx, repo_postgres.UpdateEierpreisParams{
		Eierklasse: arg.Eierklasse,
		PreisVon:   toFloat(arg.PreisVon),
		ID:         int32(arg.ID),
	})
	if err != nil {
		return repo.Eierpreise{}, err
	}
	return w.GetEierpreis(ctx, arg.ID)
}

func (w *PostgresWrapper) ListVerkauf(ctx context.Context) ([]repo.Verkauf, error) {
	log.Printf("[DB] ListVerkauf called")
	query := `SELECT id, id_eilagerbuchung, id_buchung, buchungsdatum, mengesmall, mengemedium, mengelarge, mengexl, preissmall, preismedium, preislarge, preisxl, gesamtpreis, bio, verbucht, charge, rabattprozent FROM VERKAUF ORDER BY buchungsdatum DESC`
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

func (w *PostgresWrapper) GetVerkauf(ctx context.Context, id int64) (repo.Verkauf, error) {
	v, err := w.pg.GetVerkauf(ctx, int32(id))
	if err != nil {
		return repo.Verkauf{}, err
	}
	return convertPgVerkauf(v), nil
}

func (w *PostgresWrapper) CreateVerkauf(ctx context.Context, arg repo.CreateVerkaufParams) (repo.Verkauf, error) {
	res, err := w.pg.CreateVerkauf(ctx, repo_postgres.CreateVerkaufParams{
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
		Bio:              toInt16(arg.Bio),
		Verbucht:         toInt16(arg.Verbucht),
		Charge:           toString(arg.Charge),
		Rabattprozent:    toFloat(arg.Rabattprozent),
	})
	if err != nil {
		return repo.Verkauf{}, err
	}
	id := int64(res.ID)
	return w.GetVerkauf(ctx, id)
}

func (w *PostgresWrapper) UpdateVerkauf(ctx context.Context, arg repo.UpdateVerkaufParams) (repo.Verkauf, error) {
	_, err := w.pg.UpdateVerkauf(ctx, repo_postgres.UpdateVerkaufParams{
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
		Bio:              toInt16(arg.Bio),
		Verbucht:         toInt16(arg.Verbucht),
		Charge:           toString(arg.Charge),
		Rabattprozent:    toFloat(arg.Rabattprozent),
		ID:               int32(arg.ID),
	})
	if err != nil {
		return repo.Verkauf{}, err
	}
	return w.GetVerkauf(ctx, arg.ID)
}

func convertPgVerkauf(v repo_postgres.Verkauf) repo.Verkauf {
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
		Bio:              v.Bio != 0,
		Verbucht:         v.Verbucht != 0,
		Charge:           v.Charge,
		Rabattprozent:    v.Rabattprozent,
	}
}
func (w *PostgresWrapper) ListDynamischeSQL(ctx context.Context) ([]repo.ListDynamischeSQLRow, error) {
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

func (w *PostgresWrapper) GetDynamischeSQL(ctx context.Context, id int64) (repo.DynamischeSql, error) {
	log.Printf("[DB] GetDynamischeSQL called for ID: %d", id)
	query := `SELECT id, beschreibung, sqlstatement, kategorie_kz, gruppen_kz, typ_kz, template_name, param_def, detail_sql, link_logic, group_field, rows_per_page, page_orientation, show_master_grid, show_detail_grid, system_kz, sqlstatement_native, detail_sql_native, root_kz, summenzeile, ist_summenzeile FROM DYNAMISCHE_SQL WHERE id = $1`
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

func (w *PostgresWrapper) CreateDynamischeSQL(ctx context.Context, arg repo.CreateDynamischeSQLParams) (repo.DynamischeSql, error) {
	log.Printf("[DB] CreateDynamischeSQL called for Beschreibung: %s", arg.Beschreibung)
	res, err := w.pg.CreateDynamischeSQL(ctx, repo_postgres.CreateDynamischeSQLParams{
		Beschreibung:       arg.Beschreibung,
		Sqlstatement:       arg.Sqlstatement,
		KategorieKz:        toString(arg.KategorieKz),
		GruppenKz:          toString(arg.GruppenKz),
		TypKz:              toString(arg.TypKz),
		TemplateName:       arg.TemplateName,
		ParamDef:           arg.ParamDef,
		DetailSql:          arg.DetailSql,
		LinkLogic:          arg.LinkLogic,
		GroupField:         arg.GroupField,
		RowsPerPage:        int32(arg.RowsPerPage),
		PageOrientation:    arg.PageOrientation,
		ShowMasterGrid:     int32(arg.ShowMasterGrid),
		ShowDetailGrid:     int32(arg.ShowDetailGrid),
		SystemKz:           arg.SystemKz,
		SqlstatementNative: arg.SqlstatementNative,
		DetailSqlNative:    arg.DetailSqlNative,
		RootKz:             toString(arg.RootKz),
		Summenzeile:        arg.Summenzeile,
		IstSummenzeile:     int32(arg.IstSummenzeile),
	})
	if err != nil {
		log.Printf("[DB] CreateDynamischeSQL Error: %v", err)
		return repo.DynamischeSql{}, err
	}
	id := int64(res.ID)
	return w.GetDynamischeSQL(ctx, id)
}

func (w *PostgresWrapper) UpdateDynamischeSQL(ctx context.Context, arg repo.UpdateDynamischeSQLParams) (repo.DynamischeSql, error) {
	log.Printf("[DB] UpdateDynamischeSQL called for ID: %d", arg.ID)
	_, err := w.pg.UpdateDynamischeSQL(ctx, repo_postgres.UpdateDynamischeSQLParams{
		Beschreibung:       arg.Beschreibung,
		Sqlstatement:       arg.Sqlstatement,
		KategorieKz:        toString(arg.KategorieKz),
		GruppenKz:          toString(arg.GruppenKz),
		TypKz:              toString(arg.TypKz),
		TemplateName:       arg.TemplateName,
		ParamDef:           arg.ParamDef,
		DetailSql:          arg.DetailSql,
		LinkLogic:          arg.LinkLogic,
		GroupField:         arg.GroupField,
		RowsPerPage:        int32(arg.RowsPerPage),
		PageOrientation:    arg.PageOrientation,
		ShowMasterGrid:     int32(arg.ShowMasterGrid),
		ShowDetailGrid:     int32(arg.ShowDetailGrid),
		SystemKz:           arg.SystemKz,
		SqlstatementNative: arg.SqlstatementNative,
		DetailSqlNative:    arg.DetailSqlNative,
		RootKz:             toString(arg.RootKz),
		Summenzeile:        arg.Summenzeile,
		IstSummenzeile:     int32(arg.IstSummenzeile),
		ID:                 int32(arg.ID),
	})
	if err != nil {
		log.Printf("[DB] UpdateDynamischeSQL Error: %v", err)
		return repo.DynamischeSql{}, err
	}
	return w.GetDynamischeSQL(ctx, arg.ID)
}

func (w *PostgresWrapper) CreateTabellenkopf(ctx context.Context, arg repo.CreateTabellenkopfParams) (repo.Tabellenkopf, error) {
	log.Printf("[DB] CreateTabellenkopf called for Typ: %s, Nr: %d", arg.Tabellentyp, arg.Tabellennummer)
	query := `INSERT INTO TABELLENKOPF (TABELLENTYP, TABELLENNUMMER, BEZEICHNUNG, ANLAGEDATUM, DATUM) VALUES ($1, $2, $3, $4, $5)`
	var id int64
	err := w.db.QueryRowContext(ctx, query+" RETURNING ID", arg.Tabellentyp, arg.Tabellennummer, arg.Bezeichnung, arg.Anlagedatum, arg.Datum).Scan(&id)
	if err != nil {
		log.Printf("[DB] CreateTabellenkopf Error: %v", err)
		return repo.Tabellenkopf{}, err
	}
	return repo.Tabellenkopf{
		ID:             id,
		Tabellentyp:    toString(arg.Tabellentyp),
		Tabellennummer: toInt64(arg.Tabellennummer),
		Bezeichnung:    arg.Bezeichnung,
		Anlagedatum:    arg.Anlagedatum,
		Datum:          arg.Datum,
	}, nil
}

func (w *PostgresWrapper) UpdateTabellenkopf(ctx context.Context, arg repo.UpdateTabellenkopfParams) (repo.Tabellenkopf, error) {
	log.Printf("[DB] UpdateTabellenkopf called for ID: %d", arg.ID)
	query := `UPDATE TABELLENKOPF SET TABELLENNUMMER = $1, BEZEICHNUNG = $2, ANLAGEDATUM = $3, DATUM = $4 WHERE ID = $5`
	_, err := w.db.ExecContext(ctx, query, arg.Tabellennummer, arg.Bezeichnung, arg.Anlagedatum, arg.Datum, arg.ID)
	if err != nil {
		log.Printf("[DB] UpdateTabellenkopf Error: %v", err)
		return repo.Tabellenkopf{}, err
	}
	// Reload to get the full object including Tabellentyp which isn't in params
	var t repo.Tabellenkopf
	var ttyp interface{}
	err = w.db.QueryRowContext(ctx, "SELECT ID, TABELLENTYP, TABELLENNUMMER, BEZEICHNUNG, ANLAGEDATUM, DATUM FROM TABELLENKOPF WHERE ID = $1", arg.ID).Scan(
		&t.ID, &ttyp, &t.Tabellennummer, &t.Bezeichnung, &t.Anlagedatum, &t.Datum,
	)
	t.Tabellentyp = toString(ttyp)
	return t, err
}

func (w *PostgresWrapper) ListTextTypen(ctx context.Context) ([]repo.TextTypen, error) {
	log.Printf("[DB] ListTextTypen called")
	rows, err := w.db.QueryContext(ctx, "SELECT ID, KZ, BEZEICHNUNG, SYSTEM_KZ, STATUS FROM TEXT_TYPEN")
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

func (w *PostgresWrapper) CreateTextTyp(ctx context.Context, arg repo.CreateTextTypParams) (repo.TextTypen, error) {
	log.Printf("[DB] CreateTextTyp called: %s", arg.Kz)
	query := "INSERT INTO TEXT_TYPEN (KZ, BEZEICHNUNG, SYSTEM_KZ, STATUS) VALUES ($1, $2, $3, $4) RETURNING ID"
	var id int64
	err := w.db.QueryRowContext(ctx, query, arg.Kz, arg.Bezeichnung, arg.SystemKz, arg.Status).Scan(&id)
	if err != nil {
		log.Printf("[DB] CreateTextTyp Error: %v", err)
		return repo.TextTypen{}, err
	}
	return repo.TextTypen{
		ID:          id,
		Kz:          arg.Kz,
		Bezeichnung: arg.Bezeichnung,
		SystemKz:    arg.SystemKz,
		Status:      arg.Status,
	}, nil
}

func (w *PostgresWrapper) UpdateTextTyp(ctx context.Context, arg repo.UpdateTextTypParams) (repo.TextTypen, error) {
	log.Printf("[DB] UpdateTextTyp called: %d", arg.ID)
	query := "UPDATE TEXT_TYPEN SET KZ = $1, BEZEICHNUNG = $2, SYSTEM_KZ = $3, STATUS = $4 WHERE ID = $5"
	_, err := w.db.ExecContext(ctx, query, arg.Kz, arg.Bezeichnung, arg.SystemKz, arg.Status, arg.ID)
	if err != nil {
		log.Printf("[DB] UpdateTextTyp Error: %v", err)
		return repo.TextTypen{}, err
	}
	var t repo.TextTypen
	err = w.db.QueryRowContext(ctx, "SELECT ID, KZ, BEZEICHNUNG, SYSTEM_KZ, STATUS FROM TEXT_TYPEN WHERE ID = $1", arg.ID).Scan(
		&t.ID, &t.Kz, &t.Bezeichnung, &t.SystemKz, &t.Status,
	)
	return t, err
}

func (w *PostgresWrapper) CreateText(ctx context.Context, arg repo.CreateTextParams) (repo.Texte, error) {
	log.Printf("[DB] CreateText called for Typ: %s", arg.TextTypKz)
	query := "INSERT INTO TEXTE (TEXT_TYP_KZ, KZ, SYSTEM_KZ, STATUS) VALUES ($1, $2, $3, $4) RETURNING ID"
	var id int64
	err := w.db.QueryRowContext(ctx, query, arg.TextTypKz, arg.Kz, arg.SystemKz, arg.Status).Scan(&id)
	if err != nil {
		log.Printf("[DB] CreateText Error: %v", err)
		return repo.Texte{}, err
	}
	return repo.Texte{
		ID:        id,
		TextTypKz: arg.TextTypKz,
		Kz:        arg.Kz,
		SystemKz:  arg.SystemKz,
		Status:    arg.Status,
	}, nil
}

func (w *PostgresWrapper) UpdateText(ctx context.Context, arg repo.UpdateTextParams) (repo.Texte, error) {
	log.Printf("[DB] UpdateText called: %d", arg.ID)
	query := "UPDATE TEXTE SET TEXT_TYP_KZ = $1, KZ = $2, SYSTEM_KZ = $3, STATUS = $4 WHERE ID = $5"
	_, err := w.db.ExecContext(ctx, query, arg.TextTypKz, arg.Kz, arg.SystemKz, arg.Status, arg.ID)
	if err != nil {
		log.Printf("[DB] UpdateText Error: %v", err)
		return repo.Texte{}, err
	}
	var t repo.Texte
	var kz interface{}
	err = w.db.QueryRowContext(ctx, "SELECT ID, TEXT_TYP_KZ, KZ, SYSTEM_KZ, STATUS FROM TEXTE WHERE ID = $1", arg.ID).Scan(
		&t.ID, &t.TextTypKz, &kz, &t.SystemKz, &t.Status,
	)
	t.Kz = toString(kz)
	return t, err
}

func (w *PostgresWrapper) CreateUebersetzung(ctx context.Context, arg repo.CreateUebersetzungParams) (repo.Uebersetzungen, error) {
	log.Printf("[DB] CreateUebersetzung called for ID: %d, Lang: %s", arg.IDTexte, arg.SpracheKz)
	query := "INSERT INTO UEBERSETZUNGEN (ID_TEXTE, SPRACHE_KZ, BETREFF, INHALT) VALUES ($1, $2, $3, $4)"
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

func (w *PostgresWrapper) UpsertUebersetzung(ctx context.Context, arg repo.UpsertUebersetzungParams) (repo.Uebersetzungen, error) {
	log.Printf("[DB] UpsertUebersetzung called for ID: %d, Lang: %s", arg.IDTexte, arg.SpracheKz)
	query := `INSERT INTO UEBERSETZUNGEN (ID_TEXTE, SPRACHE_KZ, BETREFF, INHALT) 
              VALUES ($1, $2, $3, $4) 
              ON CONFLICT (ID_TEXTE, SPRACHE_KZ) DO UPDATE SET BETREFF = EXCLUDED.BETREFF, INHALT = EXCLUDED.INHALT`
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

func (w *PostgresWrapper) ListTexte(ctx context.Context, spracheKz string) ([]repo.ListTexteRow, error) {
	log.Printf("[DB] ListTexte called for Lang: %s", spracheKz)
	query := `SELECT T.ID, T.TEXT_TYP_KZ, T.KZ, T.SYSTEM_KZ, T.STATUS, COALESCE(U.BETREFF, '') AS BETREFF, COALESCE(U.INHALT, '') AS INHALT 
              FROM TEXTE T 
              LEFT JOIN UEBERSETZUNGEN U ON T.ID = U.ID_TEXTE AND U.SPRACHE_KZ = $1`
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

func (w *PostgresWrapper) ListTexteByType(ctx context.Context, arg repo.ListTexteByTypeParams) ([]repo.ListTexteByTypeRow, error) {
	log.Printf("[DB] ListTexteByType called for Typ: %s, Lang: %s", arg.TextTypKz, arg.SpracheKz)
	query := `SELECT T.ID, T.TEXT_TYP_KZ, T.KZ, T.SYSTEM_KZ, T.STATUS, COALESCE(U.BETREFF, '') AS BETREFF, COALESCE(U.INHALT, '') AS INHALT 
              FROM TEXTE T 
              LEFT JOIN UEBERSETZUNGEN U ON T.ID = U.ID_TEXTE AND U.SPRACHE_KZ = $1 
              WHERE T.TEXT_TYP_KZ = $2`
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

func (w *PostgresWrapper) ListBenutzerProfile(ctx context.Context) ([]repo.Benutzerprofile, error) {
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

func (w *PostgresWrapper) GetBenutzerProfilByID(ctx context.Context, id int64) (repo.Benutzerprofile, error) {
	log.Printf("[MARIADB-FIX] GetBenutzerProfilByID called: %d", id)
	var i repo.Benutzerprofile
	var pkz interface{}
	err := w.db.QueryRowContext(ctx, "SELECT id, profil_kz, beschreibung, f_dashboard, f_herden_verwalten, f_einrichtungen_verwalten, f_personen_verwalten, f_buchungen_erfassen, f_auswertungen_anzeigen, f_sql_struktur_verwalten, f_benutzer_profile, f_parameter_editieren, f_kosten_verwalten, f_tabellen_anzeigen, f_texte_verwalten, f_system_verwaltung, f_backup_erstellen FROM BENUTZERPROFILE WHERE id = $1", id).Scan(
		&i.ID, &pkz, &i.Beschreibung, &i.FDashboard, &i.FHerdenVerwalten, &i.FEinrichtungenVerwalten, &i.FPersonenVerwalten, &i.FBuchungenErfassen, &i.FAuswertungenAnzeigen, &i.FSqlStrukturVerwalten, &i.FBenutzerProfile, &i.FParameterEditieren, &i.FKostenVerwalten, &i.FTabellenAnzeigen, &i.FTexteVerwalten, &i.FSystemVerwaltung, &i.FBackupErstellen,
	)
	i.ProfilKz = toString(pkz)
	return i, err
}

func (w *PostgresWrapper) GetBenutzerProfilByKZ(ctx context.Context, pkz interface{}) (repo.Benutzerprofile, error) {
	log.Printf("[MARIADB-FIX] GetBenutzerProfilByKZ called: %v", pkz)
	var i repo.Benutzerprofile
	var pkzResult interface{}
	err := w.db.QueryRowContext(ctx, "SELECT id, profil_kz, beschreibung, f_dashboard, f_herden_verwalten, f_einrichtungen_verwalten, f_personen_verwalten, f_buchungen_erfassen, f_auswertungen_anzeigen, f_sql_struktur_verwalten, f_benutzer_profile, f_parameter_editieren, f_kosten_verwalten, f_tabellen_anzeigen, f_texte_verwalten, f_system_verwaltung, f_backup_erstellen FROM BENUTZERPROFILE WHERE profil_kz = $1", pkz).Scan(
		&i.ID, &pkzResult, &i.Beschreibung, &i.FDashboard, &i.FHerdenVerwalten, &i.FEinrichtungenVerwalten, &i.FPersonenVerwalten, &i.FBuchungenErfassen, &i.FAuswertungenAnzeigen, &i.FSqlStrukturVerwalten, &i.FBenutzerProfile, &i.FParameterEditieren, &i.FKostenVerwalten, &i.FTabellenAnzeigen, &i.FTexteVerwalten, &i.FSystemVerwaltung, &i.FBackupErstellen,
	)
	i.ProfilKz = toString(pkzResult)
	return i, err
}

func (w *PostgresWrapper) ListBenutzer(ctx context.Context) ([]repo.ListBenutzerRow, error) {
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

func (w *PostgresWrapper) ListShowTV(ctx context.Context) ([]repo.Showtv, error) {
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

func (w *PostgresWrapper) ListAktionen(ctx context.Context, arg repo.ListAktionenParams) ([]repo.ListAktionenRow, error) {
	res, err := w.pg.ListAktionen(ctx, repo_postgres.ListAktionenParams{
		IDUser:    toInt32(arg.IDUser),
		StartDate: arg.StartDate,
		EndDate:   arg.EndDate,
		Kz:        arg.Kz,
		Status:    toInt32(arg.Status),
	})
	if err != nil {
		return nil, err
	}
	items := make([]repo.ListAktionenRow, len(res))
	for i, v := range res {
		items[i] = convertPgListAktionenRow(v)
	}
	return items, nil
}

func (w *PostgresWrapper) GetAktion(ctx context.Context, id int64) (repo.Aktionen, error) {
	v, err := w.pg.GetAktion(ctx, int32(id))
	if err != nil {
		return repo.Aktionen{}, err
	}
	return convertPgAktion(v), nil
}

func (w *PostgresWrapper) CreateAktion(ctx context.Context, arg repo.CreateAktionParams) (repo.Aktionen, error) {
	log.Printf("[DB] CreateAktion called for: %v", arg.Bezeichnung)
	res, err := w.pg.CreateAktion(ctx, repo_postgres.CreateAktionParams{
		AktionenKz:       toNullString(arg.AktionenKz),
		IDUser:           toNullInt32(arg.IDUser),
		Aktionsdatum:     arg.Aktionsdatum,
		Bezeichnung:      arg.Bezeichnung,
		IntervallTage:    toNullInt32(arg.IntervallTage),
		AnzahlIntervalle: toNullInt32(arg.AnzahlIntervalle),
		Erledigt:         toNullInt32(arg.Erledigt),
		IDUserErledigt:   toNullInt32(arg.IDUserErledigt),
		ErledigtAm:       toNullString(arg.ErledigtAm),
		Bemerkung:        toNullString(arg.Bemerkung),
	})
	if err != nil {
		log.Printf("[DB] CreateAktion ERROR: %v", err)
		return repo.Aktionen{}, err
	}
	id := int64(res.ID)
	log.Printf("[DB] CreateAktion SUCCESS, new ID: %d", id)
	return w.GetAktion(ctx, id)
}

func (w *PostgresWrapper) UpdateAktion(ctx context.Context, arg repo.UpdateAktionParams) (repo.Aktionen, error) {
	log.Printf("[DB] UpdateAktion called for ID: %d", arg.ID)
	_, err := w.pg.UpdateAktion(ctx, repo_postgres.UpdateAktionParams{
		AktionenKz:       toNullString(arg.AktionenKz),
		IDUser:           toNullInt32(arg.IDUser),
		Aktionsdatum:     arg.Aktionsdatum,
		Bezeichnung:      arg.Bezeichnung,
		IntervallTage:    toNullInt32(arg.IntervallTage),
		AnzahlIntervalle: toNullInt32(arg.AnzahlIntervalle),
		Erledigt:         toNullInt32(arg.Erledigt),
		IDUserErledigt:   toNullInt32(arg.IDUserErledigt),
		ErledigtAm:       toNullString(arg.ErledigtAm),
		Bemerkung:        toNullString(arg.Bemerkung),
		ID:               int32(arg.ID),
	})
	if err != nil {
		log.Printf("[DB] UpdateAktion ERROR: %v", err)
		return repo.Aktionen{}, err
	}
	log.Printf("[DB] UpdateAktion SUCCESS for ID: %d", arg.ID)
	return w.GetAktion(ctx, arg.ID)
}

func (w *PostgresWrapper) DeleteAktion(ctx context.Context, id int64) error {
	return w.pg.DeleteAktion(ctx, int32(id))
}

func convertPgAktion(v repo_postgres.Aktionen) repo.Aktionen {
	return repo.Aktionen{
		ID:               int64(v.ID),
		AktionenKz:       v.AktionenKz,
		IDUser:           sql.NullInt64{Int64: int64(v.IDUser.Int32), Valid: v.IDUser.Valid},
		Aktionsdatum:     v.Aktionsdatum,
		Bezeichnung:      v.Bezeichnung,
		IntervallTage:    sql.NullInt64{Int64: int64(v.IntervallTage.Int32), Valid: v.IntervallTage.Valid},
		AnzahlIntervalle: sql.NullInt64{Int64: int64(v.AnzahlIntervalle.Int32), Valid: v.AnzahlIntervalle.Valid},
		Erledigt:         sql.NullInt64{Int64: int64(v.Erledigt.Int32), Valid: v.Erledigt.Valid},
		IDUserErledigt:   sql.NullInt64{Int64: int64(v.IDUserErledigt.Int32), Valid: v.IDUserErledigt.Valid},
		ErledigtAm:       v.ErledigtAm,
		Bemerkung:        v.Bemerkung,
	}
}

func convertPgListAktionenRow(v repo_postgres.ListAktionenRow) repo.ListAktionenRow {
	return repo.ListAktionenRow{
		ID:               int64(v.ID),
		AktionenKz:       v.AktionenKz,
		IDUser:           toNullInt64(v.IDUser),
		Aktionsdatum:     v.Aktionsdatum,
		Bezeichnung:      v.Bezeichnung,
		IntervallTage:    toNullInt64(v.IntervallTage),
		AnzahlIntervalle: toNullInt64(v.AnzahlIntervalle),
		Erledigt:         toNullInt64(v.Erledigt),
		IDUserErledigt:   toNullInt64(v.IDUserErledigt),
		ErledigtAm:       v.ErledigtAm,
		Bemerkung:        v.Bemerkung,
		Username:         v.Username,
		UsernameErledigt: v.UsernameErledigt,
	}
}



