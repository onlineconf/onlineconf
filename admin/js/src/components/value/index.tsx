import * as React from 'react';
import { ValueProps, EditValueProps } from '../common';
import NoAccess from '../NoAccess';
import { NonNullParamType, NonNullValueProps, EditNonnullValueProps } from './common';
import { NullValuePreview, NullValueView, NullValueEdit } from './null';
import textValue from './text';
import templateValue from './template';
import symlinkValue from './symlink';
import caseValue from './case';
import listValue from './list';
import serverValue from './server';
import server2Value from './server2';

interface NonNullTypeConfig {
	preview: React.ComponentType<NonNullValueProps>;
	view: React.ComponentType<NonNullValueProps>;
	edit: React.ComponentType<EditNonnullValueProps>;
}

const typeMap: { [P in NonNullParamType]: NonNullTypeConfig } = {
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

export function ValuePreview(props: ValueProps) {
	const { type, value, ...rest } = props;
	if (type === 'application/x-null') {
		return <NullValuePreview {...rest} type="application/x-null" value={null}/>;
	} else {
		const Preview = typeMap[type].preview;
		return <Preview {...rest} type={type} value={value !== null ? value : ''}/>;
	}
}

interface ValueViewProps extends ValueProps {
	accessible: boolean;
}

export function ValueView(props: ValueViewProps) {
	const { type, value, ...rest } = props;
	if (!props.accessible) {
		return <NoAccess/>;
	} else if (type === 'application/x-null') {
		return <NullValueView {...rest} type="application/x-null" value={null}/>;
	} else {
		const View = typeMap[type].view;
		return <View {...rest} type={type} value={value !== null ? value : ''} />;
	}
}

export function ValueEdit(props: EditValueProps) {
	const { type, value, ...rest } = props;
	if (type === 'application/x-null') {
		return <NullValueEdit {...rest} type="application/x-null" value={null}/>;
	} else {
		const Edit = typeMap[type].edit;
		return <Edit {...rest} type={type} value={value !== null ? value : ''}/>;
	}
}
