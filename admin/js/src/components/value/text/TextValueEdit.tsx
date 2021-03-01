import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { makeStyles } from '@material-ui/core/styles';
import { InputBaseComponentProps } from '@material-ui/core/InputBase';
import TextField from '@material-ui/core/TextField';

import { EditNonnullValueProps } from '../common';
import CodeMirrorComponent from '../../CodeMirror';

const useStyles = makeStyles({
	codemirror: {
		width: 'inherit',
		'& > .CodeMirror': {
			maxHeight: '300px',
			'& .CodeMirror-scroll': {
				maxHeight: '300px',
			},
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
				editorDidMount={(editor) => { editorRef.current = editor; }}
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

	return (
		<TextField
			id="value"
			label={t('param.value')}
			fullWidth
			multiline
			variant="outlined"
			margin="dense"
			helperText={type === 'application/x-template' ? CaseHelperText : undefined}
			value={value}
			InputProps={{
				inputComponent: CodeMirrorEditor,
				inputProps: { type, onBeforeChange: onChange },
			}}
		/>
	);
}
