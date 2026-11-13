import { TreeNodeState } from './TreeNode';
import { ParamType, IParam, ValueFlags } from '../api';
import { DistributiveOmit } from '@mui/types';

export interface ValueProps extends ValueFlags {
	type: ParamType;
	value: string | null;
}

// An editor gets the value and its type; whether the value is stored
// encrypted is the business of the switch above it (TypeValueFields), the
// editor only hides what is typed when asked to.
export interface EditValueProps extends Pick<ValueProps, 'type' | 'value'> {
	// hide what is typed (password-style), for secrets
	masked?: boolean;
	onChange: (value: string | null) => void;
	onError: (error: unknown) => void;
}

// Types whose values may be stored encrypted: everything except the special
// types whose value is not free content (null has none; a symlink target and
// a case structure must stay readable — case branches carry their own flag).
// Must match EncryptionEligible in the backend (admin/go/common/encryption_storage.go).
export function encryptionEligible(type: ParamType): boolean {
	switch (type) {
		case 'application/x-null':
		case 'application/x-symlink':
		case 'application/x-case':
			return false;
		default:
			return true;
	}
}

export interface IParamNode extends DistributiveOmit<IParam, 'children'> {
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
	onError: (error: unknown) => void;
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
