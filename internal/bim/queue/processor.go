package queue

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mintfary-oss/trest-sistems/internal/bim/converter"
)

type Processor struct {
	DB      *pgxpool.Pool
	WorkDir string
}

func (p *Processor) RunOnce(ctx context.Context) (bool, error) {
	var id, op, format, input, output, modelID, projectID, createdBy string
	var attempts, maxAttempts int
	tx, err := p.DB.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `SELECT id,project_id,COALESCE(bim_model_id::text,''),operation,format,input_uri,output_uri,attempts,max_attempts,created_by::text FROM bim_import_exports WHERE status='queued' AND attempts < max_attempts ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&id, &projectID, &modelID, &op, &format, &input, &output, &attempts, &maxAttempts, &createdBy)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	attempts++
	if _, err = tx.Exec(ctx, `UPDATE bim_import_exports SET status='running',attempts=$2,started_at=now(),error='' WHERE id=$1`, id, attempts); err != nil {
		return false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return false, err
	}
	out, err := p.process(ctx, op, format, input, output)
	if err != nil {
		status := "failed"
		if attempts < maxAttempts {
			status = "queued"
		}
		_, _ = p.DB.Exec(ctx, `UPDATE bim_import_exports SET status=$2,error=$3,completed_at=CASE WHEN $2='failed' THEN now() ELSE NULL END WHERE id=$1`, id, status, err.Error())
		return true, err
	}
	checksum, manifest, hashErr := fileManifest(out)
	if hashErr != nil {
		_, _ = p.DB.Exec(ctx, `UPDATE bim_import_exports SET status='failed',error=$2,completed_at=now() WHERE id=$1`, id, hashErr.Error())
		return true, hashErr
	}
	_, err = p.DB.Exec(ctx, `UPDATE bim_import_exports SET status='completed',completed_at=now(),error='',output_uri=$2,output_checksum=$3,output_manifest=$4::jsonb WHERE id=$1`, id, out, checksum, manifest)
	if err != nil {
		return true, err
	}
	if modelID != "" && (op == "import" || op == "export") {
		if err = p.createVersion(ctx, projectID, modelID, format, input, out, checksum, manifest, createdBy); err != nil {
			return true, err
		}
	}
	return true, nil
}

func (p *Processor) process(ctx context.Context, op, format, input, output string) (string, error) {
	if op != "import" && op != "export" {
		return "", fmt.Errorf("unsupported operation %q", op)
	}
	format = strings.ToLower(format)
	if format == "gltf" || format == "glb" {
		return "", fmt.Errorf("native %s conversion adapter not enabled", format)
	}
	if format != "obj" {
		return "", fmt.Errorf("format %s adapter pending", format)
	}
	if op == "import" {
		return p.importOBJ(input, output)
	}
	return "", fmt.Errorf("OBJ export requires a BIM source adapter")
}
func (p *Processor) importOBJ(input, output string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", fmt.Errorf("input_uri is empty")
	}
	r, err := os.Open(input)
	if err != nil {
		return "", err
	}
	defer r.Close()
	m, err := converter.ParseOBJ(r)
	if err != nil {
		return "", err
	}
	if output == "" {
		output = filepath.Join(p.WorkDir, filepath.Base(input)+".gltf")
	}
	if err = os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return "", err
	}
	w, err := os.Create(output)
	if err != nil {
		return "", err
	}
	defer w.Close()
	if err = converter.EncodeGLTF(w, m); err != nil {
		return "", err
	}
	return output, nil
}

func fileManifest(path string) (string, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	sum := sha256.Sum256(b)
	m := map[string]any{"path": path, "size_bytes": len(b), "sha256": hex.EncodeToString(sum[:])}
	raw, err := json.Marshal(m)
	return hex.EncodeToString(sum[:]), string(raw), err
}
func (p *Processor) createVersion(ctx context.Context, projectID, modelID, sourceFormat, input, output, checksum, manifest, createdBy string) error {
	var exists string
	if err := p.DB.QueryRow(ctx, `SELECT project_id::text FROM bim_models WHERE id=$1`, modelID).Scan(&exists); err != nil {
		return err
	}
	if exists != projectID {
		return fmt.Errorf("BIM model does not belong to project")
	}
	var version int
	if err := p.DB.QueryRow(ctx, `SELECT COALESCE(MAX(version),0)+1 FROM bim_model_versions WHERE bim_model_id=$1`, modelID).Scan(&version); err != nil {
		return err
	}
	_, err := p.DB.Exec(ctx, `INSERT INTO bim_model_versions(bim_model_id,version,source_format,source_uri,geometry_uri,manifest,checksum,created_by) VALUES($1,$2,$3,$4,$5,$6::jsonb,$7,$8) ON CONFLICT(bim_model_id,version) DO NOTHING`, modelID, version, sourceFormat, input, output, manifest, checksum, createdBy)
	return err
}
