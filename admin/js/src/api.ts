import axios, { AxiosRequestConfig } from 'axios';
import { DistributiveOmit } from '@mui/types';

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

// ValueFlags are the orthogonal properties of a value, next to its type
// wherever a value travels: the one place the flag is declared.
export interface ValueFlags {
	// the value is stored encrypted: in a listing its data is null — a
	// masked secret, not an empty value — until explicitly revealed
	encrypted?: boolean;
}

interface ParamFields extends ValueFlags {
	path: string;
	name: string;
	mime: ParamType;
	// the parameter holds encrypted values (its own or in case branches) and
	// the current user may reveal them
	can_reveal?: boolean;
	data: string | null;
	summary: string;
	description: string;
	version: number;
	mtime: string;
}

export interface IParam extends ParamFields {
	rw: boolean | null;
	notification: Notification;
	access_modified: boolean;
	notification_modified: boolean;
	num_children: number;
	children?: IParam[];
}

export interface ParamModify extends Partial<ParamFields> {
	notification?: Notification | null;
	symlink?: boolean;
	comment?: string;
	// server-side rollback to a historical version (no plaintext round trip)
	restore?: number;
}


const commonOptions = { headers: { 'X-Requested-With': 'XMLHttpRequest' } };
const commonUrlencodedOptions = {
	...commonOptions,
	headers: { ...commonOptions.headers, 'Content-Type': 'application/x-www-form-urlencoded' },
};

// Percent-encode path segments so that characters like '#' (used in the paths
// of deleted nodes renamed aside to free their original path) are not
// interpreted as an URL fragment or query.
function encodePath(path: string) {
	return path.split('/').map(encodeURIComponent).join('/');
}

// reveal asks for the plaintext of an encrypted value: an audited request
// the server accepts as reveal=1 only
export async function getParam(path: string, symlink?: 'resolve' | 'follow', depth?: 'children' | 'subtree' | 'none', options: AxiosRequestConfig = {}, reveal?: boolean) {
	const params: { symlink?: string, depth?: string, reveal?: 1 } = { symlink, depth };
	if (reveal) {
		params.reveal = 1;
	}
	const response = await axios.get<IParam>('/config' + encodePath(path), { ...options, ...commonOptions, params });
	return response.data;
}

export async function postParam(path: string, modify: ParamModify) {
	const { data, notification, symlink, encrypted, can_reveal, restore, ...rest } = modify;
	const params = new URLSearchParams(rest as Record<string, string>);
	if (data !== undefined) {
		params.append('data', data !== null ? data : '');
	}
	if (notification !== undefined) {
		params.append('notification', notification !== null ? notification : '');
	}
	if (symlink) {
		params.append('symlink', '1');
	}
	// stated only when the caller means it: on an encrypted parameter the
	// server requires it, on a plain one an absent field keeps it plain
	if (encrypted !== undefined) {
		params.append('encrypted', encrypted ? '1' : '0');
	}
	if (restore !== undefined) {
		params.append('restore', restore.toString());
	}
	const response = await axios.post<IParam>('/config' + encodePath(path), params, commonUrlencodedOptions);
	return response.data;
}

export async function deleteParam(path: string, info: { version: number, comment: string }) {
	const response = await axios.request<void>({
		...commonUrlencodedOptions,
		url: '/config' + encodePath(path),
		method: 'DELETE',
		data: new URLSearchParams({ version: info.version.toString(), comment: info.comment }),
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

export interface IParamLog extends ValueFlags {
	id: number;
	path: string;
	version: number;
	mime: ParamType;
	can_reveal?: boolean;
	data: string | null;
	mtime: string;
	author: string;
	comment: string;
	deleted: boolean;
	rw: boolean | null;
	same: boolean;
}

export async function getParamLog(path: string, lastID?: number, options: AxiosRequestConfig = {}, reveal?: boolean) {
	const params: LogPagination & { reveal?: 1 } = { lastid: lastID };
	if (reveal) {
		params.reveal = 1;
	}
	const response = await axios.get<IParamLog[]>('/log' + encodePath(path), { ...options, ...commonOptions, params });
	return response.data;
}

export interface GlobalLogFilter {
	author?: string;
	branch?: string;
	from?: string;
	till?: string;
	all?: boolean;
}

interface GlobalLogParams extends DistributiveOmit<GlobalLogFilter, 'all'>, LogPagination {
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
	const response = await axios.get<ParamAccessGroup[]>('/access' + encodePath(path), { ...options, ...commonOptions });
	return response.data;
}

export async function postParamAccess(path: string, group: string, rw: boolean | null) {
	const response = await axios.post<ParamAccessGroup>(
		'/access' + encodePath(path),
		new URLSearchParams({ group, rw: String(rw) }),
		commonUrlencodedOptions,
	);
	return response.data;
}

export async function deleteParamAccess(path: string, group: string) {
	const response = await axios.request<ParamAccessGroup>({
		...commonUrlencodedOptions,
		url: '/access' + encodePath(path),
		method: 'DELETE',
		data: new URLSearchParams({ group }),
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
	encryption: EncryptionConfig;
}

export interface EncryptionConfig {
	// encrypted writes are accepted on this instance (switch on, key loaded)
	enabled: boolean;
	// who may reveal: users with write ('rw') or read ('ro') access
	reveal: 'rw' | 'ro';
}

export interface AvatarConfig {
	uri: string;
	domain?: string;
	gravatar?: boolean;
	extension?: boolean;
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
