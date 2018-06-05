import * as React from 'react';
import axios, { CancelTokenSource } from 'axios';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import Table from '@material-ui/core/Table';
import TableHead from '@material-ui/core/TableHead';
import TableBody from '@material-ui/core/TableBody';
import TableRow from '@material-ui/core/TableRow';
import TableCell from '@material-ui/core/TableCell';
import Button from '@material-ui/core/Button';
import { withMobileDialog } from '@material-ui/core';

import * as API from '../api';
import { ParamDialogProps } from './common';
import ParamDialogTitle from './ParamDialogTitle';
import ValueView from './ValueView';

interface ParamLogState {
	open: boolean;
	data: API.IParamLog[];
}

export default withMobileDialog<ParamDialogProps>()(
	class ParamLog extends React.Component<ParamDialogProps, ParamLogState> {

		state: ParamLogState = {
			open: false,
			data: [],
		};

		private cts?: CancelTokenSource;

		componentDidMount() {
			this.load();
		}

		componentDidUpdate(prevProps: ParamDialogProps) {
			if (this.props.path !== prevProps.path) {
				this.cancel();
				this.load();
			}
		}

		componentWillUnmount() {
			this.cancel();
		}

		private async load() {
			const { onLoaded, onError } = this.props;
			try {
				this.setState({ open: false });
				this.cts = axios.CancelToken.source();
				const data = await API.getParamLog(this.props.path, { cancelToken: this.cts.token });
				this.setState({ data, open: true });
				onLoaded();
			} catch (error) {
				if (axios.isCancel(error)) {
					onLoaded();
				} else {
					onError(error);
				}
			}
		}

		private cancel() {
			if (this.cts) {
				this.cts.cancel('Operation canceled by the user.');
				this.cts = undefined;
			}
		}

		render() {
			const { path, onClose } = this.props;
			const { data, open } = this.state;

			return (
				<Dialog
					open={open}
					onClose={onClose}
					maxWidth="md"
					fullScreen={this.props.fullScreen}
				>
					<ParamDialogTitle path={path}>Log</ParamDialogTitle>
					<DialogContent>
						<Table>
							<TableHead>
								<TableRow>
									<TableCell>#</TableCell>
									<TableCell>Value</TableCell>
									<TableCell>Time</TableCell>
									<TableCell>Author</TableCell>
									<TableCell>Comment</TableCell>
								</TableRow>
							</TableHead>
							<TableBody>
								{data.map(row => {
									return (
										<TableRow key={row.version}>
											<TableCell numeric>{row.version}</TableCell>
											<TableCell><ValueView type={row.mime} value={row.data} /></TableCell>
											<TableCell>{row.mtime}</TableCell>
											<TableCell>{row.author}</TableCell>
											<TableCell>{row.comment}</TableCell>
										</TableRow>
									);
								})}
							</TableBody>
						</Table>
					</DialogContent>
					<DialogActions>
						<Button color="primary" onClick={onClose}>Close</Button>
					</DialogActions>
				</Dialog>
			);
		}

	}
);
