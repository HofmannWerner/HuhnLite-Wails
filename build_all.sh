#!/bin/bash
set -e

# Funktion für robustes hdiutil mit Retries und sync
run_hdiutil() {
    local max_attempts=3
    local attempt=1
    sync
    until "$@" || [ $attempt -eq $max_attempts ]; do
        echo "⚠️ hdiutil failed (attempt $attempt/$max_attempts), retrying in 5 seconds..."
        sleep 5
        sync
        attempt=$((attempt + 1))
    done
    if [ $attempt -eq $max_attempts ]; then
        "$@"
    fi
}

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
    if command -v jq >/dev/null 2>&1; then
        jq --arg name "$new_name" '.name = $name | .outputfilename = $name' wails.json > wails.json.tmp && mv wails.json.tmp wails.json
    else
        sed "s/\"name\": \".*\"/\"name\": \"$new_name\"/" wails.json > wails.json.tmp && mv wails.json.tmp wails.json
        sed "s/\"outputfilename\": \".*\"/\"outputfilename\": \"$new_name\"/" wails.json > wails.json.tmp && mv wails.json.tmp wails.json
    fi
}

# Original-Namen sichern
if command -v jq >/dev/null 2>&1; then
    ORIG_NAME=$(jq -r '.name' wails.json)
else
    ORIG_NAME=$(grep '"name":' wails.json | head -n 1 | cut -d'"' -f4 | tr -d '\n\r ')
fi

echo "------------------------------------------"
echo "🔨 Baue HuhnLite-Local (SQLite)..."
set_wails_name "HuhnLite-Local"
echo "DEBUG: wails.json content for Local:"
cat wails.json
$WAILS "${BUILD_ARGS[@]}"
echo "DEBUG: build/bin content after Local build:"
ls -la build/bin || true
# Kopiere die passende settings.json daneben
cp settings.json build/bin/settings.json

# Copy help files, images, settings and db into the macOS .app bundle if it exists
if [ -d "build/bin/HuhnLite-Local.app" ]; then
    echo "📋 Copying help files, images and config into HuhnLite-Local.app..."
    cp HuhnLite-de.html HuhnLite-en.html HuhnLite-it.html build/bin/HuhnLite-Local.app/Contents/Resources/
    mkdir -p build/bin/HuhnLite-Local.app/Contents/Resources/images
    cp -R images/* build/bin/HuhnLite-Local.app/Contents/Resources/images/
    cp settings.json build/bin/HuhnLite-Local.app/Contents/Resources/settings.json
    [ -f HuhnLite.db ] && cp HuhnLite.db build/bin/HuhnLite-Local.app/Contents/Resources/HuhnLite.db
    [ -f HuhnLite_test.db ] && cp HuhnLite_test.db build/bin/HuhnLite-Local.app/Contents/Resources/HuhnLite_test.db
    [ -f HuhnLite_prod.db ] && cp HuhnLite_prod.db build/bin/HuhnLite-Local.app/Contents/Resources/HuhnLite_prod.db
fi

# Sichern der HuhnLite-Local Build-Ergebnisse, da der nächste Wails-Build build/bin/ löscht
mkdir -p build/local_temp
[ -d build/bin/HuhnLite-Local.app ] && mv build/bin/HuhnLite-Local.app build/local_temp/
[ -f build/bin/settings.json ] && mv build/bin/settings.json build/local_temp/
[ -f build/bin/HuhnLite.db ] && mv build/bin/HuhnLite.db build/local_temp/
[ -f build/bin/HuhnLite_test.db ] && mv build/bin/HuhnLite_test.db build/local_temp/
[ -f build/bin/HuhnLite_prod.db ] && mv build/bin/HuhnLite_prod.db build/local_temp/

echo "------------------------------------------"
echo "🔨 Baue HuhnLite-MariaDB (Netzwerk)..."
set_wails_name "HuhnLite-MariaDB"
echo "DEBUG: wails.json content for MariaDB:"
cat wails.json
$WAILS "${BUILD_ARGS[@]}"
echo "DEBUG: build/bin content after MariaDB build:"
ls -la build/bin || true
# Kopiere die passende settings_mariadb.json daneben
cp settings_mariadb.json build/bin/settings_mariadb.json

# Copy help files, images and config into the macOS .app bundle if it exists
if [ -d "build/bin/HuhnLite-MariaDB.app" ]; then
    echo "📋 Copying help files, images and config into HuhnLite-MariaDB.app..."
    cp HuhnLite-de.html HuhnLite-en.html HuhnLite-it.html build/bin/HuhnLite-MariaDB.app/Contents/Resources/
    mkdir -p build/bin/HuhnLite-MariaDB.app/Contents/Resources/images
    cp -R images/* build/bin/HuhnLite-MariaDB.app/Contents/Resources/images/
    cp settings_mariadb.json build/bin/HuhnLite-MariaDB.app/Contents/Resources/settings_mariadb.json
fi

# Zurückbewegen der HuhnLite-Local Build-Ergebnisse nach build/bin/
[ -d build/local_temp/HuhnLite-Local.app ] && mv build/local_temp/HuhnLite-Local.app build/bin/
[ -f build/local_temp/settings.json ] && mv build/local_temp/settings.json build/bin/
[ -f build/local_temp/HuhnLite.db ] && mv build/local_temp/HuhnLite.db build/bin/
[ -f build/local_temp/HuhnLite_test.db ] && mv build/local_temp/HuhnLite_test.db build/bin/
[ -f build/local_temp/HuhnLite_prod.db ] && mv build/local_temp/HuhnLite_prod.db build/bin/
rm -rf build/local_temp

echo "------------------------------------------"
echo "🔨 Baue HuhnLite-Server (Client-Server SQLite)..."
TARGET_OS="linux"
if [ "$(uname)" = "Darwin" ]; then
    TARGET_OS="darwin"
fi
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

echo "📋 Copying help files, images and SQL files to build/bin..."
cp HuhnLite-de.html HuhnLite-en.html HuhnLite-it.html build/bin/
mkdir -p build/bin/images
cp -R images/* build/bin/images/
[ -f HuhnLite.db ] && cp HuhnLite.db build/bin/
[ -f HuhnLite_test.db ] && cp HuhnLite_test.db build/bin/
[ -f HuhnLite_prod.db ] && cp HuhnLite_prod.db build/bin/

# Original-Zustand wiederherstellen
set_wails_name "$ORIG_NAME"

echo "------------------------------------------"
echo "✅ Fertig! Die Dateien wurden in build/bin/ bereitgestellt."
ls -lh build/bin/*.json || true

if [ "$(uname)" = "Darwin" ]; then
    echo "------------------------------------------"
    echo "📦 Erstelle macOS DMG-Dateien..."
    if [ -d "build/bin/HuhnLite-Local.app" ]; then
        run_hdiutil hdiutil create -volname "HuhnLite Local" -srcfolder "build/bin/HuhnLite-Local.app" -ov -format UDZO "build/bin/HuhnLite-Local.dmg"
        sleep 3
    fi
    if [ -d "build/bin/HuhnLite-MariaDB.app" ]; then
        run_hdiutil hdiutil create -volname "HuhnLite MariaDB" -srcfolder "build/bin/HuhnLite-MariaDB.app" -ov -format UDZO "build/bin/HuhnLite-MariaDB.dmg"
        sleep 3
    fi
    if [ -f "build/bin/HuhnLite-Server" ]; then
        mkdir -p build/bin/server-temp
        cp build/bin/HuhnLite-Server build/bin/server-temp/
        cp build/bin/settings_server.json build/bin/server-temp/settings_server.json
        run_hdiutil hdiutil create -volname "HuhnLite Server" -srcfolder build/bin/server-temp -ov -format UDZO "build/bin/HuhnLite-Server.dmg"
        rm -rf build/bin/server-temp
        sleep 3
    fi
    if [ -f "build/bin/HuhnLite-Server-MariaDB" ]; then
        mkdir -p build/bin/server-mariadb-temp
        cp build/bin/HuhnLite-Server-MariaDB build/bin/server-mariadb-temp/
        cp build/bin/settings_server_mariadb.json build/bin/server-mariadb-temp/settings_server_mariadb.json
        run_hdiutil hdiutil create -volname "HuhnLite Server MariaDB" -srcfolder build/bin/server-mariadb-temp -ov -format UDZO "build/bin/HuhnLite-Server-MariaDB.dmg"
        rm -rf build/bin/server-mariadb-temp
        sleep 3
    fi
fi
