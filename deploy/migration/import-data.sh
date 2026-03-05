#!/bin/bash
# Orchestrate full HCMUS data import in the correct dependency order.
# Run bootstrap-admin.sh first to obtain ADMIN_TOKEN.
# Usage: ./deploy/migration/import-data.sh <API_URL> <ADMIN_TOKEN> [DATA_DIR]
set -euo pipefail

API_URL="${1:?Usage: import-data.sh <API_URL> <ADMIN_TOKEN> [DATA_DIR]}"
TOKEN="${2:?Usage: import-data.sh <API_URL> <ADMIN_TOKEN> [DATA_DIR]}"
DATA_DIR="${3:-deploy/migration/data}"

AUTH=(-H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")

api_post() {
  local endpoint="$1" payload="$2"
  curl -sf -X POST "$API_URL$endpoint" "${AUTH[@]}" -d "$payload" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('id','ok'))"
}

echo "=== Step 1: Create departments ==="
while IFS= read -r dept; do
  id=$(api_post "/api/hr/departments" "$dept")
  echo "  Created department: $id"
done < <(python3 -c "import json,sys; [print(json.dumps(d)) for d in json.load(open('$DATA_DIR/departments.json'))]")

echo "=== Step 2: Bulk import teachers ==="
curl -sf -X POST "$API_URL/api/admin/import/teachers" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@$DATA_DIR/teachers.csv" | python3 -c "import sys,json; r=json.load(sys.stdin); print(f'  Imported: {r.get(\"success\",0)} ok, {r.get(\"failed\",0)} failed')"

echo "=== Step 3: Create subjects ==="
while IFS= read -r subj; do
  id=$(api_post "/api/subject/subjects" "$subj")
  echo "  Created subject: $id"
done < <(python3 -c "import json,sys; [print(json.dumps(s)) for s in json.load(open('$DATA_DIR/subjects.json'))]")

echo "=== Step 4: Set prerequisites ==="
while IFS= read -r prereq; do
  api_post "/api/subject/prerequisites" "$prereq" > /dev/null
done < <(python3 -c "import json,sys; [print(json.dumps(p)) for p in json.load(open('$DATA_DIR/prerequisites.json'))]") || true
echo "  Prerequisites set."

echo "=== Step 5: Bulk import students ==="
curl -sf -X POST "$API_URL/api/admin/import/students" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@$DATA_DIR/students.csv" | python3 -c "import sys,json; r=json.load(sys.stdin); print(f'  Imported: {r.get(\"success\",0)} ok, {r.get(\"failed\",0)} failed')"

echo "=== Step 6: Create semester ==="
SEMESTER=$(cat "$DATA_DIR/semester.json")
api_post "/api/timetable/semesters" "$SEMESTER" > /dev/null
echo "  Semester created."

echo "=== Step 7: Create rooms ==="
while IFS= read -r room; do
  api_post "/api/timetable/rooms" "$room" > /dev/null
done < <(python3 -c "import json,sys; [print(json.dumps(r)) for r in json.load(open('$DATA_DIR/rooms.json'))]")
echo "  Rooms created."

echo ""
echo "✅ Import complete. Run verify-import.sh to validate data integrity."
