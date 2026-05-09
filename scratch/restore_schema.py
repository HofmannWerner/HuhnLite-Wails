import re

with open('backend/db/schema_sqlite.sql', 'r') as f:
    content = f.read()

# 1. Basic conversions (Keep uppercase for tables/columns to match existing queries)
content = content.replace('AUTOINCREMENT', 'AUTO_INCREMENT')
content = content.replace('BOOLEAN', 'TINYINT(1)')

def convert_to_mysql(text):
    # Remove SQLite comments
    text = re.sub(r'--.*$', '', text, flags=re.MULTILINE)
    
    # Merge into single lines per statement for easier regex
    text = re.sub(r'\s+', ' ', text)
    statements = text.split(';')
    new_stmts = []
    for stmt in statements:
        stmt = stmt.strip()
        if not stmt: continue
        
        # Change TEXT to VARCHAR(191) for keys and fields with defaults
        # We keep the case of the original SQL (mostly UPPERCASE)
        words = stmt.split()
        new_words = []
        for i, word in enumerate(words):
            upper_word = word.upper()
            if upper_word == 'TEXT':
                col_name = words[i-1].strip('(),').upper()
                # Specific long fields stay TEXT
                if col_name in ['SQLSTATEMENT', 'SQLSTATEMENT_NATIVE', 'DETAIL_SQL', 'DETAIL_SQL_NATIVE', 'BESCHREIBUNG', 'INHALT', 'BEMERKUNG', 'VALUE']:
                    new_words.append('TEXT')
                else:
                    new_words.append('VARCHAR(191)')
            else:
                new_words.append(word)
        
        stmt = " ".join(new_words)
        
        # Fix ID columns
        if 'PRIMARY KEY (ID)' in stmt.upper():
            stmt = re.sub(r',\s*PRIMARY\s+KEY\s*\(ID\)', '', stmt, flags=re.IGNORECASE)
            stmt = re.sub(r'\bID\s+INTEGER\b', 'ID INTEGER PRIMARY KEY AUTO_INCREMENT', stmt, flags=re.IGNORECASE)
        
        # Ensure TEXT columns don't have DEFAULT (MySQL restriction)
        stmt = re.sub(r'TEXT\s+NOT\s+NULL\s+DEFAULT\s+\'[^\']*\'', 'TEXT NOT NULL', stmt, flags=re.IGNORECASE)
        stmt = re.sub(r'TEXT\s+DEFAULT\s+\'[^\']*\'', 'TEXT', stmt, flags=re.IGNORECASE)
        
        # Final cleanup
        stmt = stmt.strip(';')
        if not stmt.endswith(';'): stmt += ';'
        new_stmts.append(stmt)
    
    return ";\n\n".join(new_stmts)

mysql_schema = convert_to_mysql(content)

# Add system tables
mysql_schema += """

CREATE TABLE IF NOT EXISTS SYSTEMSETTINGS (
    NAME VARCHAR(191) PRIMARY KEY,
    VALUE TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS USER_STATE (
    ID INTEGER PRIMARY KEY AUTO_INCREMENT,
    USERNAME VARCHAR(191) NOT NULL,
    `KEY` VARCHAR(191) NOT NULL,
    VALUE TEXT NOT NULL,
    UNIQUE(USERNAME, `KEY`)
);
"""

with open('backend/db/schema_mysql.sql', 'w') as f:
    f.write(mysql_schema)
