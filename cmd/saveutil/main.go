// saveutil is an admin CLI for Erupe save data management.
//
// Usage:
//
//	saveutil import    --config config.json --char-id 42 --file export.json
//	saveutil export    --config config.json --char-id 42 [--output export.json]
//	saveutil grant-import --config config.json --char-id 42 [--ttl 24h]
//	saveutil revoke-import --config config.json --char-id 42
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"erupe-ce/server/channelserver/compression/nullcomp"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// dbConfig is the minimal config subset needed to connect to PostgreSQL.
type dbConfig struct {
	Database struct {
		Host     string `json:"Host"`
		Port     int    `json:"Port"`
		User     string `json:"User"`
		Password string `json:"Password"`
		Database string `json:"Database"`
	} `json:"Database"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "import":
		err = runImport(args)
	case "export":
		err = runExport(args)
	case "grant-import":
		err = runGrantImport(args)
	case "revoke-import":
		err = runRevokeImport(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `saveutil — Erupe save data admin tool

Commands:
  import       --config config.json --char-id N --file export.json
  export       --config config.json --char-id N [--output file.json]
  grant-import --config config.json --char-id N [--ttl 24h]
  revoke-import --config config.json --char-id N`)
}

// openDB parses config.json and returns an open database connection.
func openDB(configPath string) (*sqlx.DB, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg dbConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	dsn := fmt.Sprintf(
		"host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode=disable",
		cfg.Database.Host, cfg.Database.Port,
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Database,
	)
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

// generateToken returns a 32-byte cryptographically random hex token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// --- import ---

func runImport(args []string) error {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	configPath := fs.String("config", "config.json", "Path to config.json")
	charID := fs.Uint("char-id", 0, "Destination character ID")
	filePath := fs.String("file", "", "Path to export JSON file (required)")
	_ = fs.Parse(args)

	if *charID == 0 {
		return errors.New("--char-id is required")
	}
	if *filePath == "" {
		return errors.New("--file is required")
	}

	db, err := openDB(*configPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// Read and parse the export JSON.
	raw, err := os.ReadFile(*filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	var export struct {
		Character map[string]interface{} `json:"character"`
	}
	if err := json.Unmarshal(raw, &export); err != nil {
		return fmt.Errorf("parse export JSON: %w", err)
	}
	if export.Character == nil {
		return errors.New("export JSON has no 'character' key")
	}

	blobs, err := extractAllBlobs(export.Character)
	if err != nil {
		return fmt.Errorf("extract blobs: %w", err)
	}

	// Compute savedata hash.
	var savedataHash []byte
	if len(blobs["savedata"]) > 0 {
		decompressed, err := nullcomp.Decompress(blobs["savedata"])
		if err != nil {
			return fmt.Errorf("decompress savedata: %w", err)
		}
		h := sha256.Sum256(decompressed)
		savedataHash = h[:]
	}

	_, err = db.Exec(
		`UPDATE characters SET
			savedata=$1, savedata_hash=$2, decomyset=$3, hunternavi=$4,
			otomoairou=$5, partner=$6, platebox=$7, platedata=$8,
			platemyset=$9, rengokudata=$10, savemercenary=$11, gacha_items=$12,
			house_info=$13, login_boost=$14, skin_hist=$15, scenariodata=$16,
			savefavoritequest=$17, mezfes=$18,
			savedata_import_token=NULL, savedata_import_token_expiry=NULL
		 WHERE id=$19`,
		blobs["savedata"], savedataHash, blobs["decomyset"], blobs["hunternavi"],
		blobs["otomoairou"], blobs["partner"], blobs["platebox"], blobs["platedata"],
		blobs["platemyset"], blobs["rengokudata"], blobs["savemercenary"], blobs["gacha_items"],
		blobs["house_info"], blobs["login_boost"], blobs["skin_hist"], blobs["scenariodata"],
		blobs["savefavoritequest"], blobs["mezfes"],
		*charID,
	)
	if err != nil {
		return fmt.Errorf("update characters: %w", err)
	}
	fmt.Printf("Save data imported into character %d\n", *charID)
	return nil
}

// --- export ---

func runExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	configPath := fs.String("config", "config.json", "Path to config.json")
	charID := fs.Uint("char-id", 0, "Character ID to export")
	outputPath := fs.String("output", "", "Output file (default: stdout)")
	_ = fs.Parse(args)

	if *charID == 0 {
		return errors.New("--char-id is required")
	}

	db, err := openDB(*configPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	row := db.QueryRowx("SELECT * FROM characters WHERE id=$1", *charID)
	result := make(map[string]interface{})
	if err := row.MapScan(result); err != nil {
		return fmt.Errorf("query character: %w", err)
	}

	export := map[string]interface{}{"character": result}
	enc := json.NewEncoder(os.Stdout)
	if *outputPath != "" {
		f, err := os.Create(*outputPath)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer func() { _ = f.Close() }()
		enc = json.NewEncoder(f)
	}
	enc.SetIndent("", "  ")
	if err := enc.Encode(export); err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}
	if *outputPath != "" {
		fmt.Printf("Character %d exported to %s\n", *charID, *outputPath)
	}
	return nil
}

// --- grant-import ---

func runGrantImport(args []string) error {
	fs := flag.NewFlagSet("grant-import", flag.ExitOnError)
	configPath := fs.String("config", "config.json", "Path to config.json")
	charID := fs.Uint("char-id", 0, "Character ID to grant import permission for")
	ttl := fs.Duration("ttl", 24*time.Hour, "Token validity duration (e.g. 24h, 48h)")
	_ = fs.Parse(args)

	if *charID == 0 {
		return errors.New("--char-id is required")
	}

	db, err := openDB(*configPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("generate token: %w", err)
	}
	expiry := time.Now().Add(*ttl)

	res, err := db.Exec(
		`UPDATE characters SET savedata_import_token=$1, savedata_import_token_expiry=$2 WHERE id=$3`,
		token, expiry, *charID,
	)
	if err != nil {
		return fmt.Errorf("update characters: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("character %d not found", *charID)
	}

	fmt.Printf("Import token for character %d (expires %s):\n%s\n",
		*charID, expiry.Format(time.RFC3339), token)
	return nil
}

// --- revoke-import ---

func runRevokeImport(args []string) error {
	fs := flag.NewFlagSet("revoke-import", flag.ExitOnError)
	configPath := fs.String("config", "config.json", "Path to config.json")
	charID := fs.Uint("char-id", 0, "Character ID to revoke import permission for")
	_ = fs.Parse(args)

	if *charID == 0 {
		return errors.New("--char-id is required")
	}

	db, err := openDB(*configPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec(
		`UPDATE characters SET savedata_import_token=NULL, savedata_import_token_expiry=NULL WHERE id=$1`,
		*charID,
	)
	if err != nil {
		return fmt.Errorf("update characters: %w", err)
	}
	fmt.Printf("Import token revoked for character %d\n", *charID)
	return nil
}

// blobColumns is the ordered list of transferable save blob column names.
var blobColumns = []string{
	"savedata", "decomyset", "hunternavi", "otomoairou", "partner",
	"platebox", "platedata", "platemyset", "rengokudata", "savemercenary",
	"gacha_items", "house_info", "login_boost", "skin_hist", "scenariodata",
	"savefavoritequest", "mezfes",
}

// extractAllBlobs decodes all save blob columns from a character export map.
func extractAllBlobs(m map[string]interface{}) (map[string][]byte, error) {
	out := make(map[string][]byte, len(blobColumns))
	for _, col := range blobColumns {
		v, ok := m[col]
		if !ok || v == nil {
			out[col] = nil
			continue
		}
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("column %q: expected string, got %T", col, v)
		}
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("column %q: base64: %w", col, err)
		}
		out[col] = b
	}
	return out, nil
}
