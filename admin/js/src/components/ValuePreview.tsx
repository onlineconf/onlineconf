import * as React from 'react';
import typeMap from './value';
import { ValueProps } from './common';

export default function ValuePreview(props: ValueProps) {
	const Preview = props.value === null ? typeMap['application/x-null'].preview : typeMap[props.type].preview;
	return <Preview type={props.type} value={props.value} />;
}
