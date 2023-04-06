const caseType = "application/x-case";
const symlinkType = "application/x-symlink";

function getNodes(paths) {
	var params = new URLSearchParams();
	for (path of paths)
		params.append("id[]", path);

	var req = new XMLHttpRequest();
	req.open("POST", "/batch/GET/config", false);
	req.setRequestHeader("X-Requested-With", "XMLHttpRequest");
	req.send(params);
	return JSON.parse(req.response);
}

function processCase(node, serverGlobs) {
	if (node.data)
		for (var c of JSON.parse(node.data))
			if (c.server)
				serverGlobs.add(c.server);
}

function traverseCases(paths, serverGlobs) {
	console.log("cases: processing %d paths", paths.length);

	var nodes = getNodes(paths);
	paths = [];

	for (var [_, node] of Object.entries(nodes)) {
		if (node.mime == caseType)
			processCase(node, serverGlobs);

		for (var child of node.children)
			if (child.num_children)
				paths.push(child.path);
			else if (child.mime == caseType)
				processCase(child, serverGlobs);
	}

	if (paths.length)
		traverseCases(paths, serverGlobs);
}

function traverseGroups(paths, serverGlobs) {
	console.log("groups: processing %d paths", paths.length);

	var nodes = getNodes(paths);
	paths = [];

	for (var [_, node] of Object.entries(nodes)) {
		if (node.mime == symlinkType)
			paths.push(node.data);
		else if (node.data)
			serverGlobs.add(node.data);

		for (var child of node.children)
			if (child.num_children)
				paths.push(child.path);
			else if (child.mime == symlinkType)
				paths.push(child.data);
			else if (child.data)
				serverGlobs.add(child.data);
	}

	if (paths.length)
		traverseGroups(paths, serverGlobs);
}

function getServers() {
	var req = new XMLHttpRequest();
	req.open("GET", "/monitoring", false);
	req.setRequestHeader("X-Requested-With", "XMLHttpRequest");
	req.send();
	return JSON.parse(req.response).map(function(srv) { return srv.host });
}

function saveGlobs(fileName) {
	var started = Date.now();

	var serverGlobs = new Set();
	traverseCases(["/"], serverGlobs);

	var groupGlobs = new Set();
	traverseGroups(["/onlineconf/group"], groupGlobs);

	var output = new Object();
	output.server_globs = Array.from(serverGlobs);
	output.group_globs = Array.from(groupGlobs);
	output.servers = getServers();

	var file = document.createElement("a");
	file.href = "data:application/octet-stream," + encodeURIComponent(JSON.stringify(output));
	file.download = fileName;
	file.click();

	var duration = (Date.now() - started) / 1000;
	console.log("got %d server globs in %d seconds.\n%s saved.", serverGlobs.size, duration, fileName);
}

saveGlobs("server_globs.json");
