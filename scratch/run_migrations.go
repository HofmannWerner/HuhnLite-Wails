package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

func toCamelCase(s string) string {
	parts := strings.Split(strings.ToLower(s), "_")
	for i, part := range parts {
		if i > 0 && len(part) > 0 {
			parts[i] = strings.ToUpper(part[0:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func main() {
	db, err := sql.Open("sqlite", "C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer db.Close()

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

	fmt.Println("Migration run complete on SQLite database.")
}
