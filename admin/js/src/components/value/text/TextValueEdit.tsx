import * as React from 'react';
import { InputProps } from '@material-ui/core/Input';
import { TextField, createStyles, WithStyles, withStyles } from '@material-ui/core';

import { EditNonnullValueProps } from '../../common';
import CodeMirror from '../../CodeMirror';

const styles = createStyles({
	codemirror: {
		width: 'inherit',
		padding: '3px 0 7px 0',
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

export default withStyles(styles)(
	class TextValueEdit extends React.Component<EditNonnullValueProps & WithStyles<typeof styles>> {

		codeMirrorEditor = (props: EditNonnullValueProps & InputProps) => {
			const { onFocus, onBlur } = props;
			return (
				<CodeMirror
					value={props.value}
					options={{ mode: this.props.type }}
					onBeforeChange={(editor, data, value) => props.onChange({ target: { value } })}
					onFocus={onFocus ? (editor, event) => onFocus(event as any) : undefined}
					onBlur={onBlur ? (editor, event) => onBlur(event as any) : undefined}
					className={this.props.classes.codemirror}
				/>
			);
		}

		helperText() {
			return this.props.type === 'application/x-template' ? (
				<ul className={this.props.classes.helper}>
					<li>{'${hostname} - полное имя хоста'}</li>
					<li>{'${short_hostname} - сокращенное имя хоста'}</li>
					<li>{'${ip} - ip-адрес, соответствующий хосту'}</li>
					<li>{'${/path/of/parameter} - значение параметра'}</li>
				</ul>
			) : undefined;
		}

		render() {
			return (
				<TextField
					id="value"
					label="Value"
					fullWidth
					margin="dense"
					helperText={this.helperText()}
					value={this.props.value}
					onChange={this.props.onChange}
					InputProps={{ inputComponent: this.codeMirrorEditor }}
				/>
			);
		}

	}
);
