#!/bin/bash

# Find all .vue and .ts files in src
files=$(grep -l "http://localhost:8080" -r src --include="*.vue" --include="*.ts")

for f in $files; do
  echo "Processing $f ..."
  
  # Replace axios import with api import if it exists
  # Special case for BenutzerPage.vue etc already done manually but this handles others
  if grep -q "import axios from 'axios';" "$f"; then
    sed -i '' "s|import axios from 'axios';|import { api } from 'src/boot/api';|g" "$f"
  fi
  
  # Replace axios usage with api usage
  sed -i '' "s|axios\.|api\.|g" "$f"
  
  # Remove the hardcoded URL prefix
  sed -i '' "s|http://localhost:8080||g" "$f"
done

echo "Done."
