import * as React from 'react';

import { ValueProps } from './common';
import typeMap from './value';

import NoAccess from './NoAccess';

interface ValueViewProps extends ValueProps {
	accessible: boolean;
}

export default function ValueView(props: ValueViewProps) {
	if (props.accessible) {
		const View = props.value === null ? typeMap['application/x-null'].view : typeMap[props.type].view;
		return <View type={props.type} value={props.value} />;
	} else {
		return <NoAccess/>;
	}
}
