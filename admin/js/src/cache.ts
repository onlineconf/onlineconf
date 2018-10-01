import * as API from './api';

type CacheKey = 'group' | 'datacenter';

const cache = {};

async function getDict(key: CacheKey) {
	if (key in cache) {
		return cache[key];
	}

	const data = await API.getParam(`/onlineconf/${key}`, 'resolve');
	const result = {};
	if (data.children) {
		for (const node of data.children) {
			result[node.name] = node.summary || node.name;
		}
	}

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
