//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The stack is the one defined in admin/docker-compose.yml: the admin server
// published on ports 80/443 and MySQL seeded from admin/etc/*.sql, including
// the admin/admin demo account from example-auth.sql.
var (
	baseURL  string
	adminDir string
)

func TestMain(m *testing.M) {
	baseURL = os.Getenv("ONLINECONF_URL")
	if baseURL == "" {
		baseURL = "http://localhost"
	}
	var err error
	adminDir, err = filepath.Abs("../../..") // package dir is admin/go/test/integration
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	manageStack := os.Getenv("ONLINECONF_NO_COMPOSE") == ""
	if manageStack {
		fmt.Fprintln(os.Stderr, "recreating the stack with a fresh database...")
		compose("down", "-v") // ignore failure: the stack may not exist yet
		if out, err := compose("up", "-d", "--build"); err != nil {
			fmt.Fprintf(os.Stderr, "docker compose up failed: %v\n%s\n", err, out)
			os.Exit(1)
		}
	}
	if err := waitReady(4 * time.Minute); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if manageStack {
			compose("down", "-v")
		}
		os.Exit(1)
	}

	code := m.Run()

	if manageStack {
		if out, err := compose("down", "-v"); err != nil {
			fmt.Fprintf(os.Stderr, "docker compose down failed: %v\n%s\n", err, out)
			if code == 0 {
				code = 1
			}
		}
	}
	os.Exit(code)
}

func compose(args ...string) (string, error) {
	all := append([]string{"compose", "-f", filepath.Join(adminDir, "docker-compose.yml")}, args...)
	out, err := exec.Command("docker", all...).CombinedOutput()
	return string(out), err
}

// composeStdout is like compose but returns stdout only, keeping the stderr
// noise of docker compose and mysql out of parsed query results.
func composeStdout(args ...string) (string, error) {
	all := append([]string{"compose", "-f", filepath.Join(adminDir, "docker-compose.yml")}, args...)
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("docker", all...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func waitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		req, err := http.NewRequest("GET", baseURL+"/whoami", nil)
		if err != nil {
			return err
		}
		authorize(req)
		if resp, err := httpClient.Do(req); err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("admin API did not become ready at %s within %s", baseURL, timeout)
}

// --- HTTP helpers ----------------------------------------------------------

var httpClient = &http.Client{Timeout: 30 * time.Second}

// authorize adds the credentials and the X-Requested-With header required by
// the CSRF guard (csrfProtection) on every admin endpoint.
func authorize(req *http.Request) {
	req.SetBasicAuth("admin", "admin")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
}

// adminReq performs an authenticated admin API request and returns the status
// code and response body. A nil form means a request without a body.
func adminReq(t *testing.T, method, path string, form url.Values) (int, []byte) {
	t.Helper()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(method, baseURL+path, body)
	if err != nil {
		t.Fatal(err)
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	authorize(req)
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s %s: read body: %v", method, path, err)
	}
	return resp.StatusCode, b
}

func adminReqOK(t *testing.T, method, path string, form url.Values) []byte {
	t.Helper()
	status, body := adminReq(t, method, path, form)
	if status != http.StatusOK {
		t.Fatalf("%s %s: status %d, body %s", method, path, status, body)
	}
	return body
}

// --- parameter helpers -----------------------------------------------------

// uniqPath builds a parameter path unique to this test run, so the tests can
// be re-run against a persistent stack (ONLINECONF_NO_COMPOSE=1) without
// colliding with leftovers of previous runs.
func uniqPath(name string) string {
	return fmt.Sprintf("/onlineconf/it-%s-%d", name, os.Getpid())
}

func mkParam(t *testing.T, path, mime, data, comment string) {
	t.Helper()
	adminReqOK(t, "POST", "/config"+path, url.Values{"mime": {mime}, "data": {data}, "comment": {comment}})
}

// ensureParam creates a parameter unless it already exists.
func ensureParam(t *testing.T, path, mime, data, comment string) {
	t.Helper()
	if status, _ := adminReq(t, "GET", "/config"+path, nil); status == http.StatusOK {
		return
	}
	mkParam(t, path, mime, data, comment)
}

func paramVersion(t *testing.T, path string) int {
	t.Helper()
	body := adminReqOK(t, "GET", "/config"+path, nil)
	var p struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatalf("parse %s: %v in %s", path, err, body)
	}
	return p.Version
}

func editParam(t *testing.T, path, mime, data, comment string) {
	t.Helper()
	adminReqOK(t, "POST", "/config"+path, url.Values{
		"version": {strconv.Itoa(paramVersion(t, path))},
		"mime":    {mime},
		"data":    {data},
		"comment": {comment},
	})
}

func moveParam(t *testing.T, from, to, comment string) {
	t.Helper()
	adminReqOK(t, "POST", "/config"+from, url.Values{
		"version": {strconv.Itoa(paramVersion(t, from))},
		"path":    {to},
		"comment": {comment},
	})
}

// moveParamSymlink moves a parameter leaving a symlink behind at the old path.
func moveParamSymlink(t *testing.T, from, to, comment string) {
	t.Helper()
	adminReqOK(t, "POST", "/config"+from, url.Values{
		"version": {strconv.Itoa(paramVersion(t, from))},
		"path":    {to},
		"symlink": {"1"},
		"comment": {comment},
	})
}

func deleteParam(t *testing.T, path, comment string) {
	t.Helper()
	adminReqOK(t, "DELETE", "/config"+path, url.Values{
		"version": {strconv.Itoa(paramVersion(t, path))},
		"comment": {comment},
	})
}

func setNotification(t *testing.T, path, notification string) {
	t.Helper()
	adminReqOK(t, "POST", "/config"+path, url.Values{"notification": {notification}})
}

type logEntry struct {
	Version int    `json:"version"`
	Path    string `json:"path"`
	Comment string `json:"comment"`
	Deleted bool   `json:"deleted"`
}

func getLog(t *testing.T, path string) []logEntry {
	t.Helper()
	body := adminReqOK(t, "GET", "/log"+path, nil)
	var list []logEntry
	if err := json.Unmarshal(body, &list); err != nil {
		t.Fatalf("parse log of %s: %v in %s", path, err, body)
	}
	return list
}

// --- database helpers ------------------------------------------------------

// dbQuery runs a query against the stack's MySQL (whose port is not published
// on the host) through docker compose exec and returns the result rows as
// tab-separated fields.
func dbQuery(t *testing.T, query string) [][]string {
	t.Helper()
	out, err := composeStdout("exec", "-T", "onlineconf-database",
		"mysql", "-N", "-uonlineconf", "-ponlineconf", "onlineconf", "-e", query)
	if err != nil {
		t.Fatalf("db query %q: %v", query, err)
	}
	var rows [][]string
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		rows = append(rows, strings.Split(line, "\t"))
	}
	return rows
}

// dbInt runs a query expected to return a single integer cell.
func dbInt(t *testing.T, query string) int {
	t.Helper()
	rows := dbQuery(t, query)
	if len(rows) != 1 || len(rows[0]) != 1 {
		t.Fatalf("db query %q: expected a single cell, got %v", query, rows)
	}
	n, err := strconv.Atoi(rows[0][0])
	if err != nil {
		t.Fatalf("db query %q: %v", query, err)
	}
	return n
}
