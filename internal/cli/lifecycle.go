package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func composeFile(root string) string { return filepath.Join(root, "deployments", "docker-compose.yml") }

func runCompose(ctx context.Context, root string, args ...string) error {
	return command(ctx, root, "docker", append([]string{"compose", "-f", composeFile(root)}, args...)...)
}

func runLifecycle(action string, args []string) int {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	ctx := context.Background()
	var op []string
	switch action {
	case "start":
		op = []string{"up", "-d"}
	case "stop":
		op = []string{"stop"}
	case "restart":
		op = []string{"restart"}
	case "logs":
		op = []string{"logs", "--tail", getenv("TREST_LOG_TAIL", "200")}
		if len(args) > 0 {
			op = append(op, args...)
		}
	default:
		return 1
	}
	if err := runCompose(ctx, root, op...); err != nil {
		fmt.Fprintln(os.Stderr, action+":", err)
		return 2
	}
	return 0
}

func runBackup(args []string) int {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	dir := filepath.Join(root, ".trest", "backups")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	name := time.Now().UTC().Format("20060102-150405") + ".sql"
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		name = filepath.Base(args[0])
		if !strings.HasSuffix(name, ".sql") {
			name += ".sql"
		}
	}
	out := filepath.Join(dir, name)
	data, err := capture(context.Background(), root, "docker", "compose", "-f", composeFile(root), "exec", "-T", "postgres", "pg_dump", "-U", getenv("POSTGRES_USER", "trest"), "-d", getenv("POSTGRES_DB", "trest"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "backup:", err)
		return 2
	}
	if err := os.WriteFile(out, []byte(data), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "backup:", err)
		return 2
	}
	if info, err := os.Stat(out); err != nil || info.Size() == 0 {
		fmt.Fprintln(os.Stderr, "backup: empty dump")
		return 2
	}
	fmt.Println(out)
	return 0
}

func runRestore(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "restore: укажите .sql файл и повторите с TREST_CONFIRM_RESTORE=YES")
		return 2
	}
	if os.Getenv("TREST_CONFIRM_RESTORE") != "YES" {
		fmt.Fprintln(os.Stderr, "restore: отказано — установите TREST_CONFIRM_RESTORE=YES для подтверждения разрушительной операции")
		return 2
	}
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	b, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "restore:", err)
		return 2
	}
	if err := pipeCommand(context.Background(), root, b, "docker", "compose", "-f", composeFile(root), "exec", "-T", "postgres", "psql", "-U", getenv("POSTGRES_USER", "trest"), "-d", getenv("POSTGRES_DB", "trest")); err != nil {
		fmt.Fprintln(os.Stderr, "restore:", err)
		return 2
	}
	return 0
}

func runUpdate() int {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := runCompose(context.Background(), root, "pull"); err != nil {
		fmt.Fprintln(os.Stderr, "update:", err)
		return 2
	}
	if err := runCompose(context.Background(), root, "up", "-d"); err != nil {
		fmt.Fprintln(os.Stderr, "update:", err)
		return 2
	}
	return 0
}

func runRepair() int {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := runCompose(context.Background(), root, "ps", "--all"); err != nil {
		fmt.Fprintln(os.Stderr, "repair:", err)
		return 2
	}
	return 0
}
