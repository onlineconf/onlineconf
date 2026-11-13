//go:build integration

package integration

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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

func TestEncryptedCaseBranches(t *testing.T) {
	p := uniqPath("enc-case")
	value := `[{"group":"prod","mime":"application/json","encrypted":true,"value":"{\"token\":\"prod-secret\"}"},` +
		`{"mime":"application/x-case","value":"[{\"server\":\"a*\",\"mime\":\"text/plain\",\"encrypted\":true,\"value\":\"inner-secret\"},{\"mime\":\"text/plain\",\"value\":\"dev\"}]"}]`
	mkParam(t, p, "application/x-case", value, "mk case")

	stored := dbQuery(t, "SELECT Value FROM my_config_tree WHERE Path = "+sqlQuote(p))[0][0]
	if strings.Contains(stored, "prod-secret") || strings.Contains(stored, "inner-secret") {
		t.Errorf("plaintext branch stored: %s", stored)
	}
	if !strings.Contains(stored, `oc1:`) || !strings.Contains(stored, `prod`) || !strings.Contains(stored, `dev`) {
		t.Errorf("stored case lacks envelopes, conditions or plain branches: %s", stored)
	}
	if got := dbInt(t, "SELECT Encrypted FROM my_config_tree WHERE Path = "+sqlQuote(p)); got != 0 {
		t.Errorf("a case carries the node-level marker")
	}

	info := getParamInfo(t, p, "")
	if !info.CanReveal || info.Data == nil {
		t.Fatalf("masked case: can_reveal=%v data=%s", info.CanReveal, dataOf(info))
	}
	masked := *info.Data
	// the nested case is a JSON string inside the JSON, so its plain branch
	// appears escaped
	if strings.Contains(masked, "prod-secret") || strings.Contains(masked, "inner-secret") || !strings.Contains(masked, `"encrypted":true`) || !strings.Contains(masked, `dev`) {
		t.Errorf("masked case wrong: %s", masked)
	}
	revealed := dataOf(getParamInfo(t, p, "?reveal=1"))
	if !strings.Contains(revealed, "prod-secret") || !strings.Contains(revealed, "inner-secret") {
		t.Errorf("revealed case lacks the branch plaintexts: %s", revealed)
	}
	// a stale client posting the masked case back (branch values null) is refused
	status, body := setParam(t, p, url.Values{"mime": {"application/x-case"}, "data": {masked}})
	expectError(t, status, body, http.StatusBadRequest, "InvalidValue")
	// node-level marker on a case is refused too
	status, body = setParam(t, p, url.Values{"mime": {"application/x-case"}, "data": {revealed}, "encrypted": {"1"}})
	expectError(t, status, body, http.StatusBadRequest, "EncryptionNotAllowed")
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

	var ui struct {
		Encryption struct {
			Enabled bool   `json:"enabled"`
			Reveal  string `json:"reveal"`
		} `json:"encryption"`
	}
	if err := json.Unmarshal(adminReqOK(t, "GET", "/ui-config", nil), &ui); err != nil {
		t.Fatal(err)
	}
	if !ui.Encryption.Enabled || ui.Encryption.Reveal != "ro" {
		t.Errorf("/ui-config = %+v, want encryption on with the ro policy", ui)
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

func TestEncryptedDelivery(t *testing.T) {
	// the test client is identified by its IP; make every network ephemeral so
	// no reverse DNS is attempted on the docker bridge address
	ephemeral := getParamInfo(t, "/onlineconf/ephemeral-ip", "")
	if status, body := setParam(t, "/onlineconf/ephemeral-ip", url.Values{"mime": {"text/plain"}, "data": {"0.0.0.0/0"}}); status != http.StatusOK {
		t.Fatalf("ephemeral-ip: status %d, body %s", status, body)
	}
	t.Cleanup(func() {
		setParam(t, "/onlineconf/ephemeral-ip", url.Values{"mime": {ephemeral.Mime}, "data": {dataOf(ephemeral)}})
	})

	name := fmt.Sprintf("it-delivery-%d", os.Getpid())
	mkEncrypted(t, "/gopher/"+name, "application/json", `{"token": "delivered-secret"}`, "mk")
	mkParam(t, "/gopher/"+name+"-case", "application/x-case",
		`[{"mime":"text/plain","encrypted":true,"value":"case-secret"}]`, "mk")

	wire := "/onlineconf/module/TREE/gopher/" + name
	deadline := time.Now().Add(30 * time.Second)
	var delivered []string
	for {
		status, cfg := fetchClientConfig(t)
		if status == http.StatusOK {
			seen := map[string]bool{}
			delivered = delivered[:0]
			for _, n := range cfg.Nodes {
				delivered = append(delivered, n.Path)
				switch n.Path {
				case wire:
					seen[n.Path] = true
					if n.ContentType != "application/json" || n.Value == nil || !strings.Contains(*n.Value, "delivered-secret") {
						t.Errorf("delivered %s: type=%s value=%v", n.Path, n.ContentType, n.Value)
					}
				case wire + "-case":
					seen[n.Path] = true
					if n.ContentType != "text/plain" || n.Value == nil || *n.Value != "case-secret" {
						t.Errorf("delivered %s: type=%s value=%v", n.Path, n.ContentType, n.Value)
					}
				}
			}
			if seen[wire] && seen[wire+"-case"] {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("encrypted parameters were not delivered (last status %d); delivered: %v", status, delivered)
		}
		time.Sleep(2 * time.Second)
	}
}

func TestSearchSkipsEncrypted(t *testing.T) {
	p := uniqPath("enc-search")
	token := fmt.Sprintf("needle%d", os.Getpid())
	mkEncrypted(t, p, "text/plain", token, "mk")
	paths := func() []string {
		var list []paramInfo
		if err := json.Unmarshal(adminReqOK(t, "GET", "/search?term="+token, nil), &list); err != nil {
			t.Fatal(err)
		}
		var ps []string
		for _, l := range list {
			ps = append(ps, l.Path)
		}
		return ps
	}
	if got := paths(); len(got) != 0 {
		t.Errorf("search found the encrypted value: %v", got)
	}
	if status, body := setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {token}, "encrypted": {"0"}}); status != http.StatusOK {
		t.Fatalf("declassify: status %d, body %s", status, body)
	}
	if got := paths(); len(got) != 1 || got[0] != p {
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
		`[{"group":"g","mime":"text/plain","encrypted":true,"value":"branch-secret"},{"mime":"text/plain","value":"shown"}]`, "mk case")
	if got := dbInt(t, countQuery); got != before+2 {
		t.Fatalf("notifyDB messages %d -> %d, want +2", before, got)
	}
	msgs := dbQuery(t, "SELECT Message FROM my_change_notification ORDER BY ID DESC LIMIT 2")
	for _, m := range msgs {
		if strings.Contains(m[0], "notified-secret") || strings.Contains(m[0], "branch-secret") {
			t.Errorf("notifyDB message leaks a secret: %s", m[0])
		}
		if !strings.Contains(m[0], "🔒") {
			t.Errorf("notifyDB message lacks the lock: %s", m[0])
		}
	}
	if !strings.Contains(msgs[0][0], "shown") {
		t.Errorf("plain branch not rendered: %s", msgs[0][0])
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
			if ntf.Value == nil || strings.Contains(*ntf.Value, "branch-secret") || !strings.Contains(*ntf.Value, "shown") {
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

func TestRestore(t *testing.T) {
	p := uniqPath("enc-restore")
	mkParam(t, p, "text/plain", "first", "mk")
	if status, body := setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"second-secret"}, "encrypted": {"1"}}); status != http.StatusOK {
		t.Fatalf("encrypt: status %d, body %s", status, body)
	}
	// restoring the plaintext version keeps the parameter encrypted
	if status, body := setParam(t, p, url.Values{"restore": {"1"}, "comment": {"rollback"}}); status != http.StatusOK {
		t.Fatalf("restore: status %d, body %s", status, body)
	}
	info := getParamInfo(t, p, "?reveal=1")
	if info.Version != 3 || !info.Encrypted || dataOf(info) != "first" {
		t.Errorf("restored parameter: %+v data=%s", info, dataOf(info))
	}
	if got := dbInt(t, "SELECT COUNT(*) FROM my_config_audit WHERE Path = "+sqlQuote(p)); got != 1 {
		t.Errorf("a restore counted as a reveal: %d audit rows", got)
	}
	// a deleted version cannot be restored
	q := uniqPath("enc-restore-del")
	mkParam(t, q, "text/plain", "x", "mk")
	deleteParam(t, q, "rm")
	mkParam(t, q, "text/plain", "y", "mk again")
	status, body := setParam(t, q, url.Values{"restore": {"2"}, "comment": {"rollback"}})
	expectError(t, status, body, http.StatusBadRequest, "InvalidValue")
	// a stale version is refused like any write
	status, body = adminReq(t, "POST", "/config"+p, url.Values{"version": {"1"}, "restore": {"1"}, "comment": {"stale"}})
	expectError(t, status, body, http.StatusBadRequest, "VersionNotMatch")
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
	status, body := adminReq(t, "POST", "/config"+uniqPath("enc-gate2"), url.Values{"mime": {"text/plain"}, "data": {"x"}, "encrypted": {"1"}, "comment": {"mk"}})
	expectError(t, status, body, http.StatusBadRequest, "EncryptionDisabled")
	// reads, reveals and declassification keep working: the switch is not a trap
	if dataOf(getParamInfo(t, p, "?reveal=1")) != "gated" {
		t.Errorf("reveal with the switch off failed")
	}
	if status, body := setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"freed"}, "encrypted": {"0"}}); status != http.StatusOK {
		t.Errorf("declassify with the switch off: status %d, body %s", status, body)
	}
}

func TestEncryptedTemplateBlocksDelete(t *testing.T) {
	target := uniqPath("enc-target")
	mkParam(t, target, "text/plain", "host", "mk target")
	referrer := uniqPath("enc-tmpl")
	mkEncrypted(t, referrer, "application/x-template", "user:pass@${"+target+"}", "mk template")
	status, body := adminReq(t, "DELETE", "/config"+target, url.Values{
		"version": {strconv.Itoa(paramVersion(t, target))},
		"comment": {"rm"},
	})
	expectError(t, status, body, http.StatusBadRequest, "DeletedTemplate")
	if !strings.Contains(string(body), referrer) {
		t.Errorf("the error does not name the encrypted referrer: %s", body)
	}
	// a template branch inside a case counts too
	caseRef := uniqPath("enc-tmpl-case")
	mkParam(t, caseRef, "application/x-case", `[{"mime":"application/x-template","encrypted":true,"value":"${`+target+`}"}]`, "mk case")
	deleteParam(t, referrer, "rm template")
	status, body = adminReq(t, "DELETE", "/config"+target, url.Values{
		"version": {strconv.Itoa(paramVersion(t, target))},
		"comment": {"rm"},
	})
	expectError(t, status, body, http.StatusBadRequest, "DeletedTemplate")
	deleteParam(t, caseRef, "rm case")
	deleteParam(t, target, "rm target")
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
		Current    bool   `json:"current"`
		Registered bool   `json:"registered"`
		Verified   bool   `json:"verified"`
		Recorded   bool   `json:"recorded"`
	} `json:"keys"`
	Sweep struct {
		State       string `json:"state"`
		Target      string `json:"target"`
		Full        bool   `json:"full"`
		Reencrypted int    `json:"reencrypted"`
	} `json:"sweep"`
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

func adminLogsContain(t *testing.T, needle string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		logs, _ := compose("logs", "--tail", "300", "onlineconf-admin")
		if strings.Contains(logs, needle) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(2 * time.Second)
	}
}

// TestKeyRotationSweep rotates the stack's encryption key the way an
// operator does: the new key joins the key file (a restart here, through
// docker compose, whatever manages the stack), then its id is written to
// /onlineconf/value-encryption/key-id and the sweep re-encrypts everything
// under it. The operator's mistakes change nothing, and the key file's
// order means nothing once a key is registered. The test leaves the stack
// with the rotated keyring — a consistent state every other test works
// against.
func TestKeyRotationSweep(t *testing.T) {
	p := uniqPath("enc-rotate")
	mkEncrypted(t, p, "text/plain", "rotated-secret", "mk")
	setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"rotated-secret"}, "encrypted": {"1"}})
	caseP := uniqPath("enc-rotate-case")
	mkParam(t, caseP, "application/x-case", `[{"mime":"text/plain","encrypted":true,"value":"rotated-branch"}]`, "mk")

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
	encryptedRows := func(table, keyID string) int {
		return dbInt(t, "SELECT COUNT(*) FROM "+table+" WHERE (Encrypted OR Value LIKE '%\"encrypted\":true%') AND Value LIKE '%oc1:"+keyID+":%'")
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
		if k.Current != (k.ID == oldID) || k.Recorded != (k.ID == oldID) {
			t.Errorf("key status after adding a retired key: %+v", k)
		}
	}

	// 2. the promotion: the operator writes the new key's id
	writeKeyID(t, newID)
	deadline := time.Now().Add(60 * time.Second)
	for {
		status = getEncryptionStatus(t)
		if status.Sweep.State == "done" && status.Sweep.Full && status.Sweep.Target == newID {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("rotation sweep did not complete: %+v", status.Sweep)
		}
		time.Sleep(time.Second)
	}
	if status.Sweep.Reencrypted == 0 {
		t.Errorf("the sweep re-encrypted nothing: %+v", status.Sweep)
	}
	if status.Wanted.ID != newID || status.Wanted.Source != "parameter" || !status.Wanted.Held {
		t.Errorf("wanted key after the promotion = %+v", status.Wanted)
	}
	for _, k := range status.Keys {
		if !k.Registered || !k.Verified || k.Current != (k.ID == newID) || k.Recorded != (k.ID == newID) {
			t.Errorf("key status after rotation: %+v", k)
		}
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
	if revealed := dataOf(getParamInfo(t, caseP, "?reveal=1")); !strings.Contains(revealed, "rotated-branch") {
		t.Errorf("case branch lost in rotation: %s", revealed)
	}
	// a write after the promotion lands under the new key
	setParam(t, p, url.Values{"mime": {"text/plain"}, "data": {"rotated-again"}, "encrypted": {"1"}})
	if got := dbQuery(t, "SELECT Value FROM my_config_tree WHERE Path = "+sqlQuote(p)); len(got) != 1 || !strings.HasPrefix(got[0][0], "oc1:"+newID+":") {
		t.Errorf("a write after the promotion is not under the new key: %v", got)
	}

	// 3. the operator's mistakes change nothing: a key this instance does
	// not hold, and the key that was active before
	since := time.Now()
	writeKeyID(t, "nokey"+newID)
	if !adminLogsSince(t, since, "not in this instance's key file", 20*time.Second) {
		t.Errorf("an unknown key id was not reported")
	}
	since = time.Now()
	writeKeyID(t, oldID)
	if !adminLogsSince(t, since, "rotate forward", 20*time.Second) {
		t.Errorf("the old key id was not refused")
	}
	if got := registeredActive(); got != newID {
		t.Errorf("a mistaken key id changed the registered active key to %s", got)
	}
	writeKeyID(t, newID)

	// 4. the key file's order means nothing once a key is registered: the
	// old key first, a restart, and the new key stays the one in use
	installKeyring(t, append(append([]string{}, current...), newLine)...)
	if err := waitReady(2 * time.Minute); err != nil {
		t.Fatal(err)
	}
	status = getEncryptionStatus(t)
	for _, k := range status.Keys {
		if k.Current != (k.ID == newID) || k.Recorded != (k.ID == newID) {
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

	// back to the rotated state, which later tests run against
	installKeyring(t, append([]string{newLine}, current...)...)
	if err := waitReady(2 * time.Minute); err != nil {
		t.Fatal(err)
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
