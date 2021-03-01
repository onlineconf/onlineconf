import * as React from 'react';
import { AxiosError } from 'axios';
import { createStyles, Theme, withStyles, WithStyles } from '@material-ui/core/styles';
import Snackbar from '@material-ui/core/Snackbar';

const styles = (theme: Theme) => createStyles({
	content: {
		backgroundColor: theme.palette.error.dark,
	},
});

interface ErrorSnackbarProps {
	error?: Error;
}

interface ErrorSnackbarState {
	error?: Error;
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
			const response = (error as AxiosError).response;
			if (response !== undefined && typeof response.data === 'object' && response.data.message) {
				message = (
					<div>
						<div>{error.message}</div>
						<div>{response.data.message}</div>
					</div>
				);
			} else {
				message = <span>{error.message}</span>;
			}
		}
		return (
			<Snackbar
				open={this.state.open}
				message={message}
				onClose={this.handleClose}
				onExited={this.handleExited}
				ContentProps={{ className: this.props.classes.content }}
			/>
		);
	}

}

export default withStyles(styles)(ErrorSnackbar);
