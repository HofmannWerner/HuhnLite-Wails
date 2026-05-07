package api

import (
	"context"
	"database/sql"
	db "huhnlite-wails/backend/db/repo"
	"math"
)

// GewichtAufbereiten bereitet das Durchschnittsgewicht für den Tabellen-Lookup vor.
// 1. Rundung auf eine Dezimalstelle.
// 2. Falls die Dezimalstelle ungerade ist (,1, ,3, ,5, ,7), ründe auf die nächste gerade Zahl auf (,2, ,4, ,6, ,8).
// 3. Falls die Dezimalstelle ,9 ist, ründe auf ,0 auf und erhöhe die Vorkommastahl um 1.
func GewichtAufbereiten(val float64) float64 {
	rounded := math.Round(val*10) / 10
	integerPart := math.Floor(rounded)
	decimalPart := int(math.Round((rounded - integerPart) * 10))

	if decimalPart == 9 {
		return integerPart + 1
	}
	if decimalPart%2 != 0 {
		return integerPart + float64(decimalPart+1)/10
	}
	return rounded
}

// GetAlterWoche berechnet das Alter (AW) basierend auf der Legewoche (LW) und dem Legebeginn.
// Formel gemäß USER-Anforderung: AW = LW + Legebeginn_LW
// Hinweis: LW ist die Woche seit Legebeginn (startend bei 1).
func GetAlterWoche(lw int, legebeginn int) int {
	return lw + legebeginn
}

// GetEigewichtByAlter ermittelt das Durchschnittsgewicht für eine bestimmte Lebenswoche.
// 1. Holt aus TabellenKopf mit ID_TabelleAlter die TabellenNummer.
// 2. Holt mit TabellenNummer und Lebenswoche aus Tabelle LSLKlassik den Wert EigewichtWoche.
func GetEigewichtByAlter(ctx context.Context, q db.Querier, idTabelleAlter int64, aw int) (float64, error) {
	// 1. TN aus Tabellenkopf via ID_TabelleAlter
	tkopf, err := q.GetTabellenkopf(ctx, idTabelleAlter)
	if err != nil {
		return 0, err
	}

	// 2. Eigewicht aus LSLKlassik mit TN und Lebenswoche (AW)
	params := db.GetLSLByTabNumAndAgeParams{
		Tabellennummer: tkopf.Tabellennummer,
		Alterinwochen:  int64(aw),
	}
	lslEntry, err := q.GetLSLByTabNumAndAge(ctx, params)
	if err != nil {
		return 0, err
	}

	if true {
		return lslEntry.Eigewichtwo, nil
	}
	return 0, nil
}

// CalculateEggDistribution (Helper) berechnet die Stückzahlen (S, M, L, XL) aus einer Gewichttabelle.
func CalculateEggDistribution(totalA int64, weightEntry db.Gewichttabelle) (int64, int64, int64, int64) {
	if totalA <= 0 {
		return 0, 0, 0, 0
	}

	pSmall := weightEntry.Klasse7 + weightEntry.Klasse6
	pMedium := weightEntry.Klasse5 + weightEntry.Klasse4
	pLarge := weightEntry.Klasse3 + weightEntry.Klasse2
	pXl := weightEntry.Klasse1

	// 1. Initial berechnen durch Rundung
	xl := int64(math.Round(float64(totalA) * (pXl * 0.01)))
	large := int64(math.Round(float64(totalA) * (pLarge * 0.01)))
	medium := int64(math.Round(float64(totalA) * (pMedium * 0.01)))
	small := int64(math.Round(float64(totalA) * (pSmall * 0.01)))

	// 2. Differenz zur Zielsumme ermitteln
	currentSum := xl + large + medium + small
	diff := totalA - currentSum

	// 3. Differenz ausgleichen (beim größten Posten, um relative Fehler gering zu halten)
	if diff != 0 {
		if xl >= large && xl >= medium && xl >= small {
			xl += diff
		} else if large >= medium && large >= small {
			large += diff
		} else if medium >= small {
			medium += diff
		} else {
			small += diff
		}
	}

	// Sicherheitshalber negative Werte verhindern
	if xl < 0 {
		xl = 0
	}
	if large < 0 {
		large = 0
	}
	if medium < 0 {
		medium = 0
	}
	if small < 0 {
		small = 0
	}

	return small, medium, large, xl
}

// GetEierverteilungByGewicht führt die vollständige Berechnung basierend auf dem Gewicht aus.
// 1. Holt die Tabellennummer aus Tabellenkopf via idTabelleGewicht.
// 2. Sucht in GewichtTabelle nach TN und Eigewicht = dGewicht.
// 3. Berechnet die Stückzahlen S, M, L, XL.
func GetEierverteilungByGewicht(ctx context.Context, q db.Querier, totalA int64, idTabelleGewicht int64, dGewicht float64) (int64, int64, int64, int64, error) {
	// 1. TN aus Tabellenkopf
	tkopf, err := q.GetTabellenkopf(ctx, idTabelleGewicht)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// 2. Aufbereitetes Gewicht für Lookup
	aufbereitet := GewichtAufbereiten(dGewicht)

	// 3. Gewichtseintrag suchen
	params := db.GetGewichtByTabNumAndWeightParams{
		Tabellennummer: tkopf.Tabellennummer,
		ROUND:          aufbereitet,
	}
	entry, err := q.GetGewichtByTabNumAndWeight(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			// Fallback: Alles auf Small
			return totalA, 0, 0, 0, nil
		}
		return 0, 0, 0, 0, err
	}

	// 4. Verteilung berechnen
	s, m, l, xl := CalculateEggDistribution(totalA, entry)
	return s, m, l, xl, nil
}

// GetEierverteilungByAlter führt die vollständige Berechnung basierend auf der Legewoche/Alter aus.
// 1. Ermittelt EigewichtWoche aus LSLKlassik via AW (Lebenswoche).
// 2. Verwendet das ermittelte Gewicht für GetEierverteilungByGewicht.
func GetEierverteilungByAlter(ctx context.Context, q db.Querier, totalA int64, idTabelleAlter int64, idTabelleGewicht int64, lw int, legebeginn int) (int64, int64, int64, int64, error) {
	// 1. AW berechnen
	aw := GetAlterWoche(lw, legebeginn)

	// 2. Eigewicht aus LSLKlassik via GetEigewichtByAlter
	dGewicht, err := GetEigewichtByAlter(ctx, q, idTabelleAlter, aw)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// 3. Verteilung wie bei Gewichtsbasiert (via idTabelleGewicht)
	return GetEierverteilungByGewicht(ctx, q, totalA, idTabelleGewicht, dGewicht)
}

// GetDurchschnittsgewicht (Helper) berechnet das Durchschnittsgewicht aus Kontrollwiegung.
// Formel: ((KontrollGewicht - Verpackung) * 1000) / AnzahlKontrolle
func GetDurchschnittsgewicht(kontrollgewicht float64, verpackung float64, anzahl int64) float64 {
	if anzahl <= 0 {
		return 0
	}
	// 1. Nettogewicht in kg
	nettoKg := kontrollgewicht - verpackung
	// 2. Umrechnung in g
	grammTotal := nettoKg * 1000.0
	// 3. Division durch Anzahl
	dGewicht := grammTotal / float64(anzahl)

	return dGewicht
}
