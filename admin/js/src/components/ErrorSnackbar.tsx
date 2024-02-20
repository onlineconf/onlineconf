import * as React from 'react';
import { AxiosError, AxiosResponse } from 'axios';
import { Theme } from '@mui/material/styles';
import { WithStyles } from '@mui/styles';
import createStyles from '@mui/styles/createStyles';
import withStyles from '@mui/styles/withStyles';
import Snackbar from '@mui/material/Snackbar';

const styles = (theme: Theme) => createStyles({
	content: {
		backgroundColor: theme.palette.error.dark,
	},
});

interface ErrorSnackbarProps {
	error?: unknown;
}

interface ErrorSnackbarState {
	error?: unknown;
	open: boolean;
}

class ErrorSnackbar extends React.Component<ErrorSnackbarProps & WithStyles<typeof styles>, ErrorSnackbarState> {

	state: ErrorSnackbarState = {
		open: false,
	};

	constructor(props: ErrorSnackbarProps & WithStyles<typeof styles>) {
		super(props);
		if (props.error !== undefined) {
			this.state.error = props.error;
			this.state.open = true;
		}
	}

	componentDidUpdate(prevProps: ErrorSnackbarProps) {
		if (this.props.error !== prevProps.error) {
			this.setState(prevState => {
				if (prevState.open) {
					return {
						error: prevState.error,
						open: false,
					};
				} else if (this.props.error !== undefined) {
					return {
						error: this.props.error,
						open: true,
					};
				} else {
					return null;
				}
			});
		}
	}

	private handleClose = () => {
		this.setState({ open: false });
	};

	private handleExited = () => {
		if (this.props.error !== this.state.error && this.props.error !== undefined) {
			this.setState({
				error: this.props.error,
				open: true,
			});
		}
	};

	render() {
		const { error } = this.state;
		let message: React.ReactNode | undefined;
		if (error !== undefined) {
			if (error instanceof Error) {
				const response = (error as AxiosError).response;
				if (response !== undefined) {
					const data = (response as AxiosResponse).data;
					if (data !== undefined && typeof response.data === 'object' && data.message) {
						message = (
							<div>
								<div>{error.message}</div>
								<div>{data.message}</div>
							</div>
						);
					} else {
						message = <span>{error.message}</span>;
					}
				} else {
					message = <span>{error.message}</span>;
				}
			} else {
				message = <span>Unknown error</span>;
			}
		}
		return (
			<Snackbar
				open={this.state.open}
				message={message}
				onClose={this.handleClose}
				ContentProps={{ className: this.props.classes.content }}
				TransitionProps={{
					onExited: this.handleExited
				}} />
		);
	}

}

export default withStyles(styles)(ErrorSnackbar);
