import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import FormControlLabel from '@mui/material/FormControlLabel';
import Checkbox from '@mui/material/Checkbox';

import { postParam } from '../api';
import ParamDialogTitle from './ParamDialogTitle';
import PathField from './PathField';

interface ParamMoveProps {
	path: string;
	version: number;
	onMoved: (path: string) => void;
	onError: (error: unknown) => void;
	onClose: () => void;
}

interface ParamMoveState {
	path: string;
	symlink: boolean;
	comment: string;
}

class ParamMove extends React.Component<ParamMoveProps & WithTranslation, ParamMoveState> {

	state: ParamMoveState = {
		path: '',
		symlink: true,
		comment: '',
	};

	handlePathChange = (value: string) => {
		this.setState({ path: value });
	};

	handleSymlinkChange = (event: React.ChangeEvent, checked: boolean) => {
		this.setState({ symlink: checked });
	};

	handleCommentChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		this.setState({ comment: event.target.value });
	};

	handleConfirm = async (event: React.FormEvent) => {
		const { onMoved, onError } = this.props;
		event.preventDefault();
		try {
			await postParam(this.props.path, {
				path: this.state.path,
				symlink: this.state.symlink,
				version: this.props.version,
				comment: this.state.comment,
			});
			onMoved(this.state.path);
		} catch (error) {
			onError(error);
		}
	};

	render() {
		const { t } = this.props;
		return (
			<Dialog open onClose={this.props.onClose}>
				<form onSubmit={this.handleConfirm}>
					<ParamDialogTitle path={this.props.path}>{t('param.menu.move')}</ParamDialogTitle>
					<DialogContent>
						<PathField label={t('param.move.to')} required value={this.state.path} onChange={this.handlePathChange} variant="outlined" fullWidth autoFocus margin="dense"/>
						<FormControlLabel label={t('param.move.symlink')} control={<Checkbox value="1" checked={this.state.symlink} onChange={this.handleSymlinkChange}/>}/>
						<TextField label={t('param.comment')} required value={this.state.comment} onChange={this.handleCommentChange} variant="outlined" fullWidth margin="dense"/>
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

export default withTranslation()(ParamMove);
