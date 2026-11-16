//go:build integration

package integration

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ugorji/go/codec"
)

// The value encryption suite. The compose
// stack runs with the demo keyring (key id "demo"); the rotation test swaps
// the key file inside the admin container and restarts it.

type paramInfo struct {
	Path      string  `json:"path"`
	Mime      string  `json:"mime"`
	Data      *string `json:"data"`
	Encrypted bool    `json:"encrypted"`
	CanReveal bool    `json:"can_reveal"`
	Version   int     `json:"version"`
	RW        *bool   `json:"rw"`
}

// reqAs is adminReq with explicit credentials.
func reqAs(t *testing.T, user, password, method, path string, form url.Values) (int, []byte) {
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
	req.SetBasicAuth(user, password)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
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

func getParamInfoAs(t *testing.T, user, password, path, query string) (int, paramInfo) {
	t.Helper()
	status, body := reqAs(t, user, password, "GET", "/config"+path+query, nil)
	var p paramInfo
	if status == http.StatusOK {
		if err := json.Unmarshal(body, &p); err != nil {
			t.Fatalf("parse %s: %v in %s", path, err, body)
		}
	}
	return status, p
}

func getParamInfo(t *testing.T, path, query string) paramInfo {
	t.Helper()
	status, p := getParamInfoAs(t, "admin", "admin", path, query)
	if status != http.StatusOK {
		t.Fatalf("GET /config%s%s: status %d", path, query, status)
	}
	return p
}

func dataOf(p paramInfo) string {
	if p.Data == nil {
		return "<null>"
	}
	return *p.Data
}

func mkEncrypted(t *testing.T, path, mime, data, comment string) {
	t.Helper()
	adminReqOK(t, "POST", "/config"+path, url.Values{"mime": {mime}, "data": {data}, "encrypted": {"1"}, "comment": {comment}})
}

func setParam(t *testing.T, path string, form url.Values) (int, []byte) {
	t.Helper()
	form.Set("version", strconv.Itoa(paramVersion(t, path)))
	if _, ok := form["comment"]; !ok {
		form.Set("comment", "set")
	}
	return adminReq(t, "POST", "/config"+path, form)
}

func expectError(t *testing.T, status int, body []byte, wantStatus int, wantCode string) {
	t.Helper()
	if status != wantStatus || !strings.Contains(string(body), `"error":"`+wantCode+`"`) {
		t.Errorf("status %d, body %s; want %d %s", status, body, wantStatus, wantCode)
	}
}

func sqlQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// logEntryID finds the history entry of a path by its comment.
func logEntryID(t *testing.T, path, comment string) int {
	t.Helper()
	for _, e := range getLog(t, path) {
		if e.Comment == comment {
			return e.ID
		}
	}
	t.Fatalf("no entry of %s with the comment %q", path, comment)
	return 0
}

// restore rolls a live parameter back to a history entry, server-side.
func restore(t *testing.T, path string, entryID int) (int, []byte) {
	t.Helper()
	return setParam(t, path, url.Values{"restore": {strconv.Itoa(entryID)}, "comment": {"rollback"}})
}

// --- tests --------------------------------------------------------------------

func TestEncryptedCRUD(t *testing.T) {
	p := uniqPath("enc")
	mkEncrypted(t, p, "text/plain", "hunter2", "mk secret")

	// stored as an envelope under the demo key, marked, in the tree and the log
	if got := dbInt(t, "SELECT Encrypted AND Value LIKE 'oc1:%' FROM my_config_tree WHERE Path = "+sqlQuote(p)); got != 1 {
		t.Errorf("tree row is not an envelope under the demo key")
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_tree_log WHERE Path = "+sqlQuote(p)+" AND Encrypted AND Value LIKE 'oc1:%'"); got != 1 {
		t.Errorf("log row is not an envelope under the demo key")
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_tree WHERE Path = "+sqlQuote(p)+" AND Value LIKE '%hunter2%'"); got != 0 {
		t.Errorf("plaintext stored in the tree")
	}

	// listings mask; a reveal returns plaintext and is audited
	info := getParamInfo(t, p, "")
	if info.Data != nil || !info.Encrypted || !info.CanReveal {
		t.Errorf("masked GET: data=%s encrypted=%v can_reveal=%v", dataOf(info), info.Encrypted, info.CanReveal)
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_audit WHERE Path = "+sqlQuote(p)); got != 0 {
		t.Errorf("a masked GET was audited as a reveal")
	}
	info = getParamInfo(t, p, "?reveal=1")
	if dataOf(info) != "hunter2" || !info.Encrypted {
		t.Errorf("reveal: data=%s encrypted=%v", dataOf(info), info.Encrypted)
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_audit WHERE Path = "+sqlQuote(p)+" AND User = 'admin' AND Kind = 'reveal_view'"); got != 1 {
		t.Errorf("reveal audit rows = %d, want 1", got)
	}
	if status, body := adminReq(t, "GET", "/config"+p+"?reveal=all", nil); status != http.StatusBadRequest {
		t.Errorf("bogus reveal kind: status %d, body %s", status, body)
	}
	// children of a revealed parameter stay masked
	mkEncrypted(t, p+"/child", "text/plain", "child-secret", "mk child")
	var withChildren struct {
		Data     *string     `json:"data"`
		Children []paramInfo `json:"children"`
	}
	if err := json.Unmarshal(adminReqOK(t, "GET", "/config"+p+"?reveal=1&depth=children", nil), &withChildren); err != nil {
		t.Fatal(err)
	}
	if withChildren.Data == nil || len(withChildren.Children) != 1 || withChildren.Children[0].Data != nil {
		t.Errorf("reveal with depth: parent data=%v children=%+v", withChildren.Data, withChildren.Children)
	}

	// a set without the field is refused: a stale client would wipe the secret
	status, body := setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {""}})
	expectError(t, status, body, http.StatusBadRequest, "EncryptedFieldRequired")
	if dataOf(getParamInfo(t, p, "?reveal=1")) != "hunter2" {
		t.Errorf("the refused write changed the value")
	}
	// re-encrypting a new value
	if status, body := setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"hunter3"}, "encrypted": {"1"}}); status != http.StatusOK {
		t.Fatalf("set encrypted: status %d, body %s", status, body)
	}
	if dataOf(getParamInfo(t, p, "?reveal=1")) != "hunter3" {
		t.Errorf("new value not stored")
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_audit WHERE Path = "+sqlQuote(p)+" AND Kind = 'reveal_view'"); got != 4 {
		t.Errorf("reveal audit rows after the edits = %d, want 4 (view, view with children, two edits)", got)
	}
	// an invalid encrypted value is refused with the bare code: nothing of
	// the secret in the message
	status, body = setParam(t, p, url.Values{"mime": {"application/json"}, "data": {`{"zq":`}, "encrypted": {"1"}})
	expectError(t, status, body, http.StatusBadRequest, "InvalidValue")
	if strings.Contains(string(body), "zq") {
		t.Errorf("the invalid value's text is in the response: %s", body)
	}
	// the type cannot become ineligible while encrypted
	status, body = setParam(t, p, url.Values{"mime": {"application/x-symlink"}, "data": {p + "/child"}, "encrypted": {"1"}})
	expectError(t, status, body, http.StatusBadRequest, "EncryptionNotAllowed")
	// declassification stores plaintext; history stays sealed
	if status, body := setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"public"}, "encrypted": {"0"}}); status != http.StatusOK {
		t.Fatalf("declassify: status %d, body %s", status, body)
	}
	info = getParamInfo(t, p, "")
	if dataOf(info) != "public" || info.Encrypted || info.CanReveal {
		t.Errorf("declassified GET: data=%s encrypted=%v can_reveal=%v", dataOf(info), info.Encrypted, info.CanReveal)
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_tree_log WHERE Path = "+sqlQuote(p)+" AND NOT Encrypted"); got != 1 {
		t.Errorf("plaintext log rows after declassification = %d, want 1 (the new version only)", got)
	}

	// create with the field but an ineligible type, and a resurrect that resets the marker
	q := uniqPath("enc-null")
	status, body = adminReq(t, "POST", "/config"+q, url.Values{"mime": {"application/x-null"}, "encrypted": {"1"}, "comment": {"mk"}})
	expectError(t, status, body, http.StatusBadRequest, "EncryptionNotAllowed")
	mkEncrypted(t, q, "text/plain", "zombie", "mk")
	deleteParam(t, q, "rm")
	mkParam(t, q, "text/plain", "plain again", "resurrect")
	if info := getParamInfo(t, q, ""); info.Encrypted || dataOf(info) != "plain again" {
		t.Errorf("resurrected parameter inherited the marker: %+v", info)
	}
}

// a case branch is never encrypted: a branch that needs a secret is a symlink
// to an encrypted parameter
func TestCaseBranchesNotEncrypted(t *testing.T) {
	p := uniqPath("enc-case")
	plain := `[{"mime":"text/plain","value":"dev"}]`
	// a stray "encrypted" key on a branch is an unknown key, ignored like any
	// other: the branch is stored as the plaintext it is, the case carries
	// no marker
	marked := `[{"group":"prod","mime":"text/plain","encrypted":true,"value":"prod-plain"},{"mime":"text/plain","value":"dev"}]`
	mkParam(t, p, "application/x-case", marked, "mk case")
	if info := getParamInfo(t, p, ""); info.Encrypted || info.CanReveal || dataOf(info) != marked {
		t.Errorf("case with a stray marker on a branch: %+v data=%s", info, dataOf(info))
	}
	// the node-level marker on a case is refused
	status, body := setParam(t, p, url.Values{"mime": {"application/x-case"}, "data": {plain}, "encrypted": {"1"}})
	expectError(t, status, body, http.StatusBadRequest, "EncryptionNotAllowed")
	if dataOf(getParamInfo(t, p, "")) != marked {
		t.Errorf("the refused write changed the case")
	}

	// a symlink branch to an encrypted parameter: the case is served plain
	// without can_reveal, delivered with the plaintext, and keeps its target
	// from being deleted
	allowEphemeralClient(t)
	name := fmt.Sprintf("it-case-%d", os.Getpid())
	secret := "/gopher/" + name + "-secret"
	mkEncrypted(t, secret, "text/plain", "case-secret", "mk secret")
	c := "/gopher/" + name
	mkParam(t, c, "application/x-case", `[{"mime":"application/x-symlink","value":"`+secret+`"}]`, "mk case")
	info := getParamInfo(t, c, "")
	if info.Encrypted || info.CanReveal || info.Data == nil || !strings.Contains(*info.Data, secret) {
		t.Errorf("case with a symlink branch is not served plain: %+v data=%s", info, dataOf(info))
	}
	if got := dbInt(t, "SELECT Encrypted FROM my_config_tree WHERE Path = "+sqlQuote(c)); got != 0 {
		t.Errorf("a case carries the marker")
	}
	wire := "/onlineconf/module/TREE/gopher/" + name
	expectDelivered(t, awaitDelivered(t, wire), wire, "text/plain", "case-secret")
	status, body = adminReq(t, "DELETE", "/config"+secret, url.Values{"version": {strconv.Itoa(paramVersion(t, secret))}, "comment": {"rm"}})
	expectError(t, status, body, http.StatusBadRequest, "DeletedSymlink")
	if !strings.Contains(string(body), c) {
		t.Errorf("the error does not name the case: %s", body)
	}
	deleteParam(t, c, "rm case")
	deleteParam(t, secret, "rm secret")
}

func TestRevealPolicy(t *testing.T) {
	p := uniqPath("enc-policy")
	mkEncrypted(t, p, "text/plain", "policy-secret", "mk")

	// a read-only user on the parameter (meerkat has no access under /onlineconf otherwise)
	group := fmt.Sprintf("it-ro-%d", os.Getpid())
	adminReqOK(t, "POST", "/group/"+group, nil)
	adminReqOK(t, "POST", "/group/"+group+"/meerkat", nil)
	adminReqOK(t, "POST", "/access"+p, url.Values{"group": {group}, "rw": {"false"}})

	status, info := getParamInfoAs(t, "meerkat", "meerkat", p, "")
	if status != http.StatusOK || info.RW == nil || *info.RW || info.CanReveal || info.Data != nil {
		t.Errorf("read-only user, default policy: status=%d rw=%v can_reveal=%v data=%s", status, info.RW, info.CanReveal, dataOf(info))
	}
	status, body := reqAs(t, "meerkat", "meerkat", "GET", "/config"+p+"?reveal=1", nil)
	expectError(t, status, body, http.StatusForbidden, "RevealDenied")
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_audit WHERE Path = "+sqlQuote(p)); got != 0 {
		t.Errorf("a refused reveal was audited")
	}

	// the policy parameter relaxes it to read access
	policy := "/onlineconf/value-encryption/reveal"
	mkParam(t, policy, "text/plain", "ro", "relax the reveal policy")
	t.Cleanup(func() { deleteParam(t, policy, "restore the reveal policy") })

	// the UI learns who may reveal per row (can_reveal), not from /ui-config
	var ui struct {
		Encryption map[string]json.RawMessage `json:"encryption"`
	}
	body = adminReqOK(t, "GET", "/ui-config", nil)
	if err := json.Unmarshal(body, &ui); err != nil {
		t.Fatal(err)
	}
	if string(ui.Encryption["enabled"]) != "true" || ui.Encryption["reveal"] != nil {
		t.Errorf("/ui-config encryption = %s, want enabled without a policy", body)
	}
	status, info = getParamInfoAs(t, "meerkat", "meerkat", p, "")
	if status != http.StatusOK || !info.CanReveal {
		t.Errorf("read-only user, ro policy: status=%d can_reveal=%v", status, info.CanReveal)
	}
	status, info = getParamInfoAs(t, "meerkat", "meerkat", p, "?reveal=1")
	if status != http.StatusOK || dataOf(info) != "policy-secret" {
		t.Errorf("read-only user reveal under the ro policy: status=%d data=%s", status, dataOf(info))
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_audit WHERE Path = "+sqlQuote(p)+" AND User = 'meerkat'"); got != 1 {
		t.Errorf("meerkat's reveal audit rows = %d, want 1", got)
	}
}

func TestRevealFeed(t *testing.T) {
	bot := fmt.Sprintf("it-rbot-%d", os.Getpid())
	const password = "it-rbot-password"
	hash := sha256.Sum256([]byte(password))
	ensureParam(t, "/onlineconf/botapi", "application/x-null", "", "botapi root")
	ensureParam(t, "/onlineconf/botapi/bot", "application/x-null", "", "botapi accounts")
	mkParam(t, "/onlineconf/botapi/bot/"+bot, "text/plain", hex.EncodeToString(hash[:]), "bot account")
	mkParam(t, "/onlineconf/botapi/bot/"+bot+"/scopes", "application/x-list", "notifications,audit", "bot scopes")
	noScope := fmt.Sprintf("it-nbot-%d", os.Getpid())
	mkParam(t, "/onlineconf/botapi/bot/"+noScope, "text/plain", hex.EncodeToString(hash[:]), "bot account")
	mkParam(t, "/onlineconf/botapi/bot/"+noScope+"/scopes", "application/x-list", "notifications", "bot scopes")

	feed := func(bot string, lastID int) (int, map[string]json.RawMessage) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/botapi/audit/?lastID=%d&limit=100", baseURL, lastID), nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(bot, password)
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var r map[string]json.RawMessage
		if resp.StatusCode == http.StatusOK {
			if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
				t.Fatal(err)
			}
		}
		return resp.StatusCode, r
	}
	deadline := time.Now().Add(60 * time.Second)
	var status int
	var baseline map[string]json.RawMessage
	for {
		if status, baseline = feed(bot, 0); status == http.StatusOK {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("botapi credentials did not become active, last status %d", status)
		}
		time.Sleep(time.Second)
	}
	var lastID int
	if err := json.Unmarshal(baseline["lastID"], &lastID); err != nil {
		t.Fatal(err)
	}
	if status, _ := feed(noScope, 0); status != http.StatusForbidden {
		t.Errorf("audit feed without the audit scope: status %d, want 403", status)
	}

	p := uniqPath("enc-feed")
	mkEncrypted(t, p, "text/plain", "feed-secret", "mk")
	setNotification(t, p, "no-value") // the feed follows the notification opt-in
	const countQuery = "SELECT COUNT(*) FROM my_change_notification"
	before := dbInt(t, countQuery)
	getParamInfo(t, p, "?reveal=1")

	status, r := feed(bot, lastID)
	if status != http.StatusOK {
		t.Fatalf("reveal feed: status %d", status)
	}
	var reveals []struct {
		Path  string            `json:"path"`
		User  string            `json:"user"`
		Kind  string            `json:"kind"`
		Users map[string]string `json:"users"`
	}
	if err := json.Unmarshal(r["events"], &reveals); err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, rv := range reveals {
		if rv.Path == p {
			found = true
			if rv.User != "admin" || rv.Kind != "reveal_view" || rv.Users["admin"] != "rw" {
				t.Errorf("reveal event = %+v", rv)
			}
		}
	}
	if !found {
		t.Errorf("reveal of %s missing from the feed: %s", p, r["events"])
	}
	// legacy notifyDB message, without the value
	if got := dbInt(t, countQuery); got != before+1 {
		t.Errorf("notifyDB messages %d -> %d, want +1 for the reveal", before, got)
	}
	last := dbQuery(t, "SELECT Message FROM my_change_notification ORDER BY ID DESC LIMIT 1")
	if len(last) != 1 || !strings.Contains(last[0][0], "🔓 admin revealed "+p) || strings.Contains(last[0][0], "feed-secret") {
		t.Errorf("reveal notifyDB message = %v", last)
	}
}

// clientConfig mirrors the CBOR document of /client/config (the serializer's
// map keys are "modules"/"nodes" and the Go field names of the nodes).
type clientConfig struct {
	Modules []string `json:"modules"`
	Nodes   []struct {
		Path        string
		ContentType string
		Value       *string
	} `json:"nodes"`
}

func fetchClientConfig(t *testing.T) (int, *clientConfig) {
	t.Helper()
	req, err := http.NewRequest("GET", baseURL+"/client/config", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("gopher", "gopher")
	req.Header.Set("X-OnlineConf-Client-Version", "TEST")
	req.Header.Set("X-OnlineConf-Client-MTime", "")
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	var data clientConfig
	if err := codec.NewDecoder(resp.Body, &codec.CborHandle{}).Decode(&data); err != nil {
		t.Fatalf("decode /client/config: %v", err)
	}
	return resp.StatusCode, &data
}

// allowEphemeralClient makes every network ephemeral for the test's duration:
// the test client is identified by its IP, and no reverse DNS is attempted on
// the docker bridge address.
func allowEphemeralClient(t *testing.T) {
	t.Helper()
	ephemeral := getParamInfo(t, "/onlineconf/ephemeral-ip", "")
	if status, body := setParam(t, "/onlineconf/ephemeral-ip", url.Values{"mime": {"text/plain"}, "data": {"0.0.0.0/0"}}); status != http.StatusOK {
		t.Fatalf("ephemeral-ip: status %d, body %s", status, body)
	}
	t.Cleanup(func() {
		setParam(t, "/onlineconf/ephemeral-ip", url.Values{"mime": {ephemeral.Mime}, "data": {dataOf(ephemeral)}})
	})
}

type deliveredNode struct {
	ContentType string
	Value       *string
}

// awaitDelivered polls /client/config until every wanted path is delivered
// and returns the delivered nodes by path.
func awaitDelivered(t *testing.T, paths ...string) map[string]deliveredNode {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		status, cfg := fetchClientConfig(t)
		var delivered []string
		if status == http.StatusOK {
			found := map[string]deliveredNode{}
			for _, n := range cfg.Nodes {
				delivered = append(delivered, n.Path)
				found[n.Path] = deliveredNode{n.ContentType, n.Value}
			}
			complete := true
			for _, p := range paths {
				if _, ok := found[p]; !ok {
					complete = false
				}
			}
			if complete {
				return found
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("%v were not all delivered (last status %d); delivered: %v", paths, status, delivered)
		}
		time.Sleep(2 * time.Second)
	}
}

func expectDelivered(t *testing.T, nodes map[string]deliveredNode, path, contentType, value string) {
	t.Helper()
	if n := nodes[path]; n.ContentType != contentType || n.Value == nil || *n.Value != value {
		t.Errorf("delivered %s: type=%s value=%v; want %s %q", path, n.ContentType, n.Value, contentType, value)
	}
}

func TestEncryptedDelivery(t *testing.T) {
	allowEphemeralClient(t)
	name := fmt.Sprintf("it-delivery-%d", os.Getpid())
	mkEncrypted(t, "/gopher/"+name, "application/json", `{"token": "delivered-secret"}`, "mk")
	mkParam(t, "/gopher/"+name+"-case", "application/x-case",
		`[{"mime":"application/x-symlink","value":"/gopher/`+name+`"}]`, "mk")

	wire := "/onlineconf/module/TREE/gopher/" + name
	nodes := awaitDelivered(t, wire, wire+"-case")
	expectDelivered(t, nodes, wire, "application/json", `{"token": "delivered-secret"}`)
	expectDelivered(t, nodes, wire+"-case", "application/json", `{"token": "delivered-secret"}`)
}

func TestSearchSkipsEncrypted(t *testing.T) {
	p := uniqPath("enc-search")
	token := fmt.Sprintf("needle%d", os.Getpid())
	mkEncrypted(t, p, "text/plain", token, "mk")
	paths := func(term string) []string {
		var list []paramInfo
		if err := json.Unmarshal(adminReqOK(t, "GET", "/search?term="+term, nil), &list); err != nil {
			t.Fatal(err)
		}
		var ps []string
		for _, l := range list {
			ps = append(ps, l.Path)
		}
		return ps
	}
	if got := paths(token); len(got) != 0 {
		t.Errorf("search found the encrypted value: %v", got)
	}
	// nor does the stored envelope match: neither its prefix nor its key id
	stored := dbQuery(t, "SELECT Value FROM my_config_tree WHERE Path = "+sqlQuote(p))[0][0]
	for _, term := range []string{"oc1", strings.SplitN(stored, ":", 3)[1]} {
		for _, found := range paths(term) {
			if found == p {
				t.Errorf("search for %q matched the envelope", term)
			}
		}
	}
	if status, body := setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {token}, "encrypted": {"0"}}); status != http.StatusOK {
		t.Fatalf("declassify: status %d, body %s", status, body)
	}
	if got := paths(token); len(got) != 1 || got[0] != p {
		t.Errorf("search after declassification = %v, want [%s]", got, p)
	}
}

func TestNotificationsRedacted(t *testing.T) {
	n := uniqPath("enc-ntf")
	mkParam(t, n, "application/x-null", "", "mk root")
	setNotification(t, n, "with-value")
	const countQuery = "SELECT COUNT(*) FROM my_change_notification"
	before := dbInt(t, countQuery)
	mkEncrypted(t, n+"/secret", "application/json", `{"k": "notified-secret"}`, "mk secret")
	mkParam(t, n+"/case", "application/x-case",
		`[{"group":"g","mime":"application/x-symlink","value":"`+n+`/secret"},{"mime":"text/plain","value":"shown"}]`, "mk case")
	if got := dbInt(t, countQuery); got != before+2 {
		t.Fatalf("notifyDB messages %d -> %d, want +2", before, got)
	}
	// the latest message is the case's, the one before it the secret's
	msgs := dbQuery(t, "SELECT Message FROM my_change_notification ORDER BY ID DESC LIMIT 2")
	if m := msgs[1][0]; strings.Contains(m, "notified-secret") || !strings.Contains(m, "🔒") {
		t.Errorf("notifyDB message of the secret: %s", m)
	}
	if m := msgs[0][0]; !strings.Contains(m, "shown") || !strings.Contains(m, n+"/secret") || strings.Contains(m, "🔒") {
		t.Errorf("notifyDB message of the plain case: %s", m)
	}

	// the botapi feed serves null values with the marker
	rows := dbQuery(t, "SELECT ID FROM my_config_tree_log WHERE Path = "+sqlQuote(n+"/secret"))
	id, _ := strconv.Atoi(rows[0][0])
	bot := fmt.Sprintf("it-rbot-%d", os.Getpid()) // created by TestRevealFeed
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/botapi/notification/?lastID=%d&limit=100", baseURL, id-1), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth(bot, "it-rbot-password")
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("botapi notifications: status %d", resp.StatusCode)
	}
	var r struct {
		Notifications []struct {
			Path      string  `json:"path"`
			Encrypted bool    `json:"encrypted"`
			Value     *string `json:"value"`
		} `json:"notifications"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatal(err)
	}
	for _, ntf := range r.Notifications {
		switch ntf.Path {
		case n + "/secret":
			if !ntf.Encrypted || ntf.Value != nil {
				t.Errorf("botapi secret notification: %+v", ntf)
			}
		case n + "/case":
			if ntf.Encrypted || ntf.Value == nil || !strings.Contains(*ntf.Value, "shown") {
				t.Errorf("botapi case notification: %+v", ntf)
			}
		}
	}
}

func TestHistorySealedOnFlip(t *testing.T) {
	p := uniqPath("enc-seal")
	mkParam(t, p, "text/plain", "v1-plain", "mk")
	setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"v2-plain"}})
	if status, body := setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"v3-secret"}, "encrypted": {"1"}}); status != http.StatusOK {
		t.Fatalf("encrypt: status %d, body %s", status, body)
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_tree_log WHERE Path = "+sqlQuote(p)+" AND NOT (Encrypted AND Value LIKE 'oc1:%')"); got != 0 {
		t.Errorf("%d history rows are not sealed", got)
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_tree_log WHERE Path = "+sqlQuote(p)+" AND (Value LIKE '%v1-plain%' OR Value LIKE '%v2-plain%')"); got != 0 {
		t.Errorf("plaintext left in the history")
	}

	type entry struct {
		Version   int     `json:"version"`
		Data      *string `json:"data"`
		Encrypted bool    `json:"encrypted"`
		CanReveal bool    `json:"can_reveal"`
		Same      bool    `json:"same"`
	}
	var entries []entry
	if err := json.Unmarshal(adminReqOK(t, "GET", "/log"+p, nil), &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("history entries = %d, want 3", len(entries))
	}
	for _, e := range entries {
		if e.Data != nil || !e.Encrypted || !e.CanReveal {
			t.Errorf("masked history entry v%d: %+v", e.Version, e)
		}
		if e.Same != (e.Version == 3) {
			t.Errorf("same of v%d = %v", e.Version, e.Same)
		}
	}
	if err := json.Unmarshal(adminReqOK(t, "GET", "/log"+p+"?reveal=1", nil), &entries); err != nil {
		t.Fatal(err)
	}
	want := map[int]string{1: "v1-plain", 2: "v2-plain", 3: "v3-secret"}
	for _, e := range entries {
		if e.Data == nil || *e.Data != want[e.Version] {
			t.Errorf("revealed history entry v%d = %v, want %s", e.Version, e.Data, want[e.Version])
		}
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_audit WHERE Path = "+sqlQuote(p)+" AND Kind = 'reveal_history'"); got != 1 {
		t.Errorf("history reveal audit rows = %d, want 1 (one node)", got)
	}
}

// TestHistorySealedByType: sealing the history of a parameter that just became
// encrypted touches only what an encrypted parameter could have carried —
// null, symlink and case versions stay as they are, a text version is sealed
// whole — and the sealed versions reveal and restore.
func TestHistorySealedByType(t *testing.T) {
	target := uniqPath("seal-target")
	mkParam(t, target, "text/plain", "target", "mk target")
	p := uniqPath("enc-seal-types")
	mkParam(t, p, "application/x-null", "", "mk")
	caseValue := `[{"mime":"text/plain","value":"branch-plain"},{"server":"srv*","mime":"application/x-symlink","value":"` + target + `"}]`
	for _, form := range []url.Values{
		{"mime": {"application/x-symlink"}, "data": {target}, "comment": {"symlink"}},
		{"mime": {"application/x-case"}, "data": {caseValue}},
		{"mime": {"text/plain"}, "data": {"v4-plain"}},
		{"mime": {"text/plain"}, "data": {"v5-secret"}, "encrypted": {"1"}},
	} {
		if status, body := setParam(t, p, form); status != http.StatusOK {
			t.Fatalf("set %v: status %d, body %s", form, status, body)
		}
	}
	rows := dbQuery(t, "SELECT Version, ContentType, Encrypted, IFNULL(Value, 'NULL') FROM my_config_tree_log WHERE Path = "+sqlQuote(p)+" ORDER BY Version")
	if len(rows) != 5 {
		t.Fatalf("history rows = %v", rows)
	}
	byVersion := map[string][]string{}
	for _, r := range rows {
		byVersion[r[0]] = r
	}
	if r := byVersion["1"]; r[2] != "0" || r[3] != "NULL" {
		t.Errorf("null version touched: %v", r)
	}
	if r := byVersion["2"]; r[2] != "0" || r[3] != target {
		t.Errorf("symlink version touched: %v", r)
	}
	if r := byVersion["3"]; r[2] != "0" || r[3] != caseValue {
		t.Errorf("case version touched: %v", r)
	}
	for _, v := range []string{"4", "5"} {
		if r := byVersion[v]; r[2] != "1" || !strings.HasPrefix(r[3], "oc1:") {
			t.Errorf("text version %s not sealed: %v", v, r)
		}
	}
	// a later write to the sealed parameter leaves the sealed rows as they are
	sealed := dbQuery(t, "SELECT Version, Encrypted, Value FROM my_config_tree_log WHERE Path = "+sqlQuote(p)+" ORDER BY Version")
	if status, body := setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"v6-secret"}, "encrypted": {"1"}}); status != http.StatusOK {
		t.Fatalf("a later write: status %d, body %s", status, body)
	}
	if after := dbQuery(t, "SELECT Version, Encrypted, Value FROM my_config_tree_log WHERE Path = "+sqlQuote(p)+" AND Version <= 5 ORDER BY Version"); !reflect.DeepEqual(sealed, after) {
		t.Errorf("sealed rows rewritten by a later write: %v -> %v", sealed, after)
	}
	var entries []struct {
		Version int     `json:"version"`
		Data    *string `json:"data"`
	}
	if err := json.Unmarshal(adminReqOK(t, "GET", "/log"+p+"?reveal=1", nil), &entries); err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		switch e.Version {
		case 1:
			if e.Data != nil {
				t.Errorf("null version revealed as %q", *e.Data)
			}
		case 2:
			if e.Data == nil || *e.Data != target {
				t.Errorf("symlink version = %v, want %s", e.Data, target)
			}
		case 3:
			if e.Data == nil || !strings.Contains(*e.Data, "branch-plain") {
				t.Errorf("case version not revealed: %v", e.Data)
			}
		case 4:
			if e.Data == nil || *e.Data != "v4-plain" {
				t.Errorf("text version = %v, want v4-plain", e.Data)
			}
		}
	}
	// the symlink version restores as the plain symlink it is
	if status, body := restore(t, p, logEntryID(t, p, "symlink")); status != http.StatusOK {
		t.Fatalf("restore the symlink version: status %d, body %s", status, body)
	}
	if info := getParamInfo(t, p, ""); info.Mime != "application/x-symlink" || info.Encrypted || dataOf(info) != target {
		t.Errorf("restored symlink: mime=%s encrypted=%v data=%s", info.Mime, info.Encrypted, dataOf(info))
	}
}

func TestRestore(t *testing.T) {
	p := uniqPath("enc-restore")
	mkParam(t, p, "text/plain", "first", "mk")
	if status, body := setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"second-secret"}, "encrypted": {"1"}, "comment": {"encrypt"}}); status != http.StatusOK {
		t.Fatalf("encrypt: status %d, body %s", status, body)
	}
	// restoring the plaintext version keeps the parameter encrypted
	if status, body := restore(t, p, logEntryID(t, p, "mk")); status != http.StatusOK {
		t.Fatalf("restore: status %d, body %s", status, body)
	}
	info := getParamInfo(t, p, "?reveal=1")
	if info.Version != 3 || !info.Encrypted || dataOf(info) != "first" {
		t.Errorf("restored parameter: %+v data=%s", info, dataOf(info))
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_audit WHERE Path = "+sqlQuote(p)); got != 1 {
		t.Errorf("a restore counted as a reveal: %d audit rows", got)
	}
	// a deletion entry cannot be restored, nor an entry written at another path
	q := uniqPath("enc-restore-del")
	mkParam(t, q, "text/plain", "x", "mk")
	deleteParam(t, q, "rm")
	mkParam(t, q, "text/plain", "y", "mk again")
	status, body := restore(t, q, logEntryID(t, q, "rm"))
	expectError(t, status, body, http.StatusBadRequest, "InvalidValue")
	status, body = restore(t, q, logEntryID(t, p, "mk"))
	expectError(t, status, body, http.StatusBadRequest, "InvalidValue")
	// a stale version is refused like any write
	status, body = adminReq(t, "POST", "/config"+p, url.Values{"version": {"1"}, "restore": {strconv.Itoa(logEntryID(t, p, "mk"))}, "comment": {"stale"}})
	expectError(t, status, body, http.StatusBadRequest, "VersionNotMatch")
	// a non-numeric restore or version is a 400, not a 500
	status, body = adminReq(t, "POST", "/config"+p, url.Values{"version": {"1"}, "restore": {"abc"}, "comment": {"garbage"}})
	expectError(t, status, body, http.StatusBadRequest, "InvalidParam")
	status, body = adminReq(t, "POST", "/config"+p, url.Values{"version": {"abc"}, "mime": {"text/plain"}, "data": {"x"}, "comment": {"garbage"}})
	expectError(t, status, body, http.StatusBadRequest, "InvalidParam")
	status, body = adminReq(t, "DELETE", "/config"+p, url.Values{"version": {"abc"}, "comment": {"garbage"}})
	expectError(t, status, body, http.StatusBadRequest, "InvalidParam")
	// the history's reveal parameter is as strict as the parameter's
	status, body = adminReq(t, "GET", "/log"+p+"?reveal=yes", nil)
	expectError(t, status, body, http.StatusBadRequest, "InvalidParam")
	if status, _ := adminReq(t, "GET", "/log"+p+"?reveal=1", nil); status != http.StatusOK {
		t.Errorf("GET /log?reveal=1: status %d", status)
	}
}

// A path's history lists the entries of its former occupants too: a node
// soft-deleted and renamed aside when another was moved onto the path. Such
// an entry restores into the current node under the rights a reveal-and-post
// would have demanded.
func TestRestoreFormerOccupant(t *testing.T) {
	base := uniqPath("enc-occupant")
	mkParam(t, base, "application/x-null", "", "mk base")
	p, q := base+"/p", base+"/q"
	mkEncrypted(t, p, "text/plain", "old-secret", "mk zombie")
	deleteParam(t, p, "rm zombie")
	mkParam(t, q, "text/plain", "mover", "mk mover")
	moveParam(t, q, p, "takeover")
	zombie := logEntryID(t, p, "mk zombie")

	if status, body := restore(t, p, zombie); status != http.StatusOK {
		t.Fatalf("restore the former occupant's entry: status %d, body %s", status, body)
	}
	if info := getParamInfo(t, p, "?reveal=1"); !info.Encrypted || dataOf(info) != "old-secret" {
		t.Errorf("restored from the former occupant: %+v data=%s", info, dataOf(info))
	}
	// the same entry against another path
	other := uniqPath("enc-occupant-other")
	mkParam(t, other, "text/plain", "o", "mk")
	status, body := restore(t, other, zombie)
	expectError(t, status, body, http.StatusBadRequest, "InvalidValue")

	// a user with write access to the current node but read access only to
	// the former occupant (through the base: a deleted node keeps its own
	// groups, and this one has none) may not restore its secret under the rw
	// policy
	ro, rw := fmt.Sprintf("it-occ-ro-%d", os.Getpid()), fmt.Sprintf("it-occ-rw-%d", os.Getpid())
	adminReqOK(t, "POST", "/group/"+ro, nil)
	adminReqOK(t, "POST", "/group/"+ro+"/meerkat", nil)
	adminReqOK(t, "POST", "/access"+base, url.Values{"group": {ro}, "rw": {"false"}})
	adminReqOK(t, "POST", "/group/"+rw, nil)
	adminReqOK(t, "POST", "/group/"+rw+"/meerkat", nil)
	adminReqOK(t, "POST", "/access"+p, url.Values{"group": {rw}, "rw": {"true"}})
	status, body = reqAs(t, "meerkat", "meerkat", "POST", "/config"+p, url.Values{
		"version": {strconv.Itoa(paramVersion(t, p))}, "restore": {strconv.Itoa(zombie)}, "comment": {"rollback"},
	})
	expectError(t, status, body, http.StatusForbidden, "RevealDenied")
}

// A null, symlink, case or template version restored onto an encrypted
// parameter comes back as the plaintext it is, a symlink version whose
// target vanished is refused, and restoring an encrypted version onto a
// parameter declassified since makes it encrypted again and seals what was
// written in plaintext since — the only way a restore seals anything.
func TestRestoreByType(t *testing.T) {
	target := uniqPath("restore-target")
	mkParam(t, target, "text/plain", "target", "mk target")
	p := uniqPath("enc-restore-types")
	mkParam(t, p, "text/plain", "plain-one", "mk")
	caseValue := `[{"mime":"text/plain","value":"c"}]`
	for _, form := range []url.Values{
		{"mime": {"application/x-case"}, "data": {caseValue}, "comment": {"case"}},
		{"mime": {"application/x-symlink"}, "data": {target}, "comment": {"symlink"}},
		{"mime": {"text/plain"}, "data": {"sec"}, "encrypted": {"1"}, "comment": {"encrypt"}},
	} {
		if status, body := setParam(t, p, form); status != http.StatusOK {
			t.Fatalf("set %v: status %d, body %s", form, status, body)
		}
	}
	if status, body := restore(t, p, logEntryID(t, p, "case")); status != http.StatusOK {
		t.Fatalf("restore the case version: status %d, body %s", status, body)
	}
	if info := getParamInfo(t, p, ""); info.Mime != "application/x-case" || info.Encrypted || info.CanReveal || dataOf(info) != caseValue {
		t.Errorf("restored case: %+v data=%s", info, dataOf(info))
	}
	if got := dbInt(t, "SELECT Encrypted FROM my_config_tree WHERE Path = "+sqlQuote(p)); got != 0 {
		t.Errorf("the restored case carries the marker")
	}
	deleteParam(t, target, "rm target")
	status, body := restore(t, p, logEntryID(t, p, "symlink"))
	expectError(t, status, body, http.StatusBadRequest, "InvalidValue")

	s := uniqPath("enc-restore-seal")
	mkParam(t, s, "text/plain", "one", "mk")
	for _, form := range []url.Values{
		{"mime": {"text/plain"}, "data": {"two"}, "encrypted": {"1"}, "comment": {"encrypt"}},
		{"mime": {"text/plain"}, "data": {"three"}, "encrypted": {"0"}, "comment": {"declassify"}},
	} {
		if status, body := setParam(t, s, form); status != http.StatusOK {
			t.Fatalf("set %v: status %d, body %s", form, status, body)
		}
	}
	plainRows := func() int {
		return dbInt(t, "SELECT COUNT(*) FROM my_config_tree_log WHERE Path = "+sqlQuote(s)+" AND NOT Encrypted")
	}
	if got := plainRows(); got != 1 {
		t.Fatalf("plaintext rows after the declassification = %d, want 1", got)
	}
	if status, body := restore(t, s, logEntryID(t, s, "encrypt")); status != http.StatusOK {
		t.Fatalf("restore the encrypted version: status %d, body %s", status, body)
	}
	if info := getParamInfo(t, s, "?reveal=1"); !info.Encrypted || dataOf(info) != "two" {
		t.Errorf("restored encrypted version: %+v data=%s", info, dataOf(info))
	}
	if got := plainRows(); got != 0 {
		t.Errorf("plaintext rows after the restore = %d, want 0: the declassified version was not sealed", got)
	}
}

// A deleted parameter comes back from its history server-side: by the log
// entry, without a version, no reveal needed, encrypted as the version is.
func TestRestoreDeletedParameter(t *testing.T) {
	undelete := func(path string, entryID int) (int, []byte) {
		return adminReq(t, "POST", "/config"+path, url.Values{"restore": {strconv.Itoa(entryID)}, "comment": {"undelete"}})
	}
	d := uniqPath("enc-restore-deleted")
	mkEncrypted(t, d, "text/plain", "del-secret", "mk")
	oldEnvelope := dbQuery(t, "SELECT Value FROM my_config_tree WHERE Path = "+sqlQuote(d))[0][0]
	deleteParam(t, d, "rm")
	entry := logEntryID(t, d, "mk")
	if status, body := undelete(d, entry); status != http.StatusOK {
		t.Fatalf("restore into the deleted parameter: status %d, body %s", status, body)
	}
	info := getParamInfo(t, d, "")
	if info.Version != 3 || !info.Encrypted || !info.CanReveal || info.Data != nil {
		t.Errorf("restored parameter: %+v", info)
	}
	if got := dbQuery(t, "SELECT Value FROM my_config_tree WHERE Path = "+sqlQuote(d))[0][0]; !strings.HasPrefix(got, "oc1:") || got == oldEnvelope {
		t.Errorf("the restored value is not a fresh envelope: %s", got)
	}
	if dataOf(getParamInfo(t, d, "?reveal=1")) != "del-secret" {
		t.Errorf("the restored secret does not reveal")
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_audit WHERE Path = "+sqlQuote(d)); got != 1 {
		t.Errorf("audit rows = %d, want 1: the restore is not a reveal", got)
	}
	// without a version on a live parameter: a create's answer
	status, body := undelete(d, entry)
	expectError(t, status, body, http.StatusBadRequest, "AlreadyExists")

	// a plaintext life, then an encrypted one, then deleted: the first version
	// was sealed and comes back encrypted; the sealed rows stay as they are
	e := uniqPath("enc-restore-deleted-sealed")
	mkParam(t, e, "text/plain", "p-one", "mk")
	if status, body := setParam(t, e, url.Values{"mime": {"text/plain"}, "data": {"p-two"}, "encrypted": {"1"}, "comment": {"encrypt"}}); status != http.StatusOK {
		t.Fatalf("encrypt: status %d, body %s", status, body)
	}
	deleteParam(t, e, "rm")
	sealedRows := func() [][]string {
		return dbQuery(t, "SELECT Version, Encrypted, Value FROM my_config_tree_log WHERE Path = "+sqlQuote(e)+" AND Version <= 2 ORDER BY Version")
	}
	before := sealedRows()
	if status, body := undelete(e, logEntryID(t, e, "mk")); status != http.StatusOK {
		t.Fatalf("restore the sealed first version: status %d, body %s", status, body)
	}
	if info := getParamInfo(t, e, "?reveal=1"); !info.Encrypted || dataOf(info) != "p-one" {
		t.Errorf("restored sealed version: %+v data=%s", info, dataOf(info))
	}
	if after := sealedRows(); !reflect.DeepEqual(before, after) {
		t.Errorf("sealed rows touched by the restore: %v -> %v", before, after)
	}

	// a plaintext life only comes back plain
	f := uniqPath("enc-restore-deleted-plain")
	mkParam(t, f, "text/plain", "only-plain", "mk")
	deleteParam(t, f, "rm")
	if status, body := undelete(f, logEntryID(t, f, "mk")); status != http.StatusOK {
		t.Fatalf("restore the plain parameter: status %d, body %s", status, body)
	}
	if info := getParamInfo(t, f, ""); info.Encrypted || dataOf(info) != "only-plain" {
		t.Errorf("restored plain parameter: %+v", info)
	}

	// a path whose node was moved away keeps its history but has no row: the
	// restore creates a new node there
	m, m2 := uniqPath("enc-restore-moved"), uniqPath("enc-restore-moved-to")
	mkParam(t, m, "text/plain", "moved-val", "mk")
	moveParam(t, m, m2, "move")
	if status, body := undelete(m, logEntryID(t, m, "mk")); status != http.StatusOK {
		t.Fatalf("restore at the moved-away path: status %d, body %s", status, body)
	}
	if info := getParamInfo(t, m, ""); info.Version != 1 || dataOf(info) != "moved-val" {
		t.Errorf("restored at the moved-away path: %+v", info)
	}
	if dataOf(getParamInfo(t, m2, "")) != "moved-val" {
		t.Errorf("the moved node changed")
	}

	// a secret version while the switch is off, and a user without write
	// access to the parent
	g := uniqPath("enc-restore-deleted-gated")
	mkEncrypted(t, g, "text/plain", "gated-secret", "mk")
	deleteParam(t, g, "rm")
	sw := "/onlineconf/value-encryption"
	if status, body := setParam(t, sw, url.Values{"mime": {"text/plain"}, "data": {"0"}}); status != http.StatusOK {
		t.Fatalf("switch off: status %d, body %s", status, body)
	}
	status, body = undelete(g, logEntryID(t, g, "mk"))
	if s, b := setParam(t, sw, url.Values{"mime": {"text/plain"}, "data": {"1"}}); s != http.StatusOK {
		t.Fatalf("switch on: status %d, body %s", s, b)
	}
	expectError(t, status, body, http.StatusBadRequest, "EncryptionDisabled")
	status, body = reqAs(t, "meerkat", "meerkat", "POST", "/config"+g, url.Values{"restore": {strconv.Itoa(logEntryID(t, g, "mk"))}, "comment": {"undelete"}})
	expectError(t, status, body, http.StatusForbidden, "AccessDenied")
}

// A path's history lists the entries of nodes that left it alive too, and
// those nodes keep their own access: a writer of the old parent restores
// such an entry only with read access to the node — else the restore would
// hand them a value the history listing blanks for them. The undelete of a
// zombie by a writer of the parent is unaffected: a zombie keeps its groups,
// and the access function inherits per group from the nearest row.
func TestRestoreMovedAwayNode(t *testing.T) {
	base := uniqPath("enc-moved-acl")
	mkParam(t, base, "application/x-null", "", "mk base")
	rw := fmt.Sprintf("it-moved-rw-%d", os.Getpid())
	adminReqOK(t, "POST", "/group/"+rw, nil)
	adminReqOK(t, "POST", "/group/"+rw+"/meerkat", nil)
	adminReqOK(t, "POST", "/access"+base, url.Values{"group": {rw}, "rw": {"true"}})

	// the node leaves for a path under the root, where meerkat has no access
	p, away := base+"/p", fmt.Sprintf("/it-enc-moved-away-%d", os.Getpid())
	mkParam(t, p, "text/plain", "hidden-plain", "mk hidden")
	moveParam(t, p, away, "move away")
	entry := logEntryID(t, p, "mk hidden")

	// the listing blanks the entry's value for meerkat...
	status, body := reqAs(t, "meerkat", "meerkat", "GET", "/log"+p, nil)
	if status != http.StatusOK {
		t.Fatalf("GET /log%s as meerkat: status %d, body %s", p, status, body)
	}
	var entries []logEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		t.Fatal(err)
	}
	listed := false
	for _, e := range entries {
		if e.ID == entry {
			listed = true
			if e.Data != nil {
				t.Errorf("the moved node's entry shows its value to meerkat: %s", *e.Data)
			}
		}
	}
	if !listed {
		t.Fatalf("the moved node's entry is not listed at %s", p)
	}
	// ...and the restore into the old path is refused the same way
	restoreAsMeerkat := func() (int, []byte) {
		return reqAs(t, "meerkat", "meerkat", "POST", "/config"+p, url.Values{"restore": {strconv.Itoa(entry)}, "comment": {"restore"}})
	}
	status, body = restoreAsMeerkat()
	expectError(t, status, body, http.StatusForbidden, "AccessDenied")
	if s, _ := getParamInfoAs(t, "admin", "admin", p, ""); s != http.StatusNotFound {
		t.Errorf("a node appeared at %s: status %d", p, s)
	}

	// with read access to the moved node the restore goes through
	ro := fmt.Sprintf("it-moved-ro-%d", os.Getpid())
	adminReqOK(t, "POST", "/group/"+ro, nil)
	adminReqOK(t, "POST", "/group/"+ro+"/meerkat", nil)
	adminReqOK(t, "POST", "/access"+away, url.Values{"group": {ro}, "rw": {"false"}})
	if status, body := restoreAsMeerkat(); status != http.StatusOK {
		t.Fatalf("restore with read access to the moved node: status %d, body %s", status, body)
	}
	if info := getParamInfo(t, p, ""); dataOf(info) != "hidden-plain" {
		t.Errorf("restored at the old path: %+v", info)
	}

	// the undelete of a zombie by a writer of the parent needs nothing more
	z := base + "/z"
	mkParam(t, z, "text/plain", "zombie", "mk zombie")
	deleteParam(t, z, "rm zombie")
	status, body = reqAs(t, "meerkat", "meerkat", "POST", "/config"+z, url.Values{"restore": {strconv.Itoa(logEntryID(t, z, "mk zombie"))}, "comment": {"undelete"}})
	if status != http.StatusOK {
		t.Fatalf("undelete by a writer of the parent: status %d, body %s", status, body)
	}
	if info := getParamInfo(t, z, ""); dataOf(info) != "zombie" {
		t.Errorf("undeleted zombie: %+v", info)
	}
}

func TestEncryptionGated(t *testing.T) {
	p := uniqPath("enc-gate")
	mkEncrypted(t, p, "text/plain", "gated", "mk")

	sw := "/onlineconf/value-encryption"
	if status, body := setParam(t, sw, url.Values{"mime": {"text/plain"}, "data": {"0"}}); status != http.StatusOK {
		t.Fatalf("switch off: status %d, body %s", status, body)
	}
	t.Cleanup(func() { setParam(t, sw, url.Values{"mime": {"text/plain"}, "data": {"1"}}) })

	var ui struct {
		Encryption struct {
			Enabled bool `json:"enabled"`
		} `json:"encryption"`
	}
	if err := json.Unmarshal(adminReqOK(t, "GET", "/ui-config", nil), &ui); err != nil {
		t.Fatal(err)
	}
	if ui.Encryption.Enabled {
		t.Errorf("/ui-config reports encryption enabled with the switch off")
	}
	// a user without access to the avatar parameter still gets the config:
	// the avatar is read by path, not through the caller's rights
	ensureParam(t, "/onlineconf/ui", "application/x-null", "", "ui")
	ensureParam(t, "/onlineconf/ui/avatar", "application/json", `{"uri": "https://avatars.invalid/", "gravatar": false}`, "avatar")
	var uiAs struct {
		Avatar     *json.RawMessage `json:"avatar"`
		Encryption *json.RawMessage `json:"encryption"`
	}
	if s, b := reqAs(t, "meerkat", "meerkat", "GET", "/ui-config", nil); s != http.StatusOK || json.Unmarshal(b, &uiAs) != nil || uiAs.Avatar == nil || uiAs.Encryption == nil {
		t.Errorf("/ui-config for a user without access to the avatar: status %d, body %s", s, b)
	}
	status, body := adminReq(t, "POST", "/config"+uniqPath("enc-gate2"), url.Values{"mime": {"text/plain"}, "data": {"x"}, "encrypted": {"1"}, "comment": {"mk"}})
	expectError(t, status, body, http.StatusBadRequest, "EncryptionDisabled")
	// a rollback whose result would hold a secret is an encrypted write
	status, body = restore(t, p, logEntryID(t, p, "mk"))
	expectError(t, status, body, http.StatusBadRequest, "EncryptionDisabled")
	// the field is 0 or 1: a client spelling it otherwise is told, not taken
	// to declassify
	for _, v := range []string{"true", "on", ""} {
		status, body = setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"x"}, "encrypted": {v}})
		expectError(t, status, body, http.StatusBadRequest, "InvalidParam")
	}
	if info := getParamInfo(t, p, "?reveal=1"); !info.Encrypted || dataOf(info) != "gated" {
		t.Errorf("a refused write changed the parameter: %+v", info)
	}
	// reads, reveals and declassification keep working: the switch is not a trap
	if dataOf(getParamInfo(t, p, "?reveal=1")) != "gated" {
		t.Errorf("reveal with the switch off failed")
	}
	if status, body := setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"freed"}, "encrypted": {"0"}}); status != http.StatusOK {
		t.Errorf("declassify with the switch off: status %d, body %s", status, body)
	}
}

// a template is never encrypted: its ${/path} references stay searchable, and
// a secret it needs lives in an encrypted parameter it references
func TestTemplatesNotEncrypted(t *testing.T) {
	tmpl := uniqPath("enc-tmpl")
	status, body := adminReq(t, "POST", "/config"+tmpl, url.Values{
		"mime": {"application/x-template"}, "data": {"${/x}"}, "encrypted": {"1"}, "comment": {"mk"},
	})
	expectError(t, status, body, http.StatusBadRequest, "EncryptionNotAllowed")
	// a template branch with a stray marker is stored as the plaintext
	// template it is
	caseRef := uniqPath("enc-tmpl-case")
	mkParam(t, caseRef, "application/x-case", `[{"mime":"application/x-template","encrypted":true,"value":"${/x}"}]`, "mk")
	if info := getParamInfo(t, caseRef, ""); info.Encrypted || info.CanReveal {
		t.Errorf("case with a marked template branch: %+v", info)
	}

	// a plaintext template referencing an encrypted parameter is served plain,
	// delivered expanded with the plaintext, and keeps its target from being
	// deleted — as a parameter and as a branch of a case
	allowEphemeralClient(t)
	name := fmt.Sprintf("it-tmpl-%d", os.Getpid())
	secret := "/gopher/" + name + "-secret"
	mkEncrypted(t, secret, "text/plain", "s3cr3t", "mk secret")
	mkParam(t, "/gopher/"+name, "application/x-template", "user:pass@${"+secret+"}", "mk template")
	mkParam(t, "/gopher/"+name+"-case", "application/x-case", `[{"mime":"application/x-template","value":"${`+secret+`}"}]`, "mk case")
	for _, p := range []string{"/gopher/" + name, "/gopher/" + name + "-case"} {
		if info := getParamInfo(t, p, ""); info.Encrypted || info.CanReveal || info.Data == nil || !strings.Contains(*info.Data, "${"+secret+"}") {
			t.Errorf("template %s is not served plain: %+v", p, info)
		}
	}
	wire := "/onlineconf/module/TREE/gopher/" + name
	nodes := awaitDelivered(t, wire, wire+"-case")
	expectDelivered(t, nodes, wire, "text/plain", "user:pass@s3cr3t")
	expectDelivered(t, nodes, wire+"-case", "text/plain", "s3cr3t")

	rmSecret := func() (int, []byte) {
		return adminReq(t, "DELETE", "/config"+secret, url.Values{"version": {strconv.Itoa(paramVersion(t, secret))}, "comment": {"rm"}})
	}
	status, body = rmSecret()
	expectError(t, status, body, http.StatusBadRequest, "DeletedTemplate")
	if !strings.Contains(string(body), "/gopher/"+name) {
		t.Errorf("the error does not name the template: %s", body)
	}
	deleteParam(t, "/gopher/"+name, "rm template")
	status, body = rmSecret()
	expectError(t, status, body, http.StatusBadRequest, "DeletedTemplate")
	if !strings.Contains(string(body), "/gopher/"+name+"-case") {
		t.Errorf("the error does not name the case: %s", body)
	}
	deleteParam(t, "/gopher/"+name+"-case", "rm case")
	deleteParam(t, secret, "rm secret")
}

// An instance without a key file refuses to start against a database holding
// encrypted values: a throwaway container of the serving instance's image, on
// its network and with its environment except the key file variable (which
// the entrypoint maps to encryption.key_file), so that the serving instance
// is not touched and the image is not rebuilt.
func TestStartupWithoutKeyFails(t *testing.T) {
	mkEncrypted(t, uniqPath("enc-startup"), "text/plain", "s", "mk")
	id, err := composeStdout("ps", "-q", "onlineconf-admin")
	if err != nil || strings.TrimSpace(id) == "" {
		t.Fatalf("compose ps: %v %q", err, id)
	}
	inspect := func(format string) string {
		out, err := exec.Command("docker", "inspect", "--format", format, strings.TrimSpace(id)).Output()
		if err != nil {
			t.Fatalf("docker inspect: %v", err)
		}
		return strings.TrimSpace(string(out))
	}
	args := []string{"run", "--rm", "--network", inspect("{{range $name, $net := .NetworkSettings.Networks}}{{$name}}{{end}}")}
	for _, e := range strings.Split(inspect("{{range .Config.Env}}{{println .}}{{end}}"), "\n") {
		if strings.HasPrefix(e, "ONLINECONF_") && !strings.HasPrefix(e, "ONLINECONF_ENCRYPTION_KEY_FILE=") {
			args = append(args, "-e", e)
		}
	}
	args = append(args, "-e", "ONLINECONF_ENCRYPTION_KEY_FILE=", inspect("{{.Config.Image}}"))
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("an instance without a key file did not exit in time:\n%s", out)
	}
	if err == nil {
		t.Fatalf("an instance without a key file started against encrypted values:\n%s", out)
	}
	// the startup check and the resolver's first load race for the exit:
	// either refusal names the cause
	if !strings.Contains(string(out), "encrypted parameters exist but encryption.key_file is not configured") &&
		!strings.Contains(string(out), "an encrypted value cannot be decrypted with the configured keys") {
		t.Errorf("the instance failed for another reason:\n%s", out)
	}

	// 2. with no live encrypted parameter — the suite's replaced by plaintext,
	// the deleted ones and the sealed history staying — the instance comes
	// up: what the resolver loads needs no key, and a restore of a deleted
	// secret or a reveal of a sealed version is what fails, at use
	for _, row := range dbQuery(t, "SELECT Path FROM my_config_tree WHERE Encrypted = 1 AND NOT Deleted") {
		if status, body := setParam(t, row[0], url.Values{"mime": {"text/plain"}, "data": {"declassified"}, "encrypted": {"0"}, "comment": {"declassify"}}); status != http.StatusOK {
			t.Fatalf("declassify %s: status %d, body %s", row[0], status, body)
		}
	}
	if n := dbInt(t, "SELECT COUNT(*) FROM my_config_tree WHERE Encrypted = 1 AND Deleted"); n == 0 {
		t.Errorf("no deleted secret left in the database: the case is not covered")
	}
	if n := dbInt(t, "SELECT COUNT(*) FROM my_config_tree_log WHERE Encrypted = 1"); n == 0 {
		t.Errorf("no sealed history left in the database: the case is not covered")
	}
	detached, err := exec.Command("docker", append([]string{"run", "-d"}, args[2:]...)...).Output()
	if err != nil {
		t.Fatalf("docker run -d: %v", err)
	}
	container := strings.TrimSpace(string(detached))
	defer exec.Command("docker", "rm", "-f", container).Run()
	time.Sleep(10 * time.Second)
	running, _ := exec.Command("docker", "inspect", "--format", "{{.State.Running}}", container).Output()
	logs, _ := exec.Command("docker", "logs", container).CombinedOutput()
	if strings.TrimSpace(string(running)) != "true" {
		t.Fatalf("the instance without a key file did not stay up against deleted secrets and sealed history only:\n%s", logs)
	}
	if !strings.Contains(string(logs), "value encryption is not configured") {
		t.Errorf("the instance without a key file did not say so:\n%s", logs)
	}
}

// TestSweepQueryPlan pins the plan of the sweep's batch query on the MySQL
// the stack runs: a range on the Encrypted index extended by the primary
// key, in ID order, without a filesort — else a pass costs the square of
// the encrypted rows.
func TestSweepQueryPlan(t *testing.T) {
	for _, table := range []string{"my_config_tree", "my_config_tree_log"} {
		rows := dbQuery(t, "EXPLAIN SELECT ID, Value FROM "+table+" FORCE INDEX (Encrypted) WHERE Encrypted = 1 AND ID > 0 ORDER BY ID LIMIT 500")
		if len(rows) != 1 || len(rows[0]) < 12 {
			t.Fatalf("%s: unexpected EXPLAIN output %v", table, rows)
		}
		// id, select_type, table, partitions, type, possible_keys, key, key_len, ref, rows, filtered, Extra
		key, keyLen, extra := rows[0][6], rows[0][7], rows[0][11]
		if key != "Encrypted" || keyLen != "9" || strings.Contains(extra, "filesort") {
			t.Errorf("%s: key=%s key_len=%s extra=%q, want the Encrypted index with the primary key (9) and no filesort", table, key, keyLen, extra)
		}
	}
}

// --- key rotation ---------------------------------------------------------------

type encryptionStatus struct {
	Wanted struct {
		ID     string `json:"id"`
		Source string `json:"source"`
		Held   bool   `json:"held"`
	} `json:"wanted"`
	Keys []struct {
		ID         string `json:"id"`
		Registered bool   `json:"registered"`
		Verified   bool   `json:"verified"`
		Recorded   bool   `json:"recorded"`
	} `json:"keys"`
	Instance string `json:"instance"`
	Sweep    struct {
		State       string `json:"state"`
		Target      string `json:"target"`
		Reencrypted int    `json:"reencrypted"`
	} `json:"sweep"`
}

// awaitSweep waits for this instance's sweep to report done for the target:
// the condition the README gives for dropping the old key.
func awaitSweep(t *testing.T, target string) encryptionStatus {
	t.Helper()
	return awaitSweepFor(t, target, 60*time.Second)
}

func awaitSweepFor(t *testing.T, target string, timeout time.Duration) encryptionStatus {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		status := getEncryptionStatus(t)
		if status.Sweep.State == "done" && status.Sweep.Target == target {
			if status.Instance == "" {
				t.Errorf("the status does not name the instance")
			}
			return status
		}
		if time.Now().After(deadline) {
			t.Fatalf("the sweep to %s did not complete: %+v", target, status.Sweep)
		}
		time.Sleep(time.Second)
	}
}

// awaitSweepState waits for this instance's sweep to report the state.
func awaitSweepState(t *testing.T, state string, timeout time.Duration) encryptionStatus {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		status := getEncryptionStatus(t)
		if status.Sweep.State == state {
			return status
		}
		if time.Now().After(deadline) {
			t.Fatalf("the sweep did not report %s: %+v", state, status.Sweep)
		}
		time.Sleep(time.Second)
	}
}

// dbSession runs statements in one session of the stack's database, in the
// background: a lock the session takes is held until its statements end, or
// until it is killed (killDBSession).
func dbSession(statements string) <-chan error {
	done := make(chan error, 1)
	go func() {
		_, err := composeStdout("exec", "-T", "onlineconf-database",
			"mysql", "-N", "-uonlineconf", "-ponlineconf", "onlineconf", "-e", statements)
		done <- err
	}()
	return done
}

// killDBSession kills the sessions whose current statement matches the LIKE
// pattern; the locks they held go with their connections.
func killDBSession(t *testing.T, infoLike string) {
	t.Helper()
	for _, row := range dbQuery(t, "SELECT ID FROM information_schema.PROCESSLIST WHERE INFO LIKE "+sqlQuote(infoLike)) {
		dbQuery(t, "KILL "+row[0])
	}
}

// awaitSweepLock waits until the sweep's advisory lock is held, or free.
func awaitSweepLock(t *testing.T, held bool) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for {
		rows := dbQuery(t, "SELECT IS_USED_LOCK(CONCAT(DATABASE(), '/encryption-sweep'))")
		if len(rows) == 1 && (rows[0][0] != "NULL") == held {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("the sweep lock did not become held=%v: %v", held, rows)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func getEncryptionStatus(t *testing.T) encryptionStatus {
	t.Helper()
	var s encryptionStatus
	if err := json.Unmarshal(adminReqOK(t, "GET", "/encryption-status", nil), &s); err != nil {
		t.Fatal(err)
	}
	return s
}

// containerKeyring returns the key lines (id and material, active first) of
// the admin container's current key file.
func containerKeyring(t *testing.T) []string {
	t.Helper()
	out, err := composeStdout("exec", "-T", "onlineconf-admin", "cat", "/usr/local/etc/onlineconf-admin.keys")
	if err != nil {
		t.Fatalf("read the container's key file: %v", err)
	}
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		t.Fatalf("the container's key file holds no keys:\n%s", out)
	}
	return lines
}

// installKeyring replaces the admin container's key file and restarts it.
func installKeyring(t *testing.T, lines ...string) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "keys")
	if err := os.WriteFile(file, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if out, err := compose("cp", file, "onlineconf-admin:/usr/local/etc/onlineconf-admin.keys"); err != nil {
		t.Fatalf("compose cp: %v\n%s", err, out)
	}
	if out, err := compose("restart", "onlineconf-admin"); err != nil {
		t.Fatalf("compose restart: %v\n%s", err, out)
	}
}

func randomKey(t *testing.T) string {
	t.Helper()
	raw := make([]byte, 32)
	f, err := os.Open("/dev/urandom")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := io.ReadFull(f, raw); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

// TestKeyRotationSweep rotates the stack's encryption key the way an
// operator does: the new key joins the key file (a restart here, through
// docker compose, whatever manages the stack), then its id is written to
// /onlineconf/value-encryption/key-id and the sweep re-encrypts everything
// under it, with a writer racing it. A key the instance lacks changes
// nothing, a retired key is promoted again like any other, and the key
// file's order means nothing once a key is registered. The test leaves the stack
// with the new key alone in the key file — every envelope is under it, a
// consistent state every other test works against.
func TestKeyRotationSweep(t *testing.T) {
	p := uniqPath("enc-rotate")
	mkEncrypted(t, p, "text/plain", "rotated-secret", "mk")
	setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"rotated-secret"}, "encrypted": {"1"}})

	current := containerKeyring(t)
	oldID := strings.Fields(current[0])[0]
	registeredActive := func() string {
		got := dbQuery(t, "SELECT KeyID FROM my_config_encryption_key WHERE Active")
		if len(got) != 1 {
			t.Fatalf("registered active key = %v", got)
		}
		return got[0][0]
	}
	if got := registeredActive(); got != oldID {
		t.Fatalf("registered active key = %s, the container's active key is %s", got, oldID)
	}
	newID := fmt.Sprintf("rot%d", os.Getpid())
	newLine := newID + " " + randomKey(t)
	// envelopes under a key, by the stored text alone: the check must not
	// trust the sweep's own selection of rows
	encryptedRows := func(table, keyID string) int {
		return dbInt(t, "SELECT COUNT(*) FROM "+table+" WHERE Value LIKE '%oc1:"+keyID+":%'")
	}
	if encryptedRows("my_config_tree", oldID) == 0 {
		t.Fatalf("no tree rows under the current key %s", oldID)
	}

	// 1. the new key joins the key file, as a retired line: nothing changes
	installKeyring(t, append(append([]string{}, current...), newLine)...)
	if err := waitReady(2 * time.Minute); err != nil {
		t.Fatal(err)
	}
	status := getEncryptionStatus(t)
	if status.Wanted.ID != oldID || status.Wanted.Source != "registry" || !status.Wanted.Held {
		t.Errorf("wanted key after adding a retired key = %+v, want %s from the registry", status.Wanted, oldID)
	}
	for _, k := range status.Keys {
		if k.Recorded != (k.ID == oldID) {
			t.Errorf("key status after adding a retired key: %+v", k)
		}
	}

	// 2a. a request queued behind a held lock is not lost, and the latest
	// wish wins: while a session holds the sweep's advisory lock, the new
	// key's id is written — the goroutine reports waiting, without a target
	// — and then the old key's id again; once the session is killed, the
	// sweep plans with the latest wish: one pass that rewrites nothing, the
	// old key still recorded
	session := dbSession("SELECT GET_LOCK(CONCAT(DATABASE(), '/encryption-sweep'), 60); DO SLEEP(120)")
	t.Cleanup(func() { killDBSession(t, "DO SLEEP(120)%") }) // never left held behind a failure
	awaitSweepLock(t, true)
	writeKeyID(t, newID)
	if status := awaitSweepState(t, "waiting", 30*time.Second); status.Sweep.Target != "" {
		t.Errorf("a waiting sweep names a target: %+v", status.Sweep)
	}
	writeKeyID(t, oldID)
	time.Sleep(15 * time.Second) // the watcher sees the second write at a reload after the log margin
	killDBSession(t, "DO SLEEP(120)%")
	<-session
	awaitSweepLock(t, false)
	status = awaitSweep(t, oldID)
	if status.Sweep.Reencrypted != 0 {
		t.Errorf("the sweep behind the held lock rewrote %d rows: the latest wish is the key in use", status.Sweep.Reencrypted)
	}
	if got := registeredActive(); got != oldID {
		t.Errorf("registered active key after the queued requests = %s, want %s", got, oldID)
	}

	// 2. the promotion: the operator writes the new key's id; a writer keeps
	// rewriting a secret meanwhile, and every version it writes must end up
	// under the new key — a write reads the key in use under a shared lock
	// in its transaction, so the flip of the registry waits for it, or it
	// sees the new key
	w := uniqPath("enc-rotate-writer")
	mkEncrypted(t, w, "text/plain", "w0", "mk")
	stop, stopped := make(chan struct{}), make(chan struct{})
	var writes int
	var writerErr error
	go func() {
		defer close(stopped)
		for i := 1; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			if status, body := setParam(t, w, url.Values{"mime": {"text/plain"}, "data": {fmt.Sprintf("w%d", i)}, "encrypted": {"1"}}); status != http.StatusOK {
				writerErr = fmt.Errorf("concurrent write %d: status %d, body %s", i, status, body)
				return
			}
			writes = i
		}
	}()
	writeKeyID(t, newID)
	status = awaitSweep(t, newID)
	close(stop)
	<-stopped
	if writerErr != nil {
		t.Fatal(writerErr)
	}
	if writes == 0 {
		t.Errorf("the concurrent writer wrote nothing")
	}
	if status.Sweep.Reencrypted == 0 {
		t.Errorf("the sweep re-encrypted nothing: %+v", status.Sweep)
	}
	if status.Wanted.ID != newID || status.Wanted.Source != "parameter" || !status.Wanted.Held {
		t.Errorf("wanted key after the promotion = %+v", status.Wanted)
	}
	for _, k := range status.Keys {
		if !k.Registered || !k.Verified || k.Recorded != (k.ID == newID) {
			t.Errorf("key status after rotation: %+v", k)
		}
	}
	if dataOf(getParamInfo(t, w, "?reveal=1")) != fmt.Sprintf("w%d", writes) {
		t.Errorf("the concurrent writer's last value was lost")
	}
	if got := encryptedRows("my_config_tree", oldID); got != 0 {
		t.Errorf("%d tree rows still under the old key", got)
	}
	if got := encryptedRows("my_config_tree_log", oldID); got != 0 {
		t.Errorf("%d log rows still under the old key", got)
	}
	if got := encryptedRows("my_config_tree", newID); got == 0 {
		t.Errorf("no tree rows under the new key")
	}
	if v := paramVersion(t, p); v != 2 {
		t.Errorf("the sweep bumped the version: %d", v)
	}
	if dataOf(getParamInfo(t, p, "?reveal=1")) != "rotated-secret" {
		t.Errorf("value lost in rotation")
	}
	// a write after the promotion lands under the new key
	setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"rotated-again"}, "encrypted": {"1"}})
	if got := dbQuery(t, "SELECT Value FROM my_config_tree WHERE Path = "+sqlQuote(p)); len(got) != 1 || !strings.HasPrefix(got[0][0], "oc1:"+newID+":") {
		t.Errorf("a write after the promotion is not under the new key: %v", got)
	}

	// 3. a key this instance does not hold is refused and changes nothing
	since := time.Now()
	writeKeyID(t, "nokey"+newID)
	if !adminLogsSince(t, since, "not in this instance's key file", 20*time.Second) {
		t.Errorf("an unknown key id was not reported")
	}
	if got := registeredActive(); got != newID {
		t.Errorf("a mistaken key id changed the registered active key to %s", got)
	}
	// a retired key is promoted again like any other: back to the old key,
	// then forward to the new one
	writeKeyID(t, oldID)
	awaitSweep(t, oldID)
	if got := registeredActive(); got != oldID || encryptedRows("my_config_tree", newID) != 0 || encryptedRows("my_config_tree_log", newID) != 0 {
		t.Errorf("rotation back to %s: active=%s, rows under %s: %d in the tree, %d in the log", oldID, got, newID, encryptedRows("my_config_tree", newID), encryptedRows("my_config_tree_log", newID))
	}
	writeKeyID(t, newID)
	awaitSweep(t, newID)
	if got := registeredActive(); got != newID || encryptedRows("my_config_tree", oldID) != 0 || encryptedRows("my_config_tree_log", oldID) != 0 {
		t.Errorf("rotation forward to %s again: active=%s, rows under %s: %d in the tree, %d in the log", newID, got, oldID, encryptedRows("my_config_tree", oldID), encryptedRows("my_config_tree_log", oldID))
	}

	// 4. the key file's order means nothing once a key is registered: the
	// old key first, a restart, and the new key stays the one in use
	installKeyring(t, append(append([]string{}, current...), newLine)...)
	if err := waitReady(2 * time.Minute); err != nil {
		t.Fatal(err)
	}
	status = getEncryptionStatus(t)
	for _, k := range status.Keys {
		if k.Recorded != (k.ID == newID) {
			t.Errorf("key status after a restart with the old key first: %+v", k)
		}
	}
	if dataOf(getParamInfo(t, p, "?reveal=1")) != "rotated-again" {
		t.Errorf("value unreadable after the restart")
	}

	// 5. wrong material under a registered id refuses to start
	since = time.Now()
	installKeyring(t, append([]string{newID + " " + randomKey(t)}, current...)...)
	if !adminLogsSince(t, since, "differs from the key registered", 60*time.Second) {
		t.Errorf("wrong key material was not refused")
	}

	// 6. the old key leaves the key file: every envelope is under the new
	// key, so the instance comes up and reveals — the state later tests run
	// against
	installKeyring(t, newLine)
	if err := waitReady(2 * time.Minute); err != nil {
		t.Fatal(err)
	}
	if dataOf(getParamInfo(t, p, "?reveal=1")) != "rotated-again" {
		t.Errorf("value unreadable with the new key alone")
	}
	// the old key stays registered, no longer held
	status = getEncryptionStatus(t)
	for _, k := range status.Keys {
		if k.Verified != (k.ID == newID) || k.Recorded != (k.ID == newID) {
			t.Errorf("key status with the new key alone: %+v", k)
		}
	}
}

// TestSweepFlipRetried: the flip of the key registry waits for every
// in-flight encrypted write, and a write holding the shared key lock longer
// than the server's lock wait timeout fails it with a lock wait timeout —
// the sweep retries the flip instead of failing. A session holds the shared
// lock for 80 s: the first attempt times out at the server's 50 s, the
// retry succeeds once the session ends, and the log shows the retry — the
// proof, since a flip that merely waited would end the same way.
func TestSweepFlipRetried(t *testing.T) {
	current := containerKeyring(t)
	oldID := strings.Fields(current[0])[0]
	newID := fmt.Sprintf("flip%d", os.Getpid())
	installKeyring(t, append(append([]string{}, current...), newID+" "+randomKey(t))...)
	if err := waitReady(2 * time.Minute); err != nil {
		t.Fatal(err)
	}
	p := uniqPath("enc-flip")
	mkEncrypted(t, p, "text/plain", "flip-secret", "mk")

	since := time.Now()
	session := dbSession("BEGIN; SELECT KeyID FROM my_config_encryption_key WHERE Active LOCK IN SHARE MODE; DO SLEEP(80); COMMIT")
	t.Cleanup(func() { killDBSession(t, "DO SLEEP(80)%") })
	time.Sleep(3 * time.Second) // the session holds the shared lock
	writeKeyID(t, newID)
	status := awaitSweepFor(t, newID, 150*time.Second)
	if err := <-session; err != nil {
		t.Errorf("the session holding the shared lock: %v", err)
	}
	if !adminLogsSince(t, since, "retrying the flip of the key registry", 5*time.Second) || !adminLogsSince(t, since, "Lock wait timeout", 5*time.Second) {
		t.Errorf("the flip was not retried after a lock wait timeout")
	}
	if status.Sweep.Reencrypted == 0 {
		t.Errorf("the promotion re-encrypted nothing: %+v", status.Sweep)
	}
	if got := dbQuery(t, "SELECT KeyID FROM my_config_encryption_key WHERE Active"); len(got) != 1 || got[0][0] != newID {
		t.Errorf("registered active key after the retried flip = %v, want %s", got, newID)
	}
	if n := dbInt(t, "SELECT COUNT(*) FROM my_config_tree WHERE Value LIKE '%oc1:"+oldID+":%'"); n != 0 {
		t.Errorf("%d tree rows still under %s", n, oldID)
	}
	if dataOf(getParamInfo(t, p, "?reveal=1")) != "flip-secret" {
		t.Errorf("value lost in the retried promotion")
	}
}

// writeKeyID creates or sets /onlineconf/value-encryption/key-id.
func writeKeyID(t *testing.T, id string) {
	t.Helper()
	const path = "/onlineconf/value-encryption/key-id"
	if status, _ := adminReq(t, "GET", "/config"+path, nil); status == http.StatusNotFound {
		mkParam(t, path, "text/plain", id, "rotate")
		return
	}
	if status, body := setParam(t, path, url.Values{"mime": {"text/plain"}, "data": {id}}); status != http.StatusOK {
		t.Fatalf("write key id %s: status %d, body %s", id, status, body)
	}
}
