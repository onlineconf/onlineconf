import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { InputBaseComponentProps } from '@material-ui/core/InputBase';
import { TextField, createStyles, WithStyles, withStyles } from '@material-ui/core';

import { EditNonnullValueProps } from '../../common';
import CodeMirror from '../../CodeMirror';

const styles = createStyles({
	codemirror: {
		width: 'inherit',
		padding: '15px 14px',
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

export default withTranslation()(withStyles(styles)(
	class TextValueEdit extends React.Component<EditNonnullValueProps & WithStyles<typeof styles> & WithTranslation> {

		codeMirrorEditor = (props: InputBaseComponentProps) => {
			const { onFocus, onBlur } = props;
			return (
				<CodeMirror
					value={props.value}
					options={{ mode: this.props.type, scrollbarStyle: 'null' }}
					onBeforeChange={(editor, data, value) => this.props.onChange({ target: { value } })}
					onFocus={onFocus ? (editor, event) => onFocus(event as any) : undefined}
					onBlur={onBlur ? (editor, event) => onBlur(event as any) : undefined}
					className={this.props.classes.codemirror}
				/>
			);
		}

		helperText() {
			const { t } = this.props;
			return this.props.type === 'application/x-template' ? (
				<ul className={this.props.classes.helper}>
					<li>{'${hostname}'} - {t('param.template.hostname')}</li>
					<li>{'${short_hostname}'} - {t('param.template.shortHostname')}</li>
					<li>{'${ip}'} - {t('param.template.ip')}</li>
					<li>{'${/path/of/parameter}'} - {t('param.template.param')}</li>
				</ul>
			) : undefined;
		}

		render() {
			return (
				<TextField
					id="value"
					label={this.props.t('param.value')}
					fullWidth
					variant="outlined"
					margin="dense"
					helperText={this.helperText()}
					value={this.props.value}
					InputProps={{ inputComponent: this.codeMirrorEditor }}
				/>
			);
		}

	}
));
