package cli

import (
	"encoding/json"
	"fmt"
	"github.com/mintfary-oss/trest-sistems/internal/config"
	"github.com/mintfary-oss/trest-sistems/internal/doctor"
	"os"
)

func Run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 1
	}

	switch args[0] {
	case "version":
		fmt.Println("trestctl", config.Version)
		return 0
	case "doctor":
		return runDoctor()
	case "status":
		return runStatus()
	case "help", "-h", "--help":
		printUsage()
		return 0
	case "install":
		return runInstall(false)
	case "install-build":
		return runInstall(true)
	case "start", "stop", "restart", "logs":
		return runLifecycle(args[0], args[1:])
	case "backup":
		return runBackup(args[1:])
	case "restore":
		return runRestore(args[1:])
	case "migrate":
		return runMigrate()
	case "update":
		return runUpdate()
	case "repair":
		return runRepair()
	default:
		fmt.Fprintf(os.Stderr, "неизвестная команда: %s\n\n", args[0])
		printUsage()
		return 1
	}
}

func runDoctor() int {
	r := doctor.Run()
	for _, c := range r.Checks {
		fmt.Printf("[%s] %s: %s\n", c.Status, c.Name, c.Detail)
	}
	if r.HasFailures() {
		return 2
	}
	return 0
}

func runStatus() int {
	p := config.DefaultPaths()
	if err := config.EnsureDirs(p); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	data, _ := json.MarshalIndent(map[string]string{
		"state_dir": p.Root,
		"status":    "initialized",
	}, "", "  ")
	fmt.Println(string(data))
	return 0
}

func printUsage() {
	fmt.Println(`trestctl — единая точка управления Trest Systems

Использование:
  trestctl <команда>

Доступные команды:
  version     показать версию
  doctor      диагностика окружения
  status      состояние .trest
  install     идемпотентная установка через Docker Compose
  install-build установка с пересборкой образов
  logs        логи Docker Compose
  start       запуск сервисов
  stop        остановка сервисов
  restart     перезапуск сервисов
  backup      резервная копия PostgreSQL
  restore     восстановление PostgreSQL (требует явного подтверждения)
  update      обновление образов и сервисов
  repair      диагностика состояния сервисов
  migrate     применить неприменённые миграции PostgreSQL`)
}
