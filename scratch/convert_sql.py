import re

def convert_sqlite_to_mysql(sql):
    # 1. Replace RETURNING ... ; with ;
    sql = re.sub(r'(?i)returning\s+.*?;', ';', sql, flags=re.DOTALL)
    
    # 2. ON CONFLICT(...) DO UPDATE SET ... -> ON DUPLICATE KEY UPDATE ...
    def replace_upsert(match):
        columns_part = match.group(1)
        mysql_columns = re.sub(r'(?i)excluded\.([a-zA-Z0-9_]+)', r'VALUES(\1)', columns_part)
        return f"ON DUPLICATE KEY UPDATE {mysql_columns}"

    sql = re.sub(r'(?i)on conflict\([^)]+\)\s+do\s+update\s+set\s+(.*?)(?=\s*;|\Z)', replace_upsert, sql, flags=re.DOTALL)

    # 3. strftime -> YEAR/MONTH
    sql = re.sub(r"(?i)strftime\('%y',\s*([^)]+)\)", r"YEAR(\1)", sql)
    sql = re.sub(r"(?i)strftime\('%m',\s*([^)]+)\)", r"MONTH(\1)", sql)
    
    # 4. datetime('now') -> NOW()
    sql = re.sub(r"(?i)datetime\('now'\)", "NOW()", sql)
    sql = re.sub(r"(?i)date\('now'\)", "CURDATE()", sql)

    # 5. ROWID -> id
    sql = re.sub(r"(?i)rowid", "id", sql)
    
    # 5b. cast(... as integer) -> cast(... as signed)
    sql = re.sub(r"(?i)as\s+integer", "as signed", sql)

    lines = sql.split('\n')
    new_lines = []
    for i, line in enumerate(lines):
        if '-- name:' in line and ':one' in line:
            found_write = False
            for j in range(i+1, min(i+10, len(lines))):
                if re.search(r'^\s*(INSERT|UPDATE|insert|update)\s', lines[j]):
                    found_write = True
                    break
                if ';' in lines[j]: break
            
            if found_write:
                line = line.replace(':one', ':execresult')
        
        new_lines.append(line)
    
    return '\n'.join(new_lines)

with open('backend/db/queries.sql', 'r') as f:
    content = f.read()

# We keep the case of the original queries.sql
converted = convert_sqlite_to_mysql(content)

with open('backend/db/queries_mysql.sql', 'w') as f:
    f.write(converted)
