import * as React from 'react';

import { EditValueProps } from './common';
import typeMap from './value';

export default function ValueEdit(props: EditValueProps) {
	const Edit = props.value === null ? typeMap['application/x-null'].edit : typeMap[props.type].edit;
	return <Edit {...props} />;
}
