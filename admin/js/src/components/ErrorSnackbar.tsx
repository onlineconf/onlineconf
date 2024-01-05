import * as React from 'react';
import { AxiosError } from 'axios';
import { createStyles, withStyles, WithStyles } from '@mui/styles';
import { Theme } from '@mui/material/styles';
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
			if (error instanceof AxiosError) {
				const response = error.response;
				if (response !== undefined && typeof response.data === 'object' && response.data !== null && response.data.message) {
					message = (
						<div>
							<div>{error.message}</div>
							<div>{response.data.message}</div>
						</div>
					);
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
				TransitionProps={{ onExited: this.handleExited}}
				ContentProps={{ className: this.props.classes.content }}
			/>
		);
	}

}

export default withStyles(styles)(ErrorSnackbar);
