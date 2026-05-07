import os
import re

vue_dir = '/Users/wernerhofmann/Projekte/HuhnLite/frontend/src'

for root, _, files in os.walk(vue_dir):
    for file in files:
        if file.endswith('.vue') or file.endswith('.ts'):
            path = os.path.join(root, file)
            with open(path, 'r') as f:
                content = f.read()

            # 1. Clean ?.String -> nothing   (so `row.NAME?.String` becomes `row.NAME`)
            content = re.sub(r'\?\.String', r'', content)
            # 1b. Clean .String -> nothing
            content = re.sub(r'\.String\b', r'', content)

            # 2. Clean ?.Int64 -> nothing
            content = re.sub(r'\?\.Int64', r'', content)
            # 2b. Clean .Int64 -> nothing
            content = re.sub(r'\.Int64\b', r'', content)

            # 3. Clean ?.Float64 -> nothing
            content = re.sub(r'\?\.Float64', r'', content)
            # 3b. Clean .Float64 -> nothing
            content = re.sub(r'\.Float64\b', r'', content)

            # 4. Clean ?.Valid -> true
            # This is tricky because `row.NAME?.Valid ? ... : ...`
            # Let's replace ternary `row.NAME?.Valid ? row.NAME.String : row.NAME` with `row.NAME`
            content = re.sub(r'([a-zA-Z0-9_\.\(\)\[\]]+)\??\.Valid\s*\?\s*\1\.?String\s*:\s*\1', r'\1', content)
            content = re.sub(r'([a-zA-Z0-9_\.\(\)\[\]]+)\??\.Valid\s*\?\s*\1\.?Int64\s*:\s*\1', r'\1', content)

            # General ?.Valid check (if it stands alone) replace with just truthy check
            content = re.sub(r'\?\.Valid', r'', content)
            content = re.sub(r'\.Valid\b', r'', content)

            # If there's `row.ID_PERSONENTYP.Int64`, the `\.Int64` removes it. 

            with open(path, 'w') as f:
                f.write(content)

print("Frontend null wrappers removed.")
