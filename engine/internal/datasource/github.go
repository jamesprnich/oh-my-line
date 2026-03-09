package datasource

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jamesprnich/oh-my-line/engine/internal/cache"
	"github.com/jamesprnich/oh-my-line/engine/internal/debug"
	"github.com/jamesprnich/oh-my-line/engine/internal/render"
)

// GitHubResult holds all GitHub segment data.
type GitHubResult struct {
	PR         string
	Checks     string
	Reviews    string
	Actions    string
	Notifs     string
	Issues     string
	PRCount    string
	PRComments string
	Stars      string
}

// FetchGitHub fetches GitHub data for all enabled segment types.
func FetchGitHub(cwd, branch string, types map[string]bool) *GitHubResult {
	result := &GitHubResult{}

	// Require gh CLI
	if _, err := exec.LookPath("gh"); err != nil {
		return result
	}
	if cwd == "" || branch == "" {
		return result
	}

	cacheDir, err := cache.Dir()
	if err != nil {
		return result
	}

	now := time.Now().Unix()
	stale := ""

	// Group 1: PR data
	if types["gh-pr"] || types["gh-checks"] || types["gh-reviews"] || types["gh-pr-comments"] {
		prCache := filepath.Join(cacheDir, "gh-pr.json")
		data := bgFetch(prCache, 60, cwd, "pr", "view", "--json",
			"number,state,isDraft,reviewDecision,reviewRequests,reviews,comments,statusCheckRollup")

		if data != "" {
			var pr struct {
				Number         int    `json:"number"`
				State          string `json:"state"`
				IsDraft        bool   `json:"isDraft"`
				ReviewDecision string `json:"reviewDecision"`
				ReviewRequests []struct{} `json:"reviewRequests"`
				Reviews        []struct {
					State string `json:"state"`
				} `json:"reviews"`
				Comments           []struct{} `json:"comments"`
				StatusCheckRollup []struct {
					Status     string `json:"status"`
					State      string `json:"state"`
					Conclusion string `json:"conclusion"`
				} `json:"statusCheckRollup"`
			}
			if json.Unmarshal([]byte(data), &pr) == nil {
				// gh-pr
				if types["gh-pr"] {
					if pr.Number == 0 {
						result.PR = render.DIM + "no PR" + render.RST
					} else {
						var sc, stateText string
						if pr.IsDraft {
							sc = render.HexFG("#888888")
							stateText = "draft"
						} else {
							switch pr.State {
							case "OPEN":
								sc = render.HexFG("#00a000")
								stateText = "open"
							case "MERGED":
								sc = render.HexFG("#8957e5")
								stateText = "merged"
							case "CLOSED":
								sc = render.HexFG("#ff5555")
								stateText = "closed"
							}
						}
						result.PR = fmt.Sprintf("PR %s#%d%s %s%s%s%s",
							render.HexFG("#2e9599"), pr.Number, render.RST,
							sc, stateText, render.RST, stale)
					}
				}

				// gh-checks
				if types["gh-checks"] {
					total := len(pr.StatusCheckRollup)
					if total == 0 {
						if pr.Number > 0 {
							result.Checks = render.DIM + "CI —" + render.RST
						}
					} else {
						pass, fail, pending := 0, 0, 0
						for _, c := range pr.StatusCheckRollup {
							if c.Conclusion == "SUCCESS" || c.State == "SUCCESS" {
								pass++
							} else if c.Conclusion == "FAILURE" || c.Conclusion == "STARTUP_FAILURE" || c.State == "FAILURE" || c.State == "ERROR" {
								fail++
							} else if c.Status == "IN_PROGRESS" || c.Status == "QUEUED" || c.Status == "WAITING" || c.State == "PENDING" || c.State == "EXPECTED" {
								pending++
							}
						}
						var cc, statusText string
						if fail > 0 {
							cc = render.HexFG("#ff5555")
							statusText = fmt.Sprintf("fail %d", fail)
						} else if pending > 0 {
							cc = render.HexFG("#e6c800")
							statusText = fmt.Sprintf("%d/%d", pass, total)
						} else {
							cc = render.HexFG("#00a000")
							statusText = "pass"
						}
						result.Checks = "CI " + cc + statusText + render.RST
					}
				}

				// gh-reviews
				if types["gh-reviews"] && pr.Number > 0 {
					reviewCount := 0
					for _, r := range pr.Reviews {
						if r.State == "APPROVED" || r.State == "CHANGES_REQUESTED" {
							reviewCount++
						}
					}
					requested := len(pr.ReviewRequests)

					switch pr.ReviewDecision {
					case "APPROVED":
						result.Reviews = render.HexFG("#00a000") + "approved" + render.RST
					case "CHANGES_REQUESTED":
						result.Reviews = render.HexFG("#ff5555") + "changes requested" + render.RST
					case "REVIEW_REQUIRED":
						totalNeeded := reviewCount + requested
						if totalNeeded > 0 {
							result.Reviews = render.HexFG("#e6c800") + fmt.Sprintf("%d/%d reviews", reviewCount, totalNeeded) + render.RST
						} else {
							result.Reviews = render.HexFG("#e6c800") + "review required" + render.RST
						}
					default:
						if reviewCount > 0 {
							result.Reviews = render.HexFG("#00a000") + fmt.Sprintf("%d reviews", reviewCount) + render.RST
						}
					}
				}

				// gh-pr-comments
				if types["gh-pr-comments"] && len(pr.Comments) > 0 {
					result.PRComments = render.HexFG("#dcdcdc") + fmt.Sprintf("%d", len(pr.Comments)) + render.RST +
						" " + render.DIM + "comments" + render.RST
				}
			}
		}
	}

	// Group 2: Actions
	if types["gh-actions"] {
		actionsCache := filepath.Join(cacheDir, "gh-actions.json")
		data := bgFetch(actionsCache, 60, cwd, "run", "list", "--branch", branch, "--limit", "1",
			"--json", "name,status,conclusion,createdAt")

		if data != "" && data != "[]" {
			var runs []struct {
				Name       string `json:"name"`
				Status     string `json:"status"`
				Conclusion string `json:"conclusion"`
				CreatedAt  string `json:"createdAt"`
			}
			if json.Unmarshal([]byte(data), &runs) == nil && len(runs) > 0 {
				run := runs[0]
				name := run.Name
				if len(name) > 16 {
					name = name[:14] + ".."
				}
				status := run.Status
				if status == "completed" {
					status = run.Conclusion
				}

				var ac string
				switch status {
				case "success":
					ac = render.HexFG("#00a000")
				case "failure", "startup_failure":
					ac = render.HexFG("#ff5555")
				case "in_progress", "queued", "waiting":
					ac = render.HexFG("#e6c800")
				default:
					ac = render.HexFG("#888888")
				}

				out := ac + name + render.RST + " " + render.DIM + status + render.RST
				if run.CreatedAt != "" {
					if ep := isoToEpoch(run.CreatedAt); ep > 0 {
						age := int(now - ep)
						var ageStr string
						switch {
						case age >= 86400:
							ageStr = fmt.Sprintf("%dd", age/86400)
						case age >= 3600:
							ageStr = fmt.Sprintf("%dh", age/3600)
						case age >= 60:
							ageStr = fmt.Sprintf("%dm", age/60)
						}
						if ageStr != "" {
							out += " " + render.DIM + ageStr + render.RST
						}
					}
				}
				result.Actions = out
			}
		}
		if result.Actions == "" {
			result.Actions = render.DIM + "actions —" + render.RST
		}
	}

	// Group 3: Notifications
	if types["gh-notifs"] {
		notifsCache := filepath.Join(cacheDir, "gh-notifs.txt")
		data := bgFetch(notifsCache, 120, cwd, "api", "notifications", "--jq", "length")
		data = strings.TrimSpace(data)
		if data != "" && data != "0" {
			result.Notifs = render.HexFG("#dcdcdc") + data + render.RST + " " + render.DIM + "notifs" + render.RST
		}
	}

	// Group 4: Repo stats
	if types["gh-stars"] {
		starsCache := filepath.Join(cacheDir, "gh-stars.txt")
		data := bgFetch(starsCache, 300, cwd, "repo", "view", "--json", "stargazerCount", "--jq", ".stargazerCount")
		data = strings.TrimSpace(data)
		if data != "" && data != "0" {
			n, _ := strconv.Atoi(data)
			var display string
			switch {
			case n >= 1000000:
				display = fmt.Sprintf("%.1fm", float64(n)/1000000)
			case n >= 1000:
				display = fmt.Sprintf("%.1fk", float64(n)/1000)
			default:
				display = data
			}
			result.Stars = render.HexFG("#e6c800") + "★ " + display + render.RST
		}
	}

	if types["gh-issues"] {
		issuesCache := filepath.Join(cacheDir, "gh-issues.txt")
		data := bgFetch(issuesCache, 300, cwd, "issue", "list", "--state", "open", "--limit", "999",
			"--json", "number", "--jq", "length")
		data = strings.TrimSpace(data)
		if data != "" && data != "0" {
			result.Issues = render.DIM + "issues: " + render.RST + render.HexFG("#dcdcdc") + data + render.RST
		}
	}

	if types["gh-pr-count"] {
		prsCache := filepath.Join(cacheDir, "gh-prs.txt")
		data := bgFetch(prsCache, 300, cwd, "pr", "list", "--state", "open", "--limit", "999",
			"--json", "number", "--jq", "length")
		data = strings.TrimSpace(data)
		if data != "" && data != "0" {
			result.PRCount = render.DIM + "PRs: " + render.RST + render.HexFG("#dcdcdc") + data + render.RST
		}
	}

	return result
}

// bgFetch does a non-blocking gh fetch with stale cache serving.
// Uses a detached subprocess so the fetch survives parent process exit.
func bgFetch(cacheFile string, ttl int, cwd string, args ...string) string {
	tmpFile := cacheFile + ".tmp"
	pidFile := cacheFile + ".pid"

	// Check if background fetch completed
	if _, err := os.Stat(tmpFile); err == nil {
		if _, err := os.Stat(pidFile); err != nil {
			if data, err := os.ReadFile(tmpFile); err == nil && len(data) > 2 {
				os.Rename(tmpFile, cacheFile)
				debug.Log("gh", "bg-fetch promoted tmp→cache size=%d", len(data))
			} else {
				os.Remove(tmpFile)
			}
		}
	}

	// Serve fresh cache
	if content, fresh := cache.ReadFile(cacheFile, ttl); fresh && content != "" {
		return content
	}

	// Serve stale cache
	staleData := ""
	if data, err := os.ReadFile(cacheFile); err == nil {
		staleData = string(data)
	}

	// Launch background refresh as a detached subprocess.
	// A goroutine would be killed when main() returns.
	// Writes to a staging file first, then atomically renames to .tmp
	// to prevent readers from seeing partially-written data.
	if shouldLaunchBG(pidFile) {
		stageFile := cacheFile + ".stage"
		ghArgs := append([]string{"gh"}, args...)
		script := fmt.Sprintf("%s > %s && mv -f %s %s; rm -f %s",
			shellJoinArgs(ghArgs), shellQuote(stageFile),
			shellQuote(stageFile), shellQuote(tmpFile), shellQuote(pidFile))
		cmd := exec.Command("sh", "-c", script)
		cmd.Dir = cwd
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Start(); err != nil {
			debug.Log("gh", "bg-fetch start err=%v", err)
		} else {
			os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0600)
			debug.Log("gh", "bg-fetch launched pid=%d args=%v", cmd.Process.Pid, args)
		}
	}

	return staleData
}

func shouldLaunchBG(pidFile string) bool {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return true
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return true
	}
	// Verify the process is still alive AND is one of ours by checking
	// /proc/<pid>/comm. A recycled PID would have a different command name.
	if comm, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid)); err == nil {
		name := strings.TrimSpace(string(comm))
		if name != "sh" && name != "curl" && name != "gh" {
			// PID was recycled to an unrelated process
			return true
		}
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return true
	}
	return proc.Signal(nil) != nil
}
