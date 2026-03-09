package datasource

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jamesprnich/oh-my-line/engine/internal/cache"
	"github.com/jamesprnich/oh-my-line/engine/internal/render"
)

// DockerResult holds docker segment data.
type DockerResult struct {
	Summary string
	DB      string
}

// FetchDocker checks docker compose status.
func FetchDocker(cwd string, types map[string]bool) *DockerResult {
	result := &DockerResult{}

	if _, err := exec.LookPath("docker"); err != nil {
		return result
	}
	if cwd == "" {
		return result
	}

	// Find compose file
	composeFile := ""
	for _, f := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
		if _, err := os.Stat(filepath.Join(cwd, f)); err == nil {
			composeFile = filepath.Join(cwd, f)
			break
		}
	}
	if composeFile == "" {
		return result
	}

	cacheDir, err := cache.Dir()
	if err != nil {
		return result
	}

	cachePath := filepath.Join(cacheDir, "statusline-docker.json")

	// Try cache first (30s TTL)
	if content, fresh := cache.ReadFile(cachePath, 30); fresh && content != "" {
		parseDockerData(content, result, types)
		return result
	}

	// Fetch fresh
	cmd := exec.Command("docker", "compose", "-f", composeFile, "ps", "--format", "json")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		// Use stale cache
		if data, err := os.ReadFile(cachePath); err == nil {
			parseDockerData(string(data), result, types)
		}
		return result
	}

	// Normalize NDJSON to array
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var containers []json.RawMessage
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		containers = append(containers, json.RawMessage(line))
	}
	arrJSON, _ := json.Marshal(containers)
	cache.WriteFile(cachePath, string(arrJSON))
	parseDockerData(string(arrJSON), result, types)

	return result
}

type containerInfo struct {
	Service string `json:"Service"`
	Name    string `json:"Name"`
	State   string `json:"State"`
	Health  string `json:"Health"`
	Image   string `json:"Image"`
}

func parseDockerData(data string, result *DockerResult, types map[string]bool) {
	var containers []containerInfo
	if json.Unmarshal([]byte(data), &containers) != nil {
		return
	}

	if types["docker"] {
		total := len(containers)
		running := 0
		unhealthy := 0
		for _, c := range containers {
			if c.State == "running" {
				running++
			}
			if c.Health == "unhealthy" {
				unhealthy++
			}
		}

		var dc, summary string
		if total == 0 {
			summary = "no containers"
			dc = render.HexFG("#888888")
		} else if unhealthy > 0 {
			summary = fmt.Sprintf("%d unhealthy", unhealthy)
			dc = render.HexFG("#ff5555")
		} else if running == total {
			summary = fmt.Sprintf("%d/%d up", running, total)
			dc = render.HexFG("#00a000")
		} else {
			summary = fmt.Sprintf("%d/%d up", running, total)
			dc = render.HexFG("#e6c800")
		}
		result.Summary = dc + summary + render.RST
	}

	if types["docker-db"] {
		dbPatterns := []string{"postgres", "mysql", "mariadb", "redis", "mongo", "memcached"}
		for _, c := range containers {
			isDB := false
			nameLower := strings.ToLower(c.Image + " " + c.Service)
			for _, p := range dbPatterns {
				if strings.Contains(nameLower, p) {
					isDB = true
					break
				}
			}
			if !isDB {
				svcLower := strings.ToLower(c.Service)
				if svcLower == "db" || svcLower == "database" {
					isDB = true
				}
			}
			if !isDB {
				continue
			}

			name := c.Service
			if name == "" {
				name = c.Name
			}
			// Shorten common names
			nl := strings.ToLower(name)
			switch {
			case strings.Contains(nl, "postgres"):
				name = "pg"
			case strings.Contains(nl, "mysql") || strings.Contains(nl, "mariadb"):
				name = "mysql"
			case strings.Contains(nl, "mongo"):
				name = "mongo"
			case strings.Contains(nl, "redis"):
				name = "redis"
			case strings.Contains(nl, "memcached"):
				name = "memcached"
			}

			pre := render.HexFG("#2e9599")
			var sc, status string
			switch c.State {
			case "running":
				if c.Health == "unhealthy" {
					sc = render.HexFG("#ff5555")
					status = "unhealthy"
				} else {
					sc = render.HexFG("#00a000")
					status = "up"
				}
			case "exited", "dead":
				sc = render.HexFG("#ff5555")
				status = "down"
			case "restarting":
				sc = render.HexFG("#e6c800")
				status = "restarting"
			default:
				sc = render.HexFG("#888888")
				status = c.State
			}
			result.DB = pre + name + ": " + sc + status + render.RST
			break
		}
	}
}
