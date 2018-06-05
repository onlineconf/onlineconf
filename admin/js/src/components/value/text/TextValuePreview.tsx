import * as React from 'react';
import { NonNullValueProps } from '../../common';

const TextValuePreview = (props: NonNullValueProps) => (
	<span>{props.value}</span>
);

export default TextValuePreview;
