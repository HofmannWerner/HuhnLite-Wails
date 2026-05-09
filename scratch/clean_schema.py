import re

with open('backend/db/schema_mysql.sql', 'r') as f:
    content = f.read()

# Fix missing semicolons before CREATE TABLE
content = re.sub(r'([^\s;])\s+CREATE\s+TABLE', r'\1; CREATE TABLE', content, flags=re.IGNORECASE)

# Now standard split and cleanup
statements = content.split(';')
new_statements = []

for stmt in statements:
    stmt = stmt.strip()
    if not stmt: continue
    
    # Standardize spaces
    stmt = re.sub(r'\s+', ' ', stmt)
    
    # Fix ID AUTO_INCREMENT
    stmt = re.sub(r'ID\s+INTEGER\s+PRIMARY\s+KEY\s+AUTO_INCREMENT', 'ID INTEGER PRIMARY KEY AUTO_INCREMENT', stmt, flags=re.IGNORECASE)
    
    # Standardize TEXT/VARCHAR
    stmt = re.sub(r'TEXT\s*\(\s*(\d+)\s*\)', r'VARCHAR(\1)', stmt, flags=re.IGNORECASE)
    stmt = stmt.replace('VARCHAR(255)', 'VARCHAR(191)')
    
    # Fix common fields
    for col in ['KZ', 'SPRACHE_KZ', 'USERNAME', 'TEXT_TYP_KZ', 'BEZEICHNUNG', 'BEWEGUNGSDATUM', 'BUCHUNGSDATUM', 'TEMPLATE_NAME', 'PARAM_DEF', 'DETAIL_SQL', 'LINK_LOGIC', 'GROUP_FIELD', 'SYSTEM_KZ', 'SUMMENZEILE', 'EIERKLASSE', 'TVNAME', 'CHARGE']:
        stmt = re.sub(r'\b' + col + r'\s+TEXT\b', col + ' VARCHAR(191)', stmt, flags=re.IGNORECASE)
    
    # MySQL doesn't allow DEFAULT on TEXT
    if 'TEXT' in stmt.upper() and 'DEFAULT' in stmt.upper():
        stmt = re.sub(r'([a-zA-Z0-9_]+)\s+TEXT(\s+NOT\s+NULL)?\s+DEFAULT', r'\1 VARCHAR(191)\2 DEFAULT', stmt, flags=re.IGNORECASE)

    new_statements.append(stmt)

with open('backend/db/schema_mysql.sql', 'w') as f:
    for stmt in new_statements:
        f.write(stmt + ";\n\n")
