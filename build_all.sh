#!/bin/bash
set -e

# Platform Parameter (z.B. linux/amd64 oder windows/amd64)
PLATFORM=$1

# Pfad zu Wails (falls nicht im PATH)
WAILS=~/go/bin/wails

# Build Argumente aufbauen
BUILD_ARGS=("build")
if [ -n "$PLATFORM" ]; then
    BUILD_ARGS+=("-platform" "$PLATFORM")
fi

# Funktion zum Ändern des Namens in der wails.json
set_wails_name() {
    local new_name=$1
    sed "s/\"name\": \".*\"/\"name\": \"$new_name\"/" wails.json > wails.json.tmp && mv wails.json.tmp wails.json
    sed "s/\"outputfilename\": \".*\"/\"outputfilename\": \"$new_name\"/" wails.json > wails.json.tmp && mv wails.json.tmp wails.json
}

# Original-Namen sichern
ORIG_NAME=$(grep '"name":' wails.json | head -n 1 | cut -d'"' -f4 | tr -d '\n\r ')

echo "------------------------------------------"
echo "🔨 Baue HuhnLite-Local (SQLite)..."
set_wails_name "HuhnLite-Local"
$WAILS "${BUILD_ARGS[@]}"
# Kopiere die passende settings.json daneben
cp settings.json build/bin/settings.json

echo "------------------------------------------"
echo "🔨 Baue HuhnLite-MariaDB (Netzwerk)..."
set_wails_name "HuhnLite-MariaDB"
$WAILS "${BUILD_ARGS[@]}"
# Kopiere die passende settings_mariadb.json daneben
cp settings_mariadb.json build/bin/settings_mariadb.json

echo "------------------------------------------"
echo "🔨 Baue HuhnLite-Server (Client-Server SQLite)..."
TARGET_OS="linux"
if [ -n "$PLATFORM" ]; then
    IFS='/' read -r -a platform_parts <<< "$PLATFORM"
    TARGET_OS="${platform_parts[0]}"
    TARGET_ARCH="${platform_parts[1]}"
    export GOOS="$TARGET_OS"
    export GOARCH="$TARGET_ARCH"
fi

SERVER_NAME="HuhnLite-Server"
if [ "$TARGET_OS" = "windows" ]; then
    SERVER_NAME="HuhnLite-Server.exe"
fi

go build -ldflags="-w -s" -o "build/bin/$SERVER_NAME" .

if [ -n "$PLATFORM" ]; then
    unset GOOS
    unset GOARCH
fi

cp settings_server.json build/bin/settings_server.json

echo "------------------------------------------"
echo "🔨 Baue HuhnLite-Server-MariaDB (Client-Server MariaDB)..."
if [ -n "$PLATFORM" ]; then
    export GOOS="$TARGET_OS"
    export GOARCH="$TARGET_ARCH"
fi

SERVER_MARIADB_NAME="HuhnLite-Server-MariaDB"
if [ "$TARGET_OS" = "windows" ]; then
    SERVER_MARIADB_NAME="HuhnLite-Server-MariaDB.exe"
fi

go build -ldflags="-w -s" -o "build/bin/$SERVER_MARIADB_NAME" .

if [ -n "$PLATFORM" ]; then
    unset GOOS
    unset GOARCH
fi

cp settings_server_mariadb.json build/bin/settings_server_mariadb.json

# Original-Zustand wiederherstellen
set_wails_name "$ORIG_NAME"

echo "------------------------------------------"
echo "✅ Fertig! Die Dateien wurden in build/bin/ bereitgestellt."
ls -lh build/bin/*.json || true
