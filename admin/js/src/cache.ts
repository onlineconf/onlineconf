import * as API from './api';

export const dictionaryKeys = ['group', 'datacenter', 'service'] as const;
export type DictionaryKey = typeof dictionaryKeys[number];

const cache: { [K in DictionaryKey]?: { [k: string]: string } } = {};

const cacheDepth: { [K in DictionaryKey]: 'children' | 'subtree' } = {
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

async function getDict(key: DictionaryKey) {
	if (key in cache) {
		return cache[key]!;
	}

	const depth = cacheDepth[key];
	const data = await API.getParam(`/onlineconf/${key}`, 'resolve', depth);
	const result = {};
	fillDictionary(result, data, '', '', depth === 'subtree');

	cache[key] = result;
	setTimeout(() => delete(cache[key]), 60000);

	return result;
}

const promise: { [K in DictionaryKey]?: Promise<{ [k: string]: string }> } = {};

export async function getDictionary(key: DictionaryKey) {
	if (!(key in promise)) {
		promise[key] = getDict(key);
	}
	const result = await promise[key];
	delete(promise[key]);
	return result;
}
