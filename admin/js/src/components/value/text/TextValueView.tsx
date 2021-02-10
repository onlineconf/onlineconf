import * as React from 'react';

import { NonNullValueProps } from '../common';
import CodeMirror from '../../CodeMirror';

const TextValueView = (props: NonNullValueProps) => (
	<CodeMirror
		value={props.value}
		options={{
			mode: props.type,
			readOnly: true,
			tabindex: -1,

		}}
		onBeforeChange={() => null}
	/>
);

export default TextValueView;
