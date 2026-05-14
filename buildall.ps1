# buildall.ps1

# Pfad zu Wails (prüft ob es im PATH ist, sonst Standard-Go-Pfad)
$WAILS = "wails"
if (!(Get-Command $WAILS -ErrorAction SilentlyContinue)) {
    $WAILS = "$env:USERPROFILE\go\bin\wails.exe"
}

# Funktion zum Ändern des Namens in der wails.json
function Set-WailsName($newName) {
    $config = Get-Content wails.json | ConvertFrom-Json
    $config.name = $newName
    $config.outputfilename = $newName
    $config | ConvertTo-Json -Depth 10 | Set-Content wails.json
}

# Original-Namen sichern
if (Test-Path "wails.json") {
    $originalConfig = Get-Content wails.json | ConvertFrom-Json
    $origName = $originalConfig.name
} else {
    Write-Host "❌ Fehler: wails.json nicht gefunden!" -ForegroundColor Red
    exit
}

Write-Host "------------------------------------------" -ForegroundColor Cyan
Write-Host "🔨 Baue HuhnLite-Local (SQLite)..."
Set-WailsName "HuhnLite-Local"
& $WAILS build

if (Test-Path "settings.json") {
    if (!(Test-Path "build\bin")) { New-Item -ItemType Directory -Path "build\bin" -Force }
    Copy-Item "settings.json" "build\bin\settings.json" -Force
    Write-Host "✅ settings.json kopiert." -ForegroundColor Gray
}

Write-Host "------------------------------------------" -ForegroundColor Cyan
Write-Host "🔨 Baue HuhnLite-MariaDB (Netzwerk)..."
Set-WailsName "HuhnLite-MariaDB"
& $WAILS build

if (Test-Path "settings_mariadb.json") {
    if (!(Test-Path "build\bin")) { New-Item -ItemType Directory -Path "build\bin" -Force }
    Copy-Item "settings_mariadb.json" "build\bin\settings_mariadb.json" -Force
    Write-Host "✅ settings_mariadb.json kopiert." -ForegroundColor Gray
}

# Original-Zustand wiederherstellen
Set-WailsName $origName

Write-Host "------------------------------------------" -ForegroundColor Green
Write-Host "✅ Fertig! Die Dateien wurden in build\bin\ bereitgestellt."
if (Test-Path "build\bin") {
    Get-ChildItem "build\bin\*.json"
}
