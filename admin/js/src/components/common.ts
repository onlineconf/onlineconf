import { TreeNodeState } from './TreeNode';
import { ParamType, IParam } from '../api';
import { Omit } from '@material-ui/types';

export interface ValueProps {
	type: ParamType;
	value: string | null;
}

export interface EditValueProps extends ValueProps {
	onChange: (value: string | null) => void;
	onError: (error: Error) => void;
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

const typesObj: { [P in ParamType]: null } = {
	'application/x-null': null,
	'text/plain': null,
	'application/x-template': null,
	'application/json': null,
	'application/x-yaml': null,
	'application/x-symlink': null,
	'application/x-case': null,
	'application/x-list': null,
	'application/x-server': null,
	'application/x-server2': null,
};
export const types = Object.keys(typesObj) as ParamType[];

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
