import axios, { AxiosRequestConfig } from 'axios';
import * as qs from 'querystring';
import { Omit } from '@material-ui/types';

export type ParamType = 'application/x-null'
	| 'text/plain'
	| 'application/json'
	| 'application/x-yaml'
	| 'application/x-symlink'
	| 'application/x-case'
	| 'application/x-template'
	| 'application/x-list'
	| 'application/x-server'
	| 'application/x-server2';

export type Notification = 'none' | 'no-value' | 'with-value';

interface ParamFields {
	path: string;
	name: string;
	mime: ParamType;
	data: string | null;
	summary: string;
	description: string;
	version: number;
	mtime: string;
	rw: boolean | null;
	notification: Notification;
}

export interface IParam extends ParamFields {
	access_modified: boolean;
	notification_modified: boolean;
	num_children: number;
	children?: IParam[];
}

export interface ParamModify extends Partial<Omit<ParamFields, 'notification'>> {
	notification?: Notification | null;
	symlink?: boolean;
	comment?: string;
}

interface ParamParams extends Omit<ParamModify, 'symlink'> {
	symlink?: 1;
}

const commonOptions = { headers: { 'X-Requested-With': 'XMLHttpRequest' } };
const commonUrlencodedOptions = {
	...commonOptions,
	headers: { ...commonOptions.headers, 'Content-Type': 'application/x-www-form-urlencoded' },
};

export async function getParam(path: string, symlink?: 'resolve' | 'follow', depth?: 'children' | 'subtree', options: AxiosRequestConfig = {}) {
	const params = { symlink, depth };
	const response = await axios.get<IParam>('/config' + path, { ...options, ...commonOptions, params });
	return response.data;
}

export async function postParam(path: string, modify: ParamModify) {
	const { symlink, ...rest } = modify;
	const params: ParamParams = rest;
	if (symlink) {
		params.symlink = 1;
	}
	const data = qs.stringify(params);
	const response = await axios.post<IParam>('/config' + path, data, commonUrlencodedOptions);
	return response.data;
}

export async function deleteParam(path: string, info: { version: number, comment: string }) {
	const response = await axios.request<void>({
		...commonUrlencodedOptions,
		url: '/config' + path,
		method: 'DELETE',
		data: qs.stringify(info),
	});
	return response.data;
}

export async function batchGetParams(paths: string[]) {
	const data = new URLSearchParams();
	for (const path of paths) {
		data.append('id[]', path);
	}
	const response = await axios.post<{[key: string]: IParam}>('/batch/GET/config', data, commonOptions);
	return response.data;
}

export async function getParams(paths: Set<string>) {
	if (paths.size === 0) {
		return {};
	} else if (paths.size === 1) {
		const path: string = paths.values().next().value;
		return { [path]: await getParam(path) };
	} else {
		return await batchGetParams(Array.from(paths));
	}
}

export async function search(term: string) {
	const response = await axios.get<IParam[]>('/search', { ...commonOptions, params: { term } });
	return response.data;
}


export const logLimit = 50;

interface LogPagination {
	lastid?: number;
}

export interface IParamLog {
	id: number;
	path: string;
	version: number;
	mime: ParamType;
	data: string;
	mtime: string;
	author: string;
	comment: string;
	deleted: boolean;
	rw: boolean | null;
	same: boolean;
}

export async function getParamLog(path: string, lastID?: number, options: AxiosRequestConfig = {}) {
	const params: LogPagination = { lastid: lastID };
	const response = await axios.get<IParamLog[]>('/log' + path, { ...options, ...commonOptions, params });
	return response.data;
}

export interface GlobalLogFilter {
	author?: string;
	branch?: string;
	from?: string;
	till?: string;
	all?: boolean;
}

interface GlobalLogParams extends Omit<GlobalLogFilter, 'all'>, LogPagination {
	all?: 1;
}

export async function getGlobalLog(filter: GlobalLogFilter, lastID?: number, options: AxiosRequestConfig = {}) {
	const params: GlobalLogParams = {};
	for (const k of ['author', 'branch', 'from', 'till'] as const) {
		if (filter[k] !== undefined && filter[k] !== '') {
			params[k] = filter[k];
		}
	}
	if (filter.all) {
		params.all = 1;
	}
	params.lastid = lastID;
	const response = await axios.get<IParamLog[]>('/global-log', { ...options, ...commonOptions, params });
	return response.data;
}


export interface Server {
	host: string;
	mtime: string | null;
	mtime_alert: boolean;
	online: string;
	online_alert: boolean;
	package: string;
}

export async function getServers() {
	const response = await axios.get<Server[]>('/monitoring', commonOptions);
	return response.data;
}

export async function deleteServer(host: string) {
	const response = await axios.delete('/monitoring/' + host, commonOptions);
	return response.data;
}


export async function getAccess() {
	const access: { [key: string]: string[] } = {};
	const response = await axios.get<string[]>('/group/', commonOptions);
	const promises = response.data.map(group => {
		return axios.get<string[]>(`/group/${group}`, commonOptions).then(resp => {
			access[group] = resp.data;
		});
	});
	await Promise.all(promises);
	return access;
}

export async function createGroup(group: string) {
	const response = await axios.post('/group/' + group, undefined, commonOptions);
	return response.data;
}

export async function deleteGroup(group: string) {
	const response = await axios.delete('/group/' + group, commonOptions);
	return response.data;
}

export async function addUser(group: string, user: string) {
	const response = await axios.post('/group/' + group + '/' + user, undefined, commonOptions);
	return response.data;
}

export async function removeUser(group: string, user: string) {
	const response = await axios.delete('/group/' + group + '/' + user, commonOptions);
	return response.data;
}


export interface ParamAccessGroup {
	group: string;
	overridden: boolean;
	rw: boolean | null;
}

export async function getParamAccess(path: string, options: AxiosRequestConfig = {}) {
	const response = await axios.get<ParamAccessGroup[]>('/access' + path, { ...options, ...commonOptions });
	return response.data;
}

export async function postParamAccess(path: string, group: string, rw: boolean | null) {
	const response = await axios.post<ParamAccessGroup>(
		'/access' + path,
		qs.stringify({ group, rw: rw === null ? 'null' : rw }),
		commonUrlencodedOptions,
	);
	return response.data;
}

export async function deleteParamAccess(path: string, group: string) {
	const response = await axios.request<ParamAccessGroup>({
		...commonUrlencodedOptions,
		url: '/access' + path,
		method: 'DELETE',
		data: qs.stringify({ group }),
	});
	return response.data;
}

interface WhoAmI {
	username: string;
	can_edit_groups: boolean;
}

export async function getWhoAmI() {
	const response = await axios.get<WhoAmI>('/whoami', commonOptions);
	return response.data;
}

export async function getUsers(term: string) {
	const response = await axios.get<string[]>('/user', { ...commonOptions, params: { term } });
	return response.data;
}

export interface UIConfig {
	avatar?: AvatarConfig;
}

export interface AvatarConfig {
	uri: string;
	domain?: string;
	gravatar?: boolean;
	rename?: { [K: string]: string };
	link?: {
		uri: string;
		rename?: { [K: string]: string };
	};
}

export async function getUIConfig() {
	const response = await axios.get<UIConfig>('/ui-config', commonOptions);
	return response.data;
}
