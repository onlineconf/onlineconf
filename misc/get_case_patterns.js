const caseType = "application/x-case";
const fileName = "server_globs.json";

function getNodes(paths) {
	var params = new URLSearchParams();
	for (path of paths)
		params.append("id[]", path);

	var req = new XMLHttpRequest();
	req.open("POST", "https://onlineconf.my.cloud.devmail.ru/batch/GET/config", false);
	req.setRequestHeader("X-Requested-With", "XMLHttpRequest");
	req.send(params);
	return JSON.parse(req.response);
}

function processCase(node, serverGlobs) {
	for (var c of JSON.parse(node.data))
		if (c.server)
			serverGlobs.add(c.server);
}

function process(paths, serverGlobs) {
	console.log("processing %d paths", paths.length);

	var nodes = getNodes(paths);
	paths = [];

	for (var [path, node] of Object.entries(nodes)) {
		if (node.mime == caseType)
			processCase(node, serverGlobs);

		for (var child of node.children)
			if (child.num_children)
				paths.push(child.path);
			else if (child.mime == caseType)
				processCase(child, serverGlobs);
	}

	if (paths.length)
		process(paths, serverGlobs);
}

function saveGlobs() {
	var serverGlobs = new Set();
	process("/", serverGlobs);

	var output = new Object();
	output.Servers = Array.from(serverGlobs);

	var file = document.createElement("a");
	file.href = "data:application/octet-stream," + encodeURIComponent(JSON.stringify(output));
	file.download = fileName;
	file.click();

	console.log(fileName, "saved");
}

saveGlobs();
