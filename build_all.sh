#!/bin/bash
set -e

# Funktion für robustes hdiutil mit Retries und sync
run_hdiutil() {
    local max_attempts=3
    local attempt=1
    sync
    until "$@" || [ $attempt -eq $max_attempts ]; do
        echo "⚠️ hdiutil failed (attempt $attempt/$max_attempts), retrying in 5 seconds..."
        hdiutil detach "/Volumes/HuhnLite"* -force 2>/dev/null || true
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
BUILD_ARGS=("build" "-v" "2")
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

# Pre-build backup of PDFs if they exist in build/bin/
mkdir -p build/local_temp
for f in build/bin/HuhnLite_*.PDF build/bin/HuhnLite_*.pdf; do
    [ -f "$f" ] && cp "$f" build/local_temp/
done

echo "------------------------------------------"
echo "🔨 Baue HuhnLite-Local (SQLite)..."
set_wails_name "HuhnLite-Local"
echo "DEBUG: wails.json content for Local:"
cat wails.json
$WAILS "${BUILD_ARGS[@]}"
echo "DEBUG: build/bin content after Local build:"
ls -la build/bin || true
# Kopiere die passende settings.json daneben falls noch nicht vorhanden
[ ! -f build/bin/settings.json ] && cp settings.json build/bin/settings.json

# Erstelle Mandanten-Verzeichnisse in build/bin und kopiere Standard-DBs hinein falls nicht vorhanden
mkdir -p build/bin/mandant_1 build/bin/mandant_2
[ -f HuhnLite.db ] && [ ! -f build/bin/mandant_1/HuhnLite.db ] && cp HuhnLite.db build/bin/mandant_1/HuhnLite.db
[ -f HuhnLite.db ] && [ ! -f build/bin/mandant_2/HuhnLite.db ] && cp HuhnLite.db build/bin/mandant_2/HuhnLite.db
[ -f HuhnLite_test.db ] && [ ! -f build/bin/mandant_1/HuhnLite_test.db ] && cp HuhnLite_test.db build/bin/mandant_1/HuhnLite_test.db
[ -f HuhnLite_test.db ] && [ ! -f build/bin/mandant_2/HuhnLite_test.db ] && cp HuhnLite_test.db build/bin/mandant_2/HuhnLite_test.db
[ -f HuhnLite_prod.db ] && [ ! -f build/bin/mandant_1/HuhnLite_prod.db ] && cp HuhnLite_prod.db build/bin/mandant_1/HuhnLite_prod.db
[ -f HuhnLite_prod.db ] && [ ! -f build/bin/mandant_2/HuhnLite_prod.db ] && cp HuhnLite_prod.db build/bin/mandant_2/HuhnLite_prod.db

# Copy help files, settings and db into the macOS .app bundle if it exists
if [ -d "build/bin/HuhnLite-Local.app" ]; then
    echo "📋 Copying help files and config into HuhnLite-Local.app..."
    cp settings.json build/bin/HuhnLite-Local.app/Contents/Resources/settings.json
    [ -f settings_server.json ] && cp settings_server.json build/bin/HuhnLite-Local.app/Contents/Resources/settings_server.json
    [ -f settings_server_mariadb.json ] && cp settings_server_mariadb.json build/bin/HuhnLite-Local.app/Contents/Resources/settings_server_mariadb.json
    [ -f HuhnLite.db ] && cp HuhnLite.db build/bin/HuhnLite-Local.app/Contents/Resources/HuhnLite.db
    [ -f HuhnLite_test.db ] && cp HuhnLite_test.db build/bin/HuhnLite-Local.app/Contents/Resources/HuhnLite_test.db
    [ -f HuhnLite_prod.db ] && cp HuhnLite_prod.db build/bin/HuhnLite-Local.app/Contents/Resources/HuhnLite_prod.db
    
    # Copy PDF help files to macOS bundle (from root or local_temp backup)
    for f in HuhnLite_*.PDF HuhnLite_*.pdf build/local_temp/HuhnLite_*.PDF build/local_temp/HuhnLite_*.pdf; do
        [ -f "$f" ] && cp "$f" build/bin/HuhnLite-Local.app/Contents/Resources/
    done

    # Copy PDF.js to macOS bundle
    if [ -d pdfjs ]; then
        mkdir -p build/bin/HuhnLite-Local.app/Contents/Resources/pdfjs
        cp -R pdfjs/* build/bin/HuhnLite-Local.app/Contents/Resources/pdfjs/
    fi
    
    # Create mandant directories inside macOS bundle
    mkdir -p build/bin/HuhnLite-Local.app/Contents/Resources/mandant_1 build/bin/HuhnLite-Local.app/Contents/Resources/mandant_2
    [ -f HuhnLite.db ] && cp HuhnLite.db build/bin/HuhnLite-Local.app/Contents/Resources/mandant_1/HuhnLite.db && cp HuhnLite.db build/bin/HuhnLite-Local.app/Contents/Resources/mandant_2/HuhnLite.db
    [ -f HuhnLite_test.db ] && cp HuhnLite_test.db build/bin/HuhnLite-Local.app/Contents/Resources/mandant_1/HuhnLite_test.db && cp HuhnLite_test.db build/bin/HuhnLite-Local.app/Contents/Resources/mandant_2/HuhnLite_test.db
    [ -f HuhnLite_prod.db ] && cp HuhnLite_prod.db build/bin/HuhnLite-Local.app/Contents/Resources/mandant_1/HuhnLite_prod.db && cp HuhnLite_prod.db build/bin/HuhnLite-Local.app/Contents/Resources/mandant_2/HuhnLite_prod.db
fi

# Vor dem nächsten Wails-Build die HuhnLite-Local Build-Ergebnisse temporär sichern,
# da Wails bei jedem Build den Inhalt von build/bin/ leert!
mkdir -p build/local_temp
[ -d build/bin/HuhnLite-Local.app ] && mv build/bin/HuhnLite-Local.app build/local_temp/
[ -f build/bin/settings.json ] && mv build/bin/settings.json build/local_temp/
[ -f build/bin/HuhnLite.db ] && mv build/bin/HuhnLite.db build/local_temp/
[ -f build/bin/HuhnLite_test.db ] && mv build/bin/HuhnLite_test.db build/local_temp/
[ -f build/bin/HuhnLite_prod.db ] && mv build/bin/HuhnLite_prod.db build/local_temp/
[ -d build/bin/mandant_1 ] && mv build/bin/mandant_1 build/local_temp/
[ -d build/bin/mandant_2 ] && mv build/bin/mandant_2 build/local_temp/
[ -d build/bin/pdfjs ] && mv build/bin/pdfjs build/local_temp/
for f in build/bin/HuhnLite_*.PDF build/bin/HuhnLite_*.pdf; do
    [ -f "$f" ] && mv "$f" build/local_temp/
done

echo "------------------------------------------"
echo "🔨 Baue HuhnLite-MariaDB (Netzwerk)..."
set_wails_name "HuhnLite-MariaDB"
echo "DEBUG: wails.json content for MariaDB:"
cat wails.json
$WAILS "${BUILD_ARGS[@]}"
echo "DEBUG: build/bin content after MariaDB build:"
ls -la build/bin || true
# Kopiere die passende settings_mariadb.json daneben falls nicht vorhanden
[ ! -f build/bin/settings_mariadb.json ] && cp settings_mariadb.json build/bin/settings_mariadb.json

# Copy help files and config into the macOS .app bundle if it exists
if [ -d "build/bin/HuhnLite-MariaDB.app" ]; then
    echo "📋 Copying help files and config into HuhnLite-MariaDB.app..."
    cp settings_mariadb.json build/bin/HuhnLite-MariaDB.app/Contents/Resources/settings_mariadb.json
    [ -f settings_server.json ] && cp settings_server.json build/bin/HuhnLite-MariaDB.app/Contents/Resources/settings_server.json
    [ -f settings_server_mariadb.json ] && cp settings_server_mariadb.json build/bin/HuhnLite-MariaDB.app/Contents/Resources/settings_server_mariadb.json
    
    # Copy PDF help files to MariaDB macOS bundle (from root or local_temp backup)
    for f in HuhnLite_*.PDF HuhnLite_*.pdf build/local_temp/HuhnLite_*.PDF build/local_temp/HuhnLite_*.pdf; do
        [ -f "$f" ] && cp "$f" build/bin/HuhnLite-MariaDB.app/Contents/Resources/
    done

    # Copy PDF.js to MariaDB macOS bundle
    if [ -d pdfjs ]; then
        mkdir -p build/bin/HuhnLite-MariaDB.app/Contents/Resources/pdfjs
        cp -R pdfjs/* build/bin/HuhnLite-MariaDB.app/Contents/Resources/pdfjs/
    fi
fi

# Zurückbewegen der HuhnLite-Local Build-Ergebnisse nach build/bin/
[ -d build/local_temp/HuhnLite-Local.app ] && mv build/local_temp/HuhnLite-Local.app build/bin/
[ -f build/local_temp/settings.json ] && mv build/local_temp/settings.json build/bin/
[ -f build/local_temp/HuhnLite.db ] && mv build/local_temp/HuhnLite.db build/bin/
[ -f build/local_temp/HuhnLite_test.db ] && mv build/local_temp/HuhnLite_test.db build/bin/
[ -f build/local_temp/HuhnLite_prod.db ] && mv build/local_temp/HuhnLite_prod.db build/bin/
[ -d build/local_temp/mandant_1 ] && mv build/local_temp/mandant_1 build/bin/
[ -d build/local_temp/mandant_2 ] && mv build/local_temp/mandant_2 build/bin/
[ -d build/local_temp/pdfjs ] && mv build/local_temp/pdfjs build/bin/
for f in build/local_temp/HuhnLite_*.PDF build/local_temp/HuhnLite_*.pdf; do
    [ -f "$f" ] && mv "$f" build/bin/
done
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

[ ! -f build/bin/settings_server.json ] && cp settings_server.json build/bin/settings_server.json

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

[ ! -f build/bin/settings_server_mariadb.json ] && cp settings_server_mariadb.json build/bin/settings_server_mariadb.json

echo "------------------------------------------"
echo "🔨 Baue HuhnLite-Server-Postgres (Client-Server Postgres)..."
if [ -n "$PLATFORM" ]; then
    export GOOS="$TARGET_OS"
    export GOARCH="$TARGET_ARCH"
fi

SERVER_POSTGRES_NAME="HuhnLite-Server-Postgres"
if [ "$TARGET_OS" = "windows" ]; then
    SERVER_POSTGRES_NAME="HuhnLite-Server-Postgres.exe"
fi

go build -ldflags="-w -s" -o "build/bin/$SERVER_POSTGRES_NAME" .

if [ -n "$PLATFORM" ]; then
    unset GOOS
    unset GOARCH
fi

[ ! -f build/bin/settings_server_postgres.json ] && cp settings_server_postgres.json build/bin/settings_server_postgres.json

echo "📋 Copying help files and SQL files to build/bin..."
for f in HuhnLite_*.PDF HuhnLite_*.pdf build/local_temp/HuhnLite_*.PDF build/local_temp/HuhnLite_*.pdf; do
    [ -f "$f" ] && cp "$f" build/bin/
done
if [ -d pdfjs ]; then
    mkdir -p build/bin/pdfjs
    cp -R pdfjs/* build/bin/pdfjs/
fi
[ -f HuhnLite.db ] && [ ! -f build/bin/HuhnLite.db ] && cp HuhnLite.db build/bin/
[ -f HuhnLite_test.db ] && [ ! -f build/bin/HuhnLite_test.db ] && cp HuhnLite_test.db build/bin/
[ -f HuhnLite_prod.db ] && [ ! -f build/bin/HuhnLite_prod.db ] && cp HuhnLite_prod.db build/bin/

# Original-Zustand wiederherstellen
set_wails_name "$ORIG_NAME"

# Temp-Verzeichnis aufräumen
rm -rf build/local_temp

echo "------------------------------------------"
echo "✅ Fertig! Die Dateien wurden in build/bin/ bereitgestellt."
ls -lh build/bin/*.json || true

if [ "$(uname)" = "Darwin" ]; then
    echo "------------------------------------------"
    echo "📦 Erstelle macOS DMG-Dateien..."
    hdiutil detach "/Volumes/HuhnLite"* -force 2>/dev/null || true
    if [ -d "build/bin/HuhnLite-Local.app" ]; then
        run_hdiutil hdiutil create -volname "HuhnLite Local" -srcfolder "build/bin/HuhnLite-Local.app" -size 500m -ov -format UDZO "build/bin/HuhnLite-Local.dmg"
        sleep 3
    fi
    if [ -d "build/bin/HuhnLite-MariaDB.app" ]; then
        run_hdiutil hdiutil create -volname "HuhnLite MariaDB" -srcfolder "build/bin/HuhnLite-MariaDB.app" -size 500m -ov -format UDZO "build/bin/HuhnLite-MariaDB.dmg"
        sleep 3
    fi
    if [ -f "build/bin/HuhnLite-Server" ]; then
        mkdir -p build/bin/server-temp
        cp build/bin/HuhnLite-Server build/bin/server-temp/
        cp build/bin/settings_server.json build/bin/server-temp/settings_server.json
        [ -f HuhnLite.db ] && cp HuhnLite.db build/bin/server-temp/
        [ -f HuhnLite_test.db ] && cp HuhnLite_test.db build/bin/server-temp/
        [ -f HuhnLite_prod.db ] && cp HuhnLite_prod.db build/bin/server-temp/

        cat << 'EOF' > build/bin/server-temp/install.command
#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TARGET_DIR="$HOME/Library/Application Support/HuhnLite"
mkdir -p "$TARGET_DIR"
cp "$DIR/settings_server.json" "$TARGET_DIR/settings_server.json" 2>/dev/null || true
[ -f "$DIR/HuhnLite.db" ] && cp -n "$DIR/HuhnLite.db" "$TARGET_DIR/HuhnLite.db"
[ -f "$DIR/HuhnLite_test.db" ] && cp -n "$DIR/HuhnLite_test.db" "$TARGET_DIR/HuhnLite_test.db"
[ -f "$DIR/HuhnLite_prod.db" ] && cp -n "$DIR/HuhnLite_prod.db" "$TARGET_DIR/HuhnLite_prod.db"
echo "✅ HuhnLite-Server-Dateien nach $TARGET_DIR installiert."
echo "Starte HuhnLite-Server..."
"$DIR/HuhnLite-Server"
EOF
        chmod +x build/bin/server-temp/install.command

        run_hdiutil hdiutil create -volname "HuhnLite Server" -srcfolder build/bin/server-temp -size 300m -ov -format UDZO "build/bin/HuhnLite-Server.dmg"
        rm -rf build/bin/server-temp
        sleep 3
    fi
    if [ -f "build/bin/HuhnLite-Server-MariaDB" ]; then
        mkdir -p build/bin/server-mariadb-temp
        cp build/bin/HuhnLite-Server-MariaDB build/bin/server-mariadb-temp/
        cp build/bin/settings_server_mariadb.json build/bin/server-mariadb-temp/settings_server_mariadb.json

        cat << 'EOF' > build/bin/server-mariadb-temp/install.command
#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TARGET_DIR="$HOME/Library/Application Support/HuhnLite"
mkdir -p "$TARGET_DIR"
cp "$DIR/settings_server_mariadb.json" "$TARGET_DIR/settings_server_mariadb.json" 2>/dev/null || true
echo "✅ HuhnLite-Server-MariaDB-Dateien nach $TARGET_DIR installiert."
echo "Starte HuhnLite-Server-MariaDB..."
"$DIR/HuhnLite-Server-MariaDB"
EOF
        chmod +x build/bin/server-mariadb-temp/install.command

        run_hdiutil hdiutil create -volname "HuhnLite Server MariaDB" -srcfolder build/bin/server-mariadb-temp -size 300m -ov -format UDZO "build/bin/HuhnLite-Server-MariaDB.dmg"
        rm -rf build/bin/server-mariadb-temp
        sleep 3
    fi
    if [ -f "build/bin/HuhnLite-Server-Postgres" ]; then
        mkdir -p build/bin/server-postgres-temp
        cp build/bin/HuhnLite-Server-Postgres build/bin/server-postgres-temp/
        cp build/bin/settings_server_postgres.json build/bin/server-postgres-temp/settings_server_postgres.json

        cat << 'EOF' > build/bin/server-postgres-temp/install.command
#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TARGET_DIR="$HOME/Library/Application Support/HuhnLite"
mkdir -p "$TARGET_DIR"
cp "$DIR/settings_server_postgres.json" "$TARGET_DIR/settings_server_postgres.json" 2>/dev/null || true
echo "✅ HuhnLite-Server-Postgres-Dateien nach $TARGET_DIR installiert."
echo "Starte HuhnLite-Server-Postgres..."
"$DIR/HuhnLite-Server-Postgres"
EOF
        chmod +x build/bin/server-postgres-temp/install.command

        run_hdiutil hdiutil create -volname "HuhnLite Server Postgres" -srcfolder build/bin/server-postgres-temp -size 300m -ov -format UDZO "build/bin/HuhnLite-Server-Postgres.dmg"
        rm -rf build/bin/server-postgres-temp
        sleep 3
    fi
fi
