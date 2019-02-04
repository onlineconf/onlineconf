import * as React from 'react';
import { Dialog, DialogContent, DialogContentText, DialogActions, Button, TextField } from '@material-ui/core';

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

export default class ParamDeleteConfirm extends React.Component<ParamDeleteConfirmProps, ParamDeleteConfirmState> {

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
		return (
			<Dialog open onClose={this.props.onClose} PaperProps={{ component: 'form' as any, onSubmit: this.handleConfirm }}>
				<ParamDialogTitle path={this.props.path}>Delete</ParamDialogTitle>
				<DialogContent>
					<DialogContentText>Are you really want to delete {this.props.path}?</DialogContentText>
					<TextField label="Comment" required value={this.state.comment} onChange={this.handleCommentChange} variant="outlined" margin="dense" fullWidth autoFocus/>
				</DialogContent>
				<DialogActions>
					<Button color="primary" onClick={this.props.onClose}>Cancel</Button>
					<Button color="primary" type="submit">OK</Button>
				</DialogActions>
			</Dialog>
		);
	}

}
