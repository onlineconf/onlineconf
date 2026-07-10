//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

// tryDelete attempts to delete a live parameter and returns the status code
// and body without failing the test, so callers can assert refusals.
func tryDelete(t *testing.T, path string) (int, string) {
	t.Helper()
	version := paramVersion(t, path)
	status, body := adminReq(t, "DELETE", "/config"+path, url.Values{
		"version": {strconv.Itoa(version)},
		"comment": {"delete attempt"},
	})
	return status, string(body)
}

// requireBlockedDelete asserts that deleting path is refused with the given
// error code and that the refusal names the expected referrer.
func requireBlockedDelete(t *testing.T, path, errorCode, referrer string) {
	t.Helper()
	status, body := tryDelete(t, path)
	if status != http.StatusBadRequest || !strings.Contains(body, `"error":"`+errorCode+`"`) {
		t.Fatalf("DELETE %s: status %d body %s, want 400 %s", path, status, body, errorCode)
	}
	if !strings.Contains(body, referrer) {
		t.Errorf("DELETE %s refusal does not name the referrer %s: %s", path, referrer, body)
	}
}

// TestDeleteReferencedBySymlink checks the basic referential guarantee: a
// parameter referenced by a live symlink cannot be deleted until the symlink
// is gone.
func TestDeleteReferencedBySymlink(t *testing.T) {
	target, link := uniqPath("dep-tgt"), uniqPath("dep-ln")
	mkParam(t, target, "text/plain", "v", "mk target")
	mkParam(t, link, "application/x-symlink", target, "mk symlink")

	requireBlockedDelete(t, target, "DeletedSymlink", link)

	deleteParam(t, link, "drop symlink")
	deleteParam(t, target, "now deletable")
}

// TestDeleteReferencedThroughChain checks indirect references: a symlink
// resolved through another symlink blocks deletion of both the intermediate
// hop and the terminal node.
func TestDeleteReferencedThroughChain(t *testing.T) {
	dir := uniqPath("dep-dir")
	mkParam(t, dir, "application/x-null", "", "mk dir")
	mkParam(t, dir+"/c", "text/plain", "v", "mk leaf")
	x, y := uniqPath("dep-x"), uniqPath("dep-y")
	mkParam(t, x, "application/x-symlink", dir, "mk x")
	mkParam(t, y, "application/x-symlink", x+"/c", "mk y")

	requireBlockedDelete(t, x, "DeletedSymlink", y)        // y resolves through x
	requireBlockedDelete(t, dir+"/c", "DeletedSymlink", y) // y terminates on dir/c

	deleteParam(t, y, "drop y")
	deleteParam(t, x, "x is unreferenced now")
	deleteParam(t, dir+"/c", "leaf is unreferenced now")
	deleteParam(t, dir, "cleanup")
}

// TestTemplateThroughSymlinkBlocksHop checks the deliberate tightening over
// the legacy check: a template referencing ${<symlink>/sub} blocks deletion
// of the symlink it resolves through, and of the terminal node.
func TestTemplateThroughSymlinkBlocksHop(t *testing.T) {
	a := uniqPath("tmpl-a")
	mkParam(t, a, "application/x-null", "", "mk a")
	mkParam(t, a+"/b", "text/plain", "v", "mk b")
	x, tmpl := uniqPath("tmpl-x"), uniqPath("tmpl-t")
	mkParam(t, x, "application/x-symlink", a, "mk x")
	mkParam(t, tmpl, "application/x-template", "${"+x+"/b}", "mk template")

	requireBlockedDelete(t, x, "DeletedTemplate", tmpl)
	requireBlockedDelete(t, a+"/b", "DeletedTemplate", tmpl)

	deleteParam(t, tmpl, "drop template")
	deleteParam(t, x, "cleanup")
	deleteParam(t, a+"/b", "cleanup")
	deleteParam(t, a, "cleanup")
}

// TestCaseReferrerBlocks checks that a symlink inside a case branch protects
// its target like a plain symlink does.
func TestCaseReferrerBlocks(t *testing.T) {
	target, cs := uniqPath("case-tgt"), uniqPath("case-ref")
	mkParam(t, target, "text/plain", "v", "mk target")
	mkParam(t, cs, "application/x-case",
		`[{"mime":"application/x-symlink","value":"`+target+`","group":"some-group"}]`, "mk case")

	requireBlockedDelete(t, target, "DeletedSymlink", cs)

	deleteParam(t, cs, "drop case")
	deleteParam(t, target, "cleanup")
}

// TestSelfReferencingSymlinks checks symlinks pointing at their own parent and
// at the root: they must protect the target without ever deadlocking their own
// deletion.
func TestSelfReferencingSymlinks(t *testing.T) {
	p := uniqPath("self-p")
	mkParam(t, p, "application/x-null", "", "mk parent")
	mkParam(t, p+"/ln", "application/x-symlink", p, "symlink to own parent")

	// the parent is protected (by the child link and by non-emptiness)
	if status, body := tryDelete(t, p); status != http.StatusBadRequest {
		t.Fatalf("DELETE %s: status %d body %s, want 400", p, status, body)
	}
	// the symlink itself is freely deletable, then the parent follows
	deleteParam(t, p+"/ln", "drop self link")
	deleteParam(t, p, "now deletable")

	// a symlink to the root resolves and records its edge
	rx := uniqPath("self-root")
	mkParam(t, rx, "application/x-symlink", "/", "symlink to root")
	edges := dbInt(t, "SELECT COUNT(*) FROM my_config_tree_dep d JOIN my_config_tree t ON t.ID = d.ReferrerID WHERE t.Path = '"+rx+"'")
	if edges != 1 {
		t.Errorf("symlink to root: got %d dependency edges, want 1 (the root)", edges)
	}
	deleteParam(t, rx, "cleanup")
}

// TestNestedCaseReferrer checks that a symlink buried in a nested case still
// protects its target.
func TestNestedCaseReferrer(t *testing.T) {
	target, cs := uniqPath("ncase-tgt"), uniqPath("ncase-ref")
	mkParam(t, target, "text/plain", "v", "mk target")
	inner := `[{\"mime\":\"application/x-symlink\",\"value\":\"` + target + `\",\"group\":\"inner-group\"}]`
	mkParam(t, cs, "application/x-case",
		`[{"mime":"application/x-case","value":"`+inner+`","group":"outer-group"}]`, "mk nested case")

	requireBlockedDelete(t, target, "DeletedSymlink", cs)

	deleteParam(t, cs, "drop case")
	deleteParam(t, target, "cleanup")
}

// TestHealNestedCaseTemplate checks healing through nesting: a template
// branch inside a nested case referencing a not-yet-existing path starts
// protecting it as soon as the path is created.
func TestHealNestedCaseTemplate(t *testing.T) {
	q, cs := uniqPath("nheal-q"), uniqPath("nheal-ref")
	inner := `[{\"mime\":\"application/x-template\",\"value\":\"${` + q + `}\",\"group\":\"inner-group\"}]`
	mkParam(t, cs, "application/x-case",
		`[{"mime":"application/x-case","value":"`+inner+`","group":"outer-group"}]`, "mk nested case")
	mkParam(t, q, "text/plain", "v", "mk the referenced parameter")

	requireBlockedDelete(t, q, "DeletedSymlink", cs)

	deleteParam(t, cs, "drop case")
	deleteParam(t, q, "cleanup")
}

// TestMoveReferencedRequiresSymlink checks that moving a referenced parameter
// is refused unless a symlink is left behind: without the symlink the old path
// vanishes and the referrer would dangle (the same breakage deletion refuses);
// with the symlink the move and the symlink creation are one atomic operation
// and the referrer keeps resolving.
func TestMoveReferencedRequiresSymlink(t *testing.T) {
	target, dst, link := uniqPath("mv-tgt"), uniqPath("mv-dst"), uniqPath("mv-ln")
	mkParam(t, target, "text/plain", "v", "mk target")
	mkParam(t, link, "application/x-symlink", target, "mk referrer")

	// move away WITHOUT a symlink -> refused (link would dangle)
	status, body := adminReq(t, "POST", "/config"+target, url.Values{
		"version": {strconv.Itoa(paramVersion(t, target))},
		"path":    {dst},
		"comment": {"move without symlink"},
	})
	if status != http.StatusBadRequest || !strings.Contains(string(body), `"error":"DeletedSymlink"`) {
		t.Fatalf("move of referenced %s without symlink: status %d body %s, want 400 DeletedSymlink", target, status, body)
	}
	if !strings.Contains(string(body), link) {
		t.Errorf("refusal does not name the referrer %s: %s", link, body)
	}
	if s, _ := adminReq(t, "GET", "/config"+dst, nil); s != http.StatusNotFound {
		t.Errorf("destination %s exists after a refused move (status %d)", dst, s)
	}

	// move WITH a symlink -> allowed and atomic: the destination exists AND a
	// symlink pointing at it remains at the old path, in one operation
	moveParamSymlink(t, target, dst, "move with symlink")
	if s, _ := adminReq(t, "GET", "/config"+dst, nil); s != http.StatusOK {
		t.Fatalf("destination %s missing after move-with-symlink (status %d)", dst, s)
	}
	var left struct {
		Mime string `json:"mime"`
		Data string `json:"data"`
	}
	if err := json.Unmarshal(adminReqOK(t, "GET", "/config"+target, nil), &left); err != nil {
		t.Fatal(err)
	}
	if left.Mime != "application/x-symlink" || left.Data != dst {
		t.Errorf("leftover at old path = %+v, want a symlink to %s", left, dst)
	}
	// the chain link -> target -> dst still resolves: the leftover symlink at the
	// old path now protects dst directly, and link in turn protects that symlink
	requireBlockedDelete(t, dst, "DeletedSymlink", target)
	requireBlockedDelete(t, target, "DeletedSymlink", link)

	deleteParam(t, link, "drop referrer")
	deleteParam(t, target, "drop leftover symlink")
	deleteParam(t, dst, "now unreferenced")
}

// TestMoveReferencedSubtreeRefused checks that the refusal covers the whole
// moved subtree: a referrer to a descendant blocks a no-symlink move of an
// ancestor.
func TestMoveReferencedSubtreeRefused(t *testing.T) {
	root, dst, link := uniqPath("mvs-root"), uniqPath("mvs-dst"), uniqPath("mvs-ln")
	mkParam(t, root, "application/x-null", "", "mk root")
	mkParam(t, root+"/child", "text/plain", "v", "mk child")
	mkParam(t, link, "application/x-symlink", root+"/child", "referrer to a descendant")

	status, body := adminReq(t, "POST", "/config"+root, url.Values{
		"version": {strconv.Itoa(paramVersion(t, root))},
		"path":    {dst},
		"comment": {"move subtree"},
	})
	if status != http.StatusBadRequest || !strings.Contains(string(body), link) {
		t.Fatalf("move of subtree with a referenced descendant: status %d body %s, want 400 naming %s", status, body, link)
	}

	deleteParam(t, link, "drop referrer")
	deleteParam(t, root+"/child", "cleanup")
	deleteParam(t, root, "cleanup")
}

// TestHealDanglingTemplate checks that a template referencing a
// not-yet-existing path starts protecting it as soon as the path is created.
func TestHealDanglingTemplate(t *testing.T) {
	q, tmpl := uniqPath("heal-q"), uniqPath("heal-t")
	mkParam(t, tmpl, "application/x-template", "${"+q+"/w}", "dangling template")
	mkParam(t, q, "application/x-null", "", "mk q")
	mkParam(t, q+"/w", "text/plain", "v", "mk w")

	requireBlockedDelete(t, q+"/w", "DeletedTemplate", tmpl)

	deleteParam(t, tmpl, "drop template")
	deleteParam(t, q+"/w", "cleanup")
	deleteParam(t, q, "cleanup")
}

// TestBackfillProtectsExampleData checks the startup backfill: the example
// database predates the dependency table, so its referrers are only protected
// if the admin server populated the table on boot. The backfill retries in
// the background until the database is reachable, so wait for it to land
// before asserting (probing with the delete itself would destroy the node on
// a premature success).
func TestBackfillProtectsExampleData(t *testing.T) {
	deadline := time.Now().Add(60 * time.Second)
	for dbInt(t, "SELECT COUNT(*) FROM my_config_tree_dep") == 0 {
		if time.Now().After(deadline) {
			t.Fatal("the startup backfill did not populate my_config_tree_dep")
		}
		time.Sleep(time.Second)
	}
	// /gopher/statistics/graphite is a template interpolating
	// ${/infrastructure/graphite/host} in the example data
	requireBlockedDelete(t, "/infrastructure/graphite/host", "DeletedTemplate", "/gopher/statistics/graphite")
}

// TestDeepChainTrackedAtDefaultDepth checks that the default symlink depth (10)
// tracks dependencies through a chain of stacked directory symlinks deeper than
// the previous default of 5. The terminal /base/leaf is reachable only by the
// referrer that resolves all the way through the chain, so it is protected only
// if the walk followed every hop.
func TestDeepChainTrackedAtDefaultDepth(t *testing.T) {
	b := uniqPath("deep")
	mkParam(t, b, "application/x-null", "", "mk base dir")
	mkParam(t, b+"/base", "application/x-null", "", "mk base")
	mkParam(t, b+"/base/leaf", "text/plain", "v", "mk leaf")

	// s1 -> base, s2 -> s1, ... s6 -> s5: resolving sN/leaf traverses N
	// directory-symlink hops (6 > old default of 5) down to base/leaf
	prev := b + "/base"
	for i := 1; i <= 6; i++ {
		s := fmt.Sprintf("%s/s%d", b, i)
		mkParam(t, s, "application/x-symlink", prev, "mk hop")
		prev = s
	}
	ref := b + "/ref"
	mkParam(t, ref, "application/x-symlink", b+"/s6/leaf", "mk ref")

	// the terminal is protected only because the walk reached depth 6
	requireBlockedDelete(t, b+"/base/leaf", "DeletedSymlink", ref)

	deleteParam(t, ref, "drop ref")
	deleteParam(t, b+"/base/leaf", "now deletable")
}

// TestDepthTunableViaParameter checks that the maximum symlink depth is
// controlled by /onlineconf/deleted-key-symlinks-check-depth and that a change
// takes effect without a restart. The rebuild runs in the background (so a high
// depth cannot freeze the request that set it), so the recorded depth in
// my_config_tree_dep_meta is observed to rise after a raise; a lower triggers no
// rebuild, leaving the existing deeper edges (the safe direction).
func TestDepthTunableViaParameter(t *testing.T) {
	const depthParam = "/onlineconf/deleted-key-symlinks-check-depth"
	t.Cleanup(func() {
		if status, _ := adminReq(t, "GET", "/config"+depthParam, nil); status == http.StatusOK {
			deleteParam(t, depthParam, "restore default depth")
		}
	})

	metaDepth := func() int { return dbInt(t, "SELECT Depth FROM my_config_tree_dep_meta LIMIT 1") }
	metaRows := func() int { return dbInt(t, "SELECT COUNT(*) FROM my_config_tree_dep_meta") }
	waitMetaDepth := func(want int) {
		t.Helper()
		deadline := time.Now().Add(60 * time.Second)
		for {
			if metaRows() > 0 && metaDepth() == want {
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("dependency rebuild did not reach depth %d (meta = %v rows)", want, metaRows())
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
	setDepth := func(v int) {
		s := strconv.Itoa(v)
		if status, _ := adminReq(t, "GET", "/config"+depthParam, nil); status == http.StatusOK {
			editParam(t, depthParam, "text/plain", s, "set depth")
		} else {
			mkParam(t, depthParam, "text/plain", s, "set depth")
		}
	}

	// wait for the startup backfill to record a depth (default 10, unless a
	// prior test already raised it)
	deadline := time.Now().Add(60 * time.Second)
	for metaRows() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("the startup backfill did not record a depth in my_config_tree_dep_meta")
		}
		time.Sleep(200 * time.Millisecond)
	}
	base := metaDepth()

	setDepth(base + 5)      // raise -> background rebuild
	waitMetaDepth(base + 5) // ... eventually recorded

	setDepth(base + 2)          // lower -> no rebuild
	time.Sleep(2 * time.Second) // give any (erroneous) rebuild time to run
	if got := metaDepth(); got != base+5 {
		t.Errorf("after lowering depth, table depth = %d, want unchanged %d (lower must not rebuild)", got, base+5)
	}
}

// TestHealAfterDisabledDeleteRecreate checks the full flag-era repair path: a
// target subtree destroyed with the check disabled and later recreated must be
// protected again — including for a referrer that resolves through an
// intermediate symlink and never textually mentions the recreated paths (the
// dangling-point tracking plus the one-level dependent recompute in heal).
func TestHealAfterDisabledDeleteRecreate(t *testing.T) {
	const flag = "/onlineconf/disable-deleted-key-symlinks-check"
	t.Cleanup(func() {
		if status, _ := adminReq(t, "GET", "/config"+flag, nil); status == http.StatusOK {
			deleteParam(t, flag, "restore the check")
		}
	})

	g, link, b := uniqPath("heal-gone"), uniqPath("heal-link"), uniqPath("heal-b")
	mkParam(t, g, "application/x-null", "", "mk g")
	mkParam(t, g+"/sub", "application/x-null", "", "mk sub")
	mkParam(t, g+"/sub/c", "text/plain", "v", "mk c")
	mkParam(t, link, "application/x-symlink", g+"/sub", "mk link")
	mkParam(t, b, "application/x-symlink", link+"/c", "mk b (resolves through link)")

	// destroy the target subtree with the check disabled
	mkParam(t, flag, "text/plain", "1", "disable the check")
	deleteParam(t, g+"/sub/c", "flag-era delete")
	deleteParam(t, g+"/sub", "flag-era delete")
	deleteParam(t, g, "flag-era delete")
	deleteParam(t, flag, "re-enable the check")

	// recreate the subtree top-down: healing must re-attach link (textual
	// mention) and b (dangling point + dependent of link)
	mkParam(t, g, "application/x-null", "", "recreate g")
	mkParam(t, g+"/sub", "application/x-null", "", "recreate sub")
	mkParam(t, g+"/sub/c", "text/plain", "v2", "recreate c")

	// the deep referrer b protects the recreated leaf even though nothing
	// textually mentions it
	requireBlockedDelete(t, g+"/sub/c", "DeletedSymlink", b)

	deleteParam(t, b, "drop deep referrer")
	deleteParam(t, link, "drop link")
	deleteParam(t, g+"/sub/c", "now deletable")
	deleteParam(t, g+"/sub", "cleanup")
	deleteParam(t, g, "cleanup")
}

// TestFailClosedBeforeBackfill checks the first-boot window: while the
// dependency table has never been built (no recorded depth), deletions are
// refused with a retriable error instead of passing unchecked.
func TestFailClosedBeforeBackfill(t *testing.T) {
	p := uniqPath("notready")
	mkParam(t, p, "text/plain", "v", "mk")

	saved := dbQuery(t, "SELECT Depth FROM my_config_tree_dep_meta")
	dbQuery(t, "DELETE FROM my_config_tree_dep_meta")
	status, body := tryDelete(t, p)
	if len(saved) == 1 && len(saved[0]) == 1 { // restore before asserting
		dbQuery(t, "INSERT INTO my_config_tree_dep_meta (Depth) VALUES ("+saved[0][0]+")")
	}

	if status != http.StatusServiceUnavailable || !strings.Contains(body, `"error":"DepsNotReady"`) {
		t.Fatalf("DELETE %s before backfill: status %d body %s, want 503 DepsNotReady", p, status, body)
	}
	deleteParam(t, p, "cleanup (dependency data restored)")
}

// TestDepthTruncationLogged checks that a dependency walk abandoned at the
// depth cap is logged: protection narrowing must be visible, not silent.
func TestDepthTruncationLogged(t *testing.T) {
	base := uniqPath("trunc")
	mkParam(t, base, "application/x-null", "", "mk base")
	mkParam(t, base+"/target", "application/x-null", "", "mk target")
	mkParam(t, base+"/target/leaf", "text/plain", "v", "mk leaf")
	prev := base + "/target"
	const hops = 11 // one more than the default depth of 10
	for i := 1; i <= hops; i++ {
		s := fmt.Sprintf("%s/s%d", base, i)
		mkParam(t, s, "application/x-symlink", prev, "mk hop")
		prev = s
	}
	ref := base + "/ref"
	mkParam(t, ref, "application/x-symlink", fmt.Sprintf("%s/s%d/leaf", base, hops),
		"referrer through too many hops")

	logs, err := compose("logs", "onlineconf-admin")
	if err != nil {
		t.Fatalf("reading admin logs: %v\n%s", err, logs)
	}
	if !strings.Contains(logs, "dependency resolution truncated") {
		t.Error("truncation warning not found in the admin server logs")
	}

	deleteParam(t, ref, "drop referrer")
	for i := hops; i >= 1; i-- {
		deleteParam(t, fmt.Sprintf("%s/s%d", base, i), "drop hop")
	}
	deleteParam(t, base+"/target/leaf", "cleanup")
	deleteParam(t, base+"/target", "cleanup")
	deleteParam(t, base, "cleanup")
}

// TestDisableFlagSkipsCheck checks the /onlineconf/disable-deleted-key-symlinks-check
// escape hatch: with the flag set, referenced parameters are deletable. Keep
// this test last in the file — it temporarily disables the protection.
func TestDisableFlagSkipsCheck(t *testing.T) {
	const flag = "/onlineconf/disable-deleted-key-symlinks-check"
	if status, _ := adminReq(t, "GET", "/config"+flag, nil); status == http.StatusOK {
		editParam(t, flag, "text/plain", "1", "enable")
	} else {
		mkParam(t, flag, "text/plain", "1", "enable")
	}
	t.Cleanup(func() {
		if status, _ := adminReq(t, "GET", "/config"+flag, nil); status == http.StatusOK {
			deleteParam(t, flag, "restore the check")
		}
	})

	target, link := uniqPath("flag-tgt"), uniqPath("flag-ln")
	mkParam(t, target, "text/plain", "v", "mk target")
	mkParam(t, link, "application/x-symlink", target, "mk symlink")

	if status, body := tryDelete(t, target); status != http.StatusOK {
		t.Fatalf("DELETE %s with the check disabled: status %d body %s, want 200", target, status, body)
	}
	deleteParam(t, link, "drop the now-dangling symlink")
}
