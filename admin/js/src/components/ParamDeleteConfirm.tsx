import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';

import { deleteParam } from '../api';
import ParamDialogTitle from './ParamDialogTitle';

interface ParamDeleteConfirmProps {
	path: string;
	version: number;
	onDeleted: () => void;
	onError: (error: Error) => void;
	onClose: () => void;
}

interface ParamDeleteConfirmState {
	comment: string;
}

class ParamDeleteConfirm extends React.Component<ParamDeleteConfirmProps & WithTranslation, ParamDeleteConfirmState> {

	state: ParamDeleteConfirmState = {
		comment: '',
	};

	private handleCommentChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		this.setState({ comment: event.target.value });
	}

	private handleConfirm = async (event: React.FormEvent) => {
		const { onDeleted, onError } = this.props;
		event.preventDefault();
		try {
			await deleteParam(this.props.path, { version: this.props.version, comment: this.state.comment });
			onDeleted();
		} catch (error) {
			onError(error);
		}
	}

	render() {
		const { t } = this.props;
		return (
			<Dialog open onClose={this.props.onClose}>
				<form onSubmit={this.handleConfirm}>
					<ParamDialogTitle path={this.props.path}>{t('param.menu.delete')}</ParamDialogTitle>
					<DialogContent>
						<DialogContentText>{t('param.delete.confirm', { param: this.props.path })}</DialogContentText>
						<TextField label={t('param.comment')} required value={this.state.comment} onChange={this.handleCommentChange} variant="outlined" margin="dense" fullWidth autoFocus/>
					</DialogContent>
					<DialogActions>
						<Button color="primary" onClick={this.props.onClose}>{t('button.cancel')}</Button>
						<Button color="primary" type="submit">{t('button.ok')}</Button>
					</DialogActions>
				</form>
			</Dialog>
		);
	}

}

export default withTranslation()(ParamDeleteConfirm);
