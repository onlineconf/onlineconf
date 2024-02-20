import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import Button from '@mui/material/Button';

import NotificationControl from './NotificationControl';
import ParamDialogTitle from './ParamDialogTitle';

import { Notification, postParam, IParam } from '../api';

interface ParamNotificationProps {
	path: string;
	overridden: boolean;
	value: Notification;
	onChange: (param: IParam) => void;
	onError: (error: unknown) => void;
	onClose: () => void;
}

class ParamNotification extends React.Component<ParamNotificationProps & WithTranslation> {

	handleChange = async (value: Notification | null) => {
		try {
			const data = await postParam(this.props.path, { notification: value });
			this.props.onChange(data);
		} catch (error) {
			this.props.onError(error);
		}
	};

	render() {
		return (
			<Dialog open onClose={this.props.onClose}>
				<ParamDialogTitle path={this.props.path}>{this.props.t('param.menu.notifications')}</ParamDialogTitle>
				<DialogContent>
					<NotificationControl overridden={this.props.overridden} value={this.props.value} onChange={this.handleChange}/>
				</DialogContent>
				<DialogActions>
					<Button color="primary" onClick={this.props.onClose}>{this.props.t('button.close')}</Button>
				</DialogActions>
			</Dialog>
		);
	}

}

export default withTranslation()(ParamNotification);
