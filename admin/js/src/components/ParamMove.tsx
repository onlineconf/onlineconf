import * as React from 'react';
import { Dialog, DialogContent, DialogActions, Button, TextField, FormControlLabel, Checkbox } from '@material-ui/core';

import { postParam } from '../api';
import ParamDialogTitle from './ParamDialogTitle';
import PathField from './PathField';

interface ParamMoveProps {
	path: string;
	version: number;
	onMoved: (path: string) => void;
	onError: (error: Error) => void;
	onClose: () => void;
}

interface ParamMoveState {
	path: string;
	symlink: boolean;
	comment: string;
}

export default class ParamMove extends React.Component<ParamMoveProps, ParamMoveState> {

	state: ParamMoveState = {
		path: '',
		symlink: true,
		comment: '',
	};

	handlePathChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		this.setState({ path: event.target.value });
	}

	handleSymlinkChange = (event: React.ChangeEvent, checked: boolean) => {
		this.setState({ symlink: checked });
	}

	handleCommentChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		this.setState({ comment: event.target.value });
	}

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
	}

	render() {
		return (
			<Dialog open onClose={this.props.onClose} PaperProps={{ component: 'form' as any, onSubmit: this.handleConfirm }}>
				<ParamDialogTitle path={this.props.path}>Move</ParamDialogTitle>
				<DialogContent>
					<PathField label="Move to" required value={this.state.path} onChange={this.handlePathChange} fullWidth autoFocus margin="dense"/>
					<FormControlLabel label="Leave symlink" control={<Checkbox value="1" checked={this.state.symlink} onChange={this.handleSymlinkChange}/>}/>
					<TextField label="Comment" required value={this.state.comment} onChange={this.handleCommentChange} fullWidth margin="dense"/>
				</DialogContent>
				<DialogActions>
					<Button onClick={this.props.onClose}>Cancel</Button>
					<Button color="primary" type="submit">OK</Button>
				</DialogActions>
			</Dialog>
		);
	}

}
