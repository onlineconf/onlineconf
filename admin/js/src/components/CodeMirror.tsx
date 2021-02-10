import * as React from 'react';
import clsx from 'clsx';
import { Controlled, IControlledCodeMirror } from 'react-codemirror2';
import { makeStyles } from '@material-ui/core/styles';

import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/javascript/javascript.js';
import 'codemirror/mode/yaml/yaml.js';
import '../template';

const useStyles = makeStyles({
	root: {
		'& > .CodeMirror': {
			height: 'auto',
			background: 'none',
			fontFamily: 'monospace, monospace',
			'& .CodeMirror-lines': {
				padding: 0,
				'& .CodeMirror-line': {
					padding: 0,
				}
			}
		},
	},
});

export default function CodeMirror(props: IControlledCodeMirror) {
	const classes = useStyles();
	return <Controlled {...props} className={clsx(props.className, classes.root)} />;
}
