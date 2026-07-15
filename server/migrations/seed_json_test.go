package migrations

import (
	"strconv"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestRawSQLValue(t *testing.T) {
	tests := []struct {
		name    string
		val     interface{}
		wantRaw string
		wantIs  bool
		wantErr bool
	}{
		{"not an object", float64(5), "", false, false},
		{"string value", "hello", "", false, false},
		{"valid raw", map[string]interface{}{"raw": "NOW()"}, "NOW()", true, false},
		{"missing raw key", map[string]interface{}{"other": "x"}, "", false, true},
		{"extra key", map[string]interface{}{"raw": "NOW()", "other": "x"}, "", false, true},
		{"non-string raw value", map[string]interface{}{"raw": 5.0}, "", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, isRaw, err := rawSQLValue(tt.val)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if isRaw != tt.wantIs || got != tt.wantRaw {
				t.Errorf("got (%q, %v), want (%q, %v)", got, isRaw, tt.wantRaw, tt.wantIs)
			}
		})
	}
}

func TestNormalizeSeedValue(t *testing.T) {
	tests := []struct {
		name string
		val  interface{}
		want interface{}
	}{
		{"integral float", float64(500000), int64(500000)},
		{"negative integral float", float64(-1), int64(-1)},
		{"fractional float", float64(1.5), float64(1.5)},
		{"string passthrough", "abc", "abc"},
		{"bool passthrough", true, true},
		{"nil passthrough", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSeedValue(tt.val)
			if got != tt.want {
				t.Errorf("got %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestRowColumns(t *testing.T) {
	row := map[string]interface{}{"type": "personal", "points_req": 500000, "repeatable": false}
	got := rowColumns(row)
	want := []string{"points_req", "repeatable", "type"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
			break
		}
	}
}

func TestHasExactColumns(t *testing.T) {
	columns := []string{"a", "b"}
	tests := []struct {
		name string
		row  map[string]interface{}
		want bool
	}{
		{"exact match", map[string]interface{}{"a": 1, "b": 2}, true},
		{"missing key", map[string]interface{}{"a": 1}, false},
		{"extra key", map[string]interface{}{"a": 1, "b": 2, "c": 3}, false},
		{"different key", map[string]interface{}{"a": 1, "c": 2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasExactColumns(tt.row, columns); got != tt.want {
				t.Errorf("hasExactColumns(%v, %v) = %v, want %v", tt.row, columns, got, tt.want)
			}
		})
	}
}

func TestBuildSeedInsert(t *testing.T) {
	block := seedJSONBlock{
		Table:      "diva_prizes",
		OnConflict: "DO NOTHING",
		Rows: []map[string]interface{}{
			{"type": "personal", "points_req": float64(500000), "repeatable": false},
			{"type": "guild", "points_req": float64(1000000), "repeatable": false},
		},
	}

	query, args, err := buildSeedInsert(block)
	if err != nil {
		t.Fatalf("buildSeedInsert failed: %v", err)
	}

	// Columns are sorted alphabetically for a deterministic query:
	// points_req, repeatable, type.
	want := "INSERT INTO diva_prizes (points_req, repeatable, type) VALUES " +
		"($1, $2, $3), ($4, $5, $6) ON CONFLICT DO NOTHING"
	if query != want {
		t.Errorf("query = %q, want %q", query, want)
	}
	if len(args) != 6 {
		t.Fatalf("len(args) = %d, want 6", len(args))
	}
	if args[0] != int64(500000) {
		t.Errorf("args[0] = %v (%T), want int64(500000)", args[0], args[0])
	}
}

func TestBuildSeedInsert_RawValue(t *testing.T) {
	block := seedJSONBlock{
		Table: "campaigns",
		Rows: []map[string]interface{}{
			{"id": float64(1), "start_time": map[string]interface{}{"raw": "NOW()"}},
		},
	}

	query, args, err := buildSeedInsert(block)
	if err != nil {
		t.Fatalf("buildSeedInsert failed: %v", err)
	}
	want := "INSERT INTO campaigns (id, start_time) VALUES ($1, NOW())"
	if query != want {
		t.Errorf("query = %q, want %q", query, want)
	}
	if len(args) != 1 {
		t.Fatalf("len(args) = %d, want 1 (raw value must not be bound as a parameter)", len(args))
	}
}

func TestBuildSeedInsert_ColumnMismatch(t *testing.T) {
	block := seedJSONBlock{
		Table: "diva_prizes",
		Rows: []map[string]interface{}{
			{"type": "personal", "points_req": float64(1)},
			{"type": "guild"},
		},
	}
	if _, _, err := buildSeedInsert(block); err == nil {
		t.Error("expected error for a row whose columns differ from row 0, got nil")
	}
}

func TestApplySeedJSONBlock_InvalidTableName(t *testing.T) {
	block := seedJSONBlock{
		Table: "diva_prizes; DROP TABLE users",
		Rows:  []map[string]interface{}{{"type": "personal"}},
	}
	err := applySeedJSONBlock(nil, block)
	if err == nil || !strings.Contains(err.Error(), "invalid table name") {
		t.Errorf("expected invalid table name error, got %v", err)
	}
}

func TestApplySeedJSONBlock_InvalidColumnName(t *testing.T) {
	block := seedJSONBlock{
		Table: "diva_prizes",
		Rows:  []map[string]interface{}{{"type; DROP TABLE users": "personal"}},
	}
	err := applySeedJSONBlock(nil, block)
	if err == nil || !strings.Contains(err.Error(), "invalid column name") {
		t.Errorf("expected invalid column name error, got %v", err)
	}
}

func TestApplySeedJSONBlock_EmptyRowsNoOp(t *testing.T) {
	block := seedJSONBlock{
		Table: "diva_prizes",
		Rows:  []map[string]interface{}{},
	}
	// No DB call should happen for an empty block, so a nil *sqlx.DB must not panic.
	if err := applySeedJSONBlock(nil, block); err != nil {
		t.Errorf("expected no error for empty rows, got %v", err)
	}
}

func TestApplySeedJSON_InvalidJSON(t *testing.T) {
	err := applySeedJSON(nil, "broken.json", []byte("{not json"))
	if err == nil {
		t.Error("expected parse error for malformed JSON, got nil")
	}
}

func TestApplySeedJSON_Integration(t *testing.T) {
	db := testDB(t)
	defer func() { _ = db.Close() }()

	logger, _ := zap.NewDevelopment()
	if _, err := Migrate(db, logger); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	data := []byte(`{
		"blocks": [
			{
				"table": "diva_prizes",
				"onConflict": "DO NOTHING",
				"rows": [
					{"type": "personal", "points_req": 42, "item_type": 26, "item_id": 0, "quantity": 1, "gr": false, "repeatable": false}
				]
			}
		]
	}`)

	if err := applySeedJSON(db, "test.json", data); err != nil {
		t.Fatalf("applySeedJSON failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM diva_prizes WHERE points_req = 42").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row with points_req=42, got %d", count)
	}
}

// TestApplySeedJSON_OnConflict uses a scratch table with its own unique
// constraint, rather than a real seed table, since none of the current seed
// tables enforce content uniqueness (their "ON CONFLICT DO NOTHING" only
// guards against a literal primary-key clash, which a fresh serial id never
// hits) - this test instead proves the onConflict clause itself is wired
// through and honored when the schema does enforce uniqueness.
func TestApplySeedJSON_OnConflict(t *testing.T) {
	db := testDB(t)
	defer func() { _ = db.Close() }()

	if _, err := db.Exec("CREATE TABLE seed_conflict_test (k INT UNIQUE, v INT)"); err != nil {
		t.Fatalf("creating scratch table failed: %v", err)
	}

	data := []byte(`{
		"blocks": [
			{
				"table": "seed_conflict_test",
				"onConflict": "(k) DO NOTHING",
				"rows": [{"k": 1, "v": 100}]
			}
		]
	}`)

	if err := applySeedJSON(db, "test.json", data); err != nil {
		t.Fatalf("applySeedJSON failed: %v", err)
	}
	if err := applySeedJSON(db, "test.json", data); err != nil {
		t.Fatalf("second applySeedJSON failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM seed_conflict_test").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected onConflict (k) DO NOTHING to dedup on re-apply, got %d rows", count)
	}
}

func TestApplySeedJSON_Truncate(t *testing.T) {
	db := testDB(t)
	defer func() { _ = db.Close() }()

	logger, _ := zap.NewDevelopment()
	if _, err := Migrate(db, logger); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	seedTwice := func(quantity int) {
		data := []byte(`{
			"blocks": [
				{
					"table": "public.cafebonus",
					"truncate": true,
					"rows": [{"time_req": 1800, "item_type": 17, "item_id": 0, "quantity": ` + strconv.Itoa(quantity) + `}]
				}
			]
		}`)
		if err := applySeedJSON(db, "test.json", data); err != nil {
			t.Fatalf("applySeedJSON failed: %v", err)
		}
	}

	seedTwice(50)
	seedTwice(99)

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM cafebonus").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected truncate to leave exactly 1 row after reseeding, got %d", count)
	}

	var quantity int
	if err := db.QueryRow("SELECT quantity FROM cafebonus").Scan(&quantity); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if quantity != 99 {
		t.Errorf("expected latest seeded quantity 99, got %d", quantity)
	}
}
