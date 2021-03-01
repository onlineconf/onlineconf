import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { createStyles, withStyles, WithStyles, Theme } from '@material-ui/core/styles';
import MenuItem from '@material-ui/core/MenuItem';
import Checkbox from '@material-ui/core/Checkbox';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import Radio from '@material-ui/core/Radio';
import RadioGroup from '@material-ui/core/RadioGroup';
import TextField from '@material-ui/core/TextField';
import InputAdornment from '@material-ui/core/InputAdornment';

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
	}

	handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		this.setState({ indeterminate: false });
		this.props.onChange(event.target.value as Notification);
	}

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
					<MenuItem value="none" disabled={!this.props.allowNone}>{t('param.notifications.none')}</MenuItem>
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
						<FormControlLabel value="none" label={t('param.notifications.none')} disabled={!overridden || !this.props.allowNone} control={<Radio/>}/>
						<FormControlLabel value="no-value" label={t('param.notifications.noValue')} disabled={!overridden} control={<Radio/>}/>
						<FormControlLabel value="with-value" label={t('param.notifications.withValue')} disabled={!overridden} control={<Radio/>}/>
					</RadioGroup>
				</div>
			);
		}
	}

}

export default withTranslation()(withStyles(styles)(NotificationControl));
