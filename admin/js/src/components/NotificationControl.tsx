import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { Theme } from '@mui/material/styles';
import { WithStyles } from '@mui/styles';
import createStyles from '@mui/styles/createStyles';
import withStyles from '@mui/styles/withStyles';
import MenuItem from '@mui/material/MenuItem';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import Radio from '@mui/material/Radio';
import RadioGroup from '@mui/material/RadioGroup';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';

import { Notification } from '../api';
import WhoAmIContext from './WhoAmIContext';

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
	},
});

interface NotificationControlProps {
	compact?: boolean;
	overridden: boolean;
	value: Notification;
	onChange: (value: Notification | null) => void;
}

interface NotificationControlState {
	indeterminate: boolean;
}

class NotificationControl extends React.Component<NotificationControlProps & WithStyles<typeof styles> & WithTranslation, NotificationControlState> {

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
	};

	handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		this.setState({ indeterminate: false });
		this.props.onChange(event.target.value as Notification);
	};

	static contextType = WhoAmIContext;

	render() {
		const { t } = this.props;
		const overridden = this.props.overridden || this.state.indeterminate;
		if (this.props.compact) {
			return (
				<TextField
					select
					label={t('param.notifications.label')}
					value={this.props.value}
					disabled={!overridden}
					onChange={this.handleChange}
					variant="outlined"
					margin="dense"
					fullWidth
					InputProps={{
						startAdornment: (
							<InputAdornment position="start">
								<Checkbox
									disabled={false}
									checked={overridden}
									indeterminate={this.state.indeterminate}
									onChange={this.handleOverride}
									className={this.props.classes.compactCheckbox}
								/>
							</InputAdornment>
						)
					}}
				>
					<MenuItem value="none" disabled={!this.context.userIsRoot}>{t('param.notifications.none')}</MenuItem>
					<MenuItem value="no-value">{t('param.notifications.noValue')}</MenuItem>
					<MenuItem value="with-value">{t('param.notifications.withValue')}</MenuItem>
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
						<FormControlLabel value="none" label={t('param.notifications.none')} disabled={!overridden || !this.context.userIsRoot} control={<Radio/>}/>
						<FormControlLabel value="no-value" label={t('param.notifications.noValue')} disabled={!overridden} control={<Radio/>}/>
						<FormControlLabel value="with-value" label={t('param.notifications.withValue')} disabled={!overridden} control={<Radio/>}/>
					</RadioGroup>
				</div>
			);
		}
	}

}

export default withTranslation()(withStyles(styles)(NotificationControl));
