import { TreeNodeState } from './TreeNode';
import { ParamType, IParam } from '../api';
import { Omit } from '@material-ui/core';

export interface ValueProps {
	type: ParamType;
	value: string | null;
}

export interface EditValueProps extends ValueProps {
	onChange: (props: { target: { value: string | null } }) => void;
	onError: (error: Error) => void;
}

export interface NonNullValueProps {
	type: ParamType;
	value: string;
}

export interface EditNonnullValueProps extends NonNullValueProps {
	onChange: (props: { target: { value: string } }) => void;
	onError: (error: Error) => void;
}

export interface Case {
	mime: ParamType;
	value: string | null;
	server?: string;
	group?: string;
	datacenter?: string;
	service?: string;
}

export interface IParamNode extends Omit<IParam, 'children'> {
	children?: { [name: string]: IParamNode };
	state: TreeNodeState;
	selected?: boolean;
	match?: boolean;
	hidden?: boolean;
	forceOpen?: boolean;
	logLoading?: boolean;
	accessLoading?: boolean;
}

export const typeNames: { [P in ParamType]: string } = {
	'application/x-null': 'Null',
	'text/plain': 'Text',
	'application/x-template': 'Template',
	'application/json': 'JSON',
	'application/x-yaml': 'YAML',
	'application/x-symlink': 'Symbolic link',
	'application/x-case': 'Case',
	'application/x-list': 'Список',
	'application/x-server': 'Список пар ip:port',
	'application/x-server2': 'Список портов для ip',
};

export interface ParamDialogProps {
	path: string;
	fullScreen?: boolean;
	onClose: () => void;
	onLoaded: () => void;
	onError: (error: Error) => void;
}

export function smartCompare(a: string, b: string) {
	const amatch = a.match(/^(\D*?)(\d+)(\D*?)$/);
	const bmatch = b.match(/^(\D*?)(\d+)(\D*?)$/);
	if (amatch && bmatch && amatch[1] === bmatch[1]) {
		const anum = parseInt(amatch[2], 10);
		const bnum = parseInt(bmatch[2], 10);
		return (anum === bnum ? amatch[3] > bmatch[3] : anum > bnum) ? 1 : -1;
	}
	else {
		return a > b ? 1 : -1;
	}
}
