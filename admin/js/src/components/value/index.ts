import { ParamType } from '../../api';
import { ValueProps } from '../common';
import nullValue from './null';
import textValue from './text';
import templateValue from './template';
import symlinkValue from './symlink';
import caseValue from './case';
import listValue from './list';
import serverValue from './server';
import server2Value from './server2';

interface TypeConfig {
	preview: React.ComponentType<ValueProps>;
	view: React.ComponentType<ValueProps>;
	edit: React.ComponentType<ValueProps>;
}

const typeMap: { [P in ParamType]: TypeConfig } = {
	'application/x-null': nullValue,
	'text/plain': textValue,
	'application/x-template': templateValue,
	'application/json': textValue,
	'application/x-yaml': textValue,
	'application/x-symlink': symlinkValue,
	'application/x-case': caseValue,
	'application/x-list': listValue,
	'application/x-server': serverValue,
	'application/x-server2': server2Value,
};

export default typeMap;
