import { ParamType } from '../../api';

export type NullParamType = 'application/x-null';
export type NonNullParamType = Exclude<ParamType, NullParamType>;

export interface NullValueProps {
	type: NullParamType;
	value: null;
}

export interface NonNullValueProps {
	type: NonNullParamType;
	value: string;
}

export interface EditNonnullValueProps extends NonNullValueProps {
	onChange: (value: string) => void;
	onError: (error: Error) => void;
}
