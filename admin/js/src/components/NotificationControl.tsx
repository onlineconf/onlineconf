import * as React from 'react';
import { RadioGroup, Checkbox, FormControlLabel, Radio, createStyles, withStyles, WithStyles, Theme, TextField, MenuItem, InputAdornment } from '@material-ui/core';

import { Notification } from '../api';

const styles = (theme: Theme) => createStyles({
	root: {
		display: 'flex',
		alignItems: 'center',
	},
	radioGroup: {
		borderLeft: `1px solid ${theme.palette.divider}`,
		paddingLeft: 14,
	},
	compactCheckbox: {
		width: 24,
		height: 24,
		marginBottom: 4,
	},
});

interface NotificationControlProps {
	compact?: boolean;
	overridden: boolean;
	value: Notification;
	allowNone: boolean;
	onChange: (value: Notification | null) => void;
}

interface NotificationControlState {
	indeterminate: boolean;
}

class NotificationControl extends React.Component<NotificationControlProps & WithStyles<typeof styles>, NotificationControlState> {

	state: NotificationControlState = {
		indeterminate: false,
	};

	handleOverride = (event: React.ChangeEvent<HTMLInputElement>) => {
		if (event.target.checked) {
			this.setState({ indeterminate: true });
		} else if (this.state.indeterminate) {
			this.setState({ indeterminate: false });
		} else {
			this.props.onChange(null);
		}
	}

	handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		this.setState({ indeterminate: false });
		this.props.onChange(event.target.value as Notification);
	}

	render() {
		const overridden = this.props.overridden || this.state.indeterminate;
		if (this.props.compact) {
			return (
				<TextField
					select
					label="Notifications"
					value={this.props.value}
					disabled={!overridden}
					onChange={this.handleChange}
					margin="dense"
					fullWidth
					InputProps={{
						startAdornment: (
							<InputAdornment position="start">
								<Checkbox
									checked={overridden}
									indeterminate={this.state.indeterminate}
									onChange={this.handleOverride}
									className={this.props.classes.compactCheckbox}
								/>
							</InputAdornment>
						)
					}}
				>
					<MenuItem value="none" disabled={!this.props.allowNone}>None</MenuItem>
					<MenuItem value="no-value">Without value</MenuItem>
					<MenuItem value="with-value">With value</MenuItem>
				</TextField>
			);
		} else {
			return (
				<div className={this.props.classes.root}>
					<Checkbox
						checked={overridden}
						indeterminate={this.state.indeterminate}
						onChange={this.handleOverride}
					/>
					<RadioGroup value={this.props.value} className={this.props.classes.radioGroup} onChange={this.handleChange}>
						<FormControlLabel value="none" label="None" disabled={!overridden || !this.props.allowNone} control={<Radio/>}/>
						<FormControlLabel value="no-value" label="Without value" disabled={!overridden} control={<Radio/>}/>
						<FormControlLabel value="with-value" label="With value" disabled={!overridden} control={<Radio/>}/>
					</RadioGroup>
				</div>
			);
		}
	}

}

export default withStyles(styles)(NotificationControl);
