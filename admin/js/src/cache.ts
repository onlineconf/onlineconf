import * as API from './api';

type CacheKey = 'group' | 'datacenter' | 'service';

const cache: { [K in CacheKey]?: { [k: string]: string } } = {};

const cacheDepth: { [K in CacheKey]: 'children' | 'subtree' } = {
	group: 'children',
	datacenter: 'children',
	service: 'subtree',
};

function fillDictionary(result: { [k: string]: string }, param: API.IParam, namePrefix: string, summaryPrefix: string, subtree: boolean) {
	if (!param.children) {
		return;
	}
	for (const node of param.children) {
		const name = namePrefix + node.name;
		const summary = summaryPrefix + (node.summary || node.name);
		result[name] = summary;
		if (subtree) {
			fillDictionary(result, node, name + '/', summary + ' / ', true);
		}
	}
}

async function getDict(key: CacheKey) {
	if (key in cache) {
		return cache[key];
	}

	const depth = cacheDepth[key];
	const data = await API.getParam(`/onlineconf/${key}`, 'resolve', depth);
	const result = {};
	fillDictionary(result, data, '', '', depth === 'subtree');

	cache[key] = result;
	setTimeout(() => delete(cache[key]), 60000);

	return result;
}

const promise = {};

export async function getDictionary(key: CacheKey) {
	if (!(key in promise)) {
		promise[key] = getDict(key);
	}
	const result = await promise[key];
	delete(promise[key]);
	return result;
}
