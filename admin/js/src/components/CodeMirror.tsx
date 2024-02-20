import * as React from 'react';
import clsx from 'clsx';
import { defineMIME } from 'codemirror';
import { Controlled, IControlledCodeMirror } from 'react-codemirror2';
import { Theme } from '@mui/material/styles';

import makeStyles from '@mui/styles/makeStyles';

import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/javascript/javascript.js';
import 'codemirror/mode/yaml/yaml.js';
import '../template';

defineMIME('application/x-yaml', 'yaml');

const useStyles = makeStyles((theme: Theme) => ({
	root: {
		'& > .CodeMirror': {
			height: 'auto',
			background: 'none',
			fontFamily: '"Roboto Mono", "Menlo", monospace',
			color: 'inherit',
			'& .CodeMirror-lines': {
				padding: 0,
				'& .CodeMirror-line': {
					padding: 0,
				}
			},
			'& .CodeMirror-cursor': {
				borderLeft: '1px solid currentcolor',
			},
			'& .cm-atom': { color: theme.palette.codemirror.atom },
			'& .cm-comment': { color: theme.palette.codemirror.comment },
			'& .cm-def': { color: theme.palette.codemirror.def },
			'& .cm-keyword': { color: theme.palette.codemirror.keyword },
			'& .cm-meta': { color: theme.palette.codemirror.meta },
			'& .cm-number': { color: theme.palette.codemirror.number },
			'& .cm-string': { color: theme.palette.codemirror.string },
			'& .cm-variable-2': { color: theme.palette.codemirror.variable2 },
		},
	},
}), { name: 'CodeMirror' });

export default function CodeMirror(props: IControlledCodeMirror) {
	const classes = useStyles();
	return <Controlled {...props} className={clsx(props.className, classes.root)} />;
}
