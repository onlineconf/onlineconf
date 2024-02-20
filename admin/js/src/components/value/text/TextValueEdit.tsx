import * as React from 'react';
import { useTranslation } from 'react-i18next';
import makeStyles from '@mui/styles/makeStyles';
import { InputBaseComponentProps } from '@mui/material/InputBase';
import TextField from '@mui/material/TextField';

import { EditNonnullValueProps } from '../common';
import CodeMirrorComponent from '../../CodeMirror';

const useStyles = makeStyles({
	field: {
		minHeight: 0,
	},
	inputRoot: {
		alignItems: 'stretch',
		minHeight: 0,
	},
	input: {
		height: 'auto',
	},
	codemirror: {
		width: 'inherit',
		height: '100%',
		'& > .CodeMirror': {
			height: '100%',
		},
	},
	helper: {
		padding: 0,
		margin: 0,
		listStyleType: 'none',
	},
});

function CodeMirrorEditor(props: InputBaseComponentProps) {
	const { className, inputRef, onFocus, onBlur, onBeforeChange, type, value } = props;
	const classes = useStyles();

	const editorRef = React.useRef<CodeMirror.Editor>();
	React.useImperativeHandle(inputRef, () => ({
		focus: () => editorRef.current?.focus(),
		value,
	}));

	return (
		<div className={className} onClick={() => editorRef.current?.focus()}>
			<CodeMirrorComponent
				value={value as string}
				options={{ mode: type, scrollbarStyle: 'null' }}
				className={classes.codemirror}
				editorDidMount={(editor) => { editorRef.current = editor }}
				onBeforeChange={onBeforeChange ? (editor, data, value) => onBeforeChange(value) : () => undefined}
				onFocus={onFocus ? (editor, event) => onFocus(event) : undefined}
				onBlur={onBlur ? (editor, event) => onBlur(event) : undefined}
			/>
		</div>
	);
}

function CaseHelperText() {
	const classes = useStyles();
	const { t } = useTranslation();

	/* eslint-disable no-template-curly-in-string */
	return (
		<ul className={classes.helper}>
			<li>{'${hostname}'} - {t('param.template.hostname')}</li>
			<li>{'${short_hostname}'} - {t('param.template.shortHostname')}</li>
			<li>{'${ip}'} - {t('param.template.ip')}</li>
			<li>{'${/path/of/parameter}'} - {t('param.template.param')}</li>
		</ul>
	);
	/* eslint-enable no-template-curly-in-string */
}

export default function TextValueEdit(props: EditNonnullValueProps) {
	const { onChange, type, value } = props;
	const { t } = useTranslation();
	const classes = useStyles();

	return (
		<TextField
			id="value"
			label={t('param.value')}
			fullWidth
			variant="outlined"
			margin="dense"
			helperText={type === 'application/x-template' ? CaseHelperText : undefined}
			value={value}
			className={classes.field}
			InputProps={{
				inputComponent: CodeMirrorEditor,
				inputProps: { type, onBeforeChange: onChange, className: classes.input },
				className: classes.inputRoot,
			}}
		/>
	);
}
