import sqlite3

conn = sqlite3.connect("HuhnLite.db")
cursor = conn.cursor()

print("FUTTER table schema:")
cursor.execute("PRAGMA table_info(FUTTER)")
for row in cursor.fetchall():
    print(row)

print("\nBUCHUNG table schema:")
cursor.execute("PRAGMA table_info(BUCHUNG)")
for row in cursor.fetchall():
    print(row)

conn.close()
