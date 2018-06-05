import * as React from 'react';
import * as classNames from 'classnames';
import { Controlled, IControlledCodeMirror } from 'react-codemirror2';
import { withStyles, WithStyles, createStyles } from '@material-ui/core/styles';

import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/javascript/javascript.js';
import 'codemirror/mode/yaml/yaml.js';
import '../template';

const styles = createStyles({
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

const CodeMirror = (props: IControlledCodeMirror & WithStyles<typeof styles>) => {
	const { classes, className: classNameProp, ...rest } = props;
	const className = classNames(classes.root, classNameProp);
	return <Controlled {...rest} className={className}/>;
};

export default withStyles(styles)(CodeMirror);
