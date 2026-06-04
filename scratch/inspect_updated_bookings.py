import sqlite3
import os
import sys

paths = [
    "C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db",
    "C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/HuhnLite.db",
    os.path.join(os.environ.get("APPDATA", ""), "HuhnLite-Wails", "HuhnLite.db")
]

for p in paths:
    print(f"\n=========================================")
    print(f"DATABASE: {p}")
    print(f"=========================================")
    if not os.path.exists(p):
        print("File not found.")
        continue
    try:
        conn = sqlite3.connect(p)
        cursor = conn.cursor()
        
        # Schema info
        cursor.execute("PRAGMA table_info(BUCHUNG)")
        cols = [row[1] for row in cursor.fetchall()]
        print(f"BUCHUNG columns: {cols}")
        
        cursor.execute("""
            SELECT ID, ID_HERDEN, BUCHUNGSDATUM, SILONR, FUTTERVERBRAUCHTIER, FUTTERKTAG, TIERBESTAND 
            FROM BUCHUNG 
            WHERE SILONR = 1 AND (FUTTERVERBRAUCHTIER > 0 OR FUTTERKTAG > 0)
            ORDER BY BUCHUNGSDATUM DESC LIMIT 10
        """)
        rows = cursor.fetchall()
        print(f"Found {len(rows)} matching updated rows:")
        for r in rows:
            print(f"  ID: {r[0]} | Herd: {r[1]} | Date: {r[2]} | Silo: {r[3]} | Verbrauch: {r[4]} | Ktag: {r[5]} | Bestand: {r[6]}")
            
        conn.close()
    except Exception as e:
        print(f"Error checking: {e}")
