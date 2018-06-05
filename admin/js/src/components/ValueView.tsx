import * as React from 'react';

import { ValueProps } from './common';
import typeMap from './value';

export default function ValueView(props: ValueProps) {
	const View = props.value === null ? typeMap['application/x-null'].view : typeMap[props.type].view;
	return <View type={props.type} value={props.value} />;
}
