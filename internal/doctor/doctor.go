package doctor

import (
	"fmt"
	"os/exec"
	"runtime"
)

type Status string

const (
	StatusOK   Status = "OK"
	StatusWarn Status = "WARN"
	StatusFail Status = "FAIL"
)

type Check struct {
	Name    string
	Status  Status
	Detail  string
	Fixable bool
}

type Report struct {
	Checks []Check
}

func (r Report) HasFailures() bool {
	for _, c := range r.Checks {
		if c.Status == StatusFail {
			return true
		}
	}
	return false
}

func Run() Report {
	return Report{Checks: []Check{
		checkOS(),
		checkDocker(),
		checkDockerCompose(),
	}}
}

func checkOS() Check {
	if runtime.GOOS == "linux" {
		return Check{Name: "Операционная система", Status: StatusOK, Detail: "linux/" + runtime.GOARCH}
	}
	return Check{
		Name:   "Операционная система",
		Status: StatusWarn,
		Detail: fmt.Sprintf("%s/%s — целевая платформа Linux", runtime.GOOS, runtime.GOARCH),
	}
}

func checkDocker() Check {
	path, err := exec.LookPath("docker")
	if err != nil {
		return Check{Name: "Docker", Status: StatusFail, Detail: "docker не найден в PATH"}
	}
	return Check{Name: "Docker", Status: StatusOK, Detail: path}
}

func checkDockerCompose() Check {
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err != nil {
		return Check{Name: "Docker Compose", Status: StatusFail, Detail: "docker compose недоступен"}
	}
	return Check{Name: "Docker Compose", Status: StatusOK, Detail: "docker compose v2 доступен"}
}
