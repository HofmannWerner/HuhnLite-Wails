; =========================================================================
; HuhnLite - Inno Setup Installationsskript
; =========================================================================

; 1. Pfad-Definitionen (relativ zum Speicherort dieser .iss-Datei anpassen)
#define BuildDir  "..\build\bin"
#define ConfigDir "..\build\bin"

[Setup]
AppId={{A7C92418-4E2F-4B81-9136-19E4A6BD42D0}
AppName=HuhnLite
AppVersion=1.0
AppPublisher=HuhnLite
; Basisverzeichnis (wird im Code dynamisch je nach Komponente gesetzt)
DefaultDirName={autopf}\HuhnLite
DefaultGroupName=HuhnLite
OutputBaseFilename=HuhnLite-Setup
Compression=lzma2
SolidCompression=yes
ArchitecturesInstallIn64BitMode=x64compatible
FlatComponentsList=no
DisableProgramGroupPage=yes
; Zielordner-Seite ausblenden, da der Pfad fest an die DB-Auswahl gekoppelt ist
DisableDirPage=yes

[Types]
Name: "full";   Description: "Standard-Installation (PostgreSQL)"
Name: "custom"; Description: "Benutzerdefinierte Auswahl"; Flags: iscustom

[Components]
; Durch 'Flags: exclusive' kann immer nur genau eine Datenbank-Variante gewählt werden
Name: "db_pg";    Description: "PostgreSQL Umgebung"; Types: full; Flags: exclusive
Name: "db_maria"; Description: "MariaDB Umgebung";                 Flags: exclusive
Name: "db_std";   Description: "Standard Umgebung";                Flags: exclusive

[Dirs]
; --- AppData-Verzeichnisse je nach gewählter Komponente anlegen ---
Name: "{userappdata}\HuhnLite-Postgres"; Components: db_pg
Name: "{userappdata}\HuhnLite-MariaDB";  Components: db_maria
Name: "{userappdata}\HuhnLite";          Components: db_std

[Files]
; =========================================================================
; 1. PROGRAMMVERZEICHNIS ({app} in C:\Program Files\HuhnLite-...)
; =========================================================================

; Select-Anwendung (dieselbe Quell-EXE wird am Ziel passend umbenannt)
Source: "{#BuildDir}\HuhnLite-select-windows-amd64.exe"; DestDir: "{app}"; DestName: "HuhnLite-Select-Postgres.exe"; Components: db_pg;    Flags: ignoreversion
Source: "{#BuildDir}\HuhnLite-select-windows-amd64.exe"; DestDir: "{app}"; DestName: "HuhnLite-Select-MariaDB.exe";  Components: db_maria; Flags: ignoreversion
Source: "{#BuildDir}\HuhnLite-select-windows-amd64.exe"; DestDir: "{app}"; DestName: "HuhnLite-select.exe";         Components: db_std;   Flags: ignoreversion

; Server-Anwendungen (bleiben unverändert im jeweiligen Ordner)
Source: "{#BuildDir}\HuhnLite-server-postgres.exe"; DestDir: "{app}"; Components: db_pg;    Flags: ignoreversion
Source: "{#BuildDir}\HuhnLite-server-mariadb.exe";  DestDir: "{app}"; Components: db_maria; Flags: ignoreversion
Source: "{#BuildDir}\HuhnLite-server.exe";          DestDir: "{app}"; Components: db_std;   Flags: ignoreversion

; Local-Anwendungen (bleiben unverändert im jeweiligen Ordner)
Source: "{#BuildDir}\HuhnLite-postgres.exe"; DestDir: "{app}"; Components: db_pg;    Flags: ignoreversion
Source: "{#BuildDir}\HuhnLite-mariadb.exe";  DestDir: "{app}"; Components: db_maria; Flags: ignoreversion
Source: "{#BuildDir}\HuhnLite-Local.exe";    DestDir: "{app}"; Components: db_std;   Flags: ignoreversion

; =========================================================================
; 2. APPDATA-VERZEICHNIS ({userappdata}\...)
; =========================================================================
; 'onlyifdoesntexist' verhindert das Überschreiben bestehender Einstellungen bei Updates
Source: "{#ConfigDir}\settings_server_postgres.json"; DestDir: "{userappdata}\HuhnLite-Postgres"; Components: db_pg;    Flags: ignoreversion
Source: "{#ConfigDir}\settings_server_mariadb.json";  DestDir: "{userappdata}\HuhnLite-MariaDB";  Components: db_maria; Flags: ignoreversion
Source: "{#ConfigDir}\settings_server.json";  DestDir: "{userappdata}\HuhnLite";  Components: db_std;   Flags: ignoreversion
Source: "{#ConfigDir}\settings_postgres.json"; DestDir: "{userappdata}\HuhnLite-Postgres"; Components: db_pg;    Flags: ignoreversion
Source: "{#ConfigDir}\settings_mariadb.json";  DestDir: "{userappdata}\HuhnLite-MariaDB";  Components: db_maria; Flags: ignoreversion
Source: "{#ConfigDir}\settings.json";  DestDir: "{userappdata}\HuhnLite"; Components: db_std;   Flags: ignoreversion
Source: "{#ConfigDir}\HuhnLite_prod.db";  DestDir: "{userappdata}\HuhnLite-MariaDB";  Components: db_maria; Flags: ignoreversion
Source: "{#ConfigDir}\HuhnLite_test.db";  DestDir: "{userappdata}\HuhnLite-MariaDB";  Components: db_maria; Flags: ignoreversion
Source: "{#ConfigDir}\mandant_1\*"; DestDir: "{userappdata}\HuhnLite\mandant_1";  Components: db_std; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "{#ConfigDir}\mandant_2\*"; DestDir: "{userappdata}\HuhnLite\mandant_2";  Components: db_std;Flags: ignoreversion recursesubdirs createallsubdirs
[Icons]
; Startmenü: Select-Verknüpfungen
Name: "{group}\HuhnLite Select"; Filename: "{app}\HuhnLite-Select-Postgres.exe"; Components: db_pg
Name: "{group}\HuhnLite Select"; Filename: "{app}\HuhnLite-Select-MariaDB.exe";  Components: db_maria
Name: "{group}\HuhnLite Select"; Filename: "{app}\HuhnLite-select.exe";         Components: db_std

; Startmenü: Local-Verknüpfungen
Name: "{group}\HuhnLite Postgres"; Filename: "{app}\HuhnLite-Postgres.exe"; Components: db_pg
Name: "{group}\HuhnLite MariaDB"; Filename: "{app}\HuhnLite-MariaDB.exe";  Components: db_maria
Name: "{group}\HuhnLite Local"; Filename: "{app}\HuhnLite.exe";         Components: db_std

; Startmenü: Server-Verknüpfungen
Name: "{group}\HuhnLite Server"; Filename: "{app}\HuhnLite-server-postgres.exe"; Components: db_pg
Name: "{group}\HuhnLite Server"; Filename: "{app}\HuhnLite-server-mariadb.exe";  Components: db_maria
Name: "{group}\HuhnLite Server"; Filename: "{app}\HuhnLite-server.exe";          Components: db_std

; Startmenü: Deinstallation
Name: "{group}\HuhnLite deinstallieren"; Filename: "{uninstallexe}"

; Desktop-Verknüpfungen (optional)
Name: "{autodesktop}\HuhnLite Select"; Filename: "{app}\HuhnLite-Select-Postgres.exe"; Components: db_pg;    Tasks: desktopicon
Name: "{autodesktop}\HuhnLite Select"; Filename: "{app}\HuhnLite-Select-MariaDB.exe";  Components: db_maria; Tasks: desktopicon
Name: "{autodesktop}\HuhnLite Select"; Filename: "{app}\HuhnLite-select.exe";         Components: db_std;   Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "Desktop-Verknüpfung erstellen"; Flags: unchecked

[Run]
; Direktstart nach Abschluss der Installation
Filename: "{app}\HuhnLite-Select-Postgres.exe"; Description: "{cm:LaunchProgram,HuhnLite Select}"; Flags: nowait postinstall skipifsilent; Components: db_pg
Filename: "{app}\HuhnLite-Select-MariaDB.exe";  Description: "{cm:LaunchProgram,HuhnLite Select}"; Flags: nowait postinstall skipifsilent; Components: db_maria
Filename: "{app}\HuhnLite-select.exe";         Description: "{cm:LaunchProgram,HuhnLite Select}"; Flags: nowait postinstall skipifsilent; Components: db_std

[Code]
// Wird beim Klick auf "Weiter" ausgeführt und passt das Zielverzeichnis dynamisch an
function NextButtonClick(CurPageID: Integer): Boolean;
begin
  Result := True;

  if CurPageID = wpSelectComponents then
  begin
    if WizardIsComponentSelected('db_pg') then
      WizardForm.DirEdit.Text := ExpandConstant('{autopf}\HuhnLite-Postgres')
    else if WizardIsComponentSelected('db_maria') then
      WizardForm.DirEdit.Text := ExpandConstant('{autopf}\HuhnLite-MariaDB')
    else
      WizardForm.DirEdit.Text := ExpandConstant('{autopf}\HuhnLite');
  end;
end;