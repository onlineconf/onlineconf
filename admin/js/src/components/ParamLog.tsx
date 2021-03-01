import * as React from 'react';
import axios, { CancelTokenSource } from 'axios';
import { withTranslation, WithTranslation } from 'react-i18next';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import Button from '@material-ui/core/Button';
import withMobileDialog from '@material-ui/core/withMobileDialog';

import * as API from '../api';
import { ParamDialogProps } from './common';
import ParamDialogTitle from './ParamDialogTitle';
import LogCard from './LogCard';

interface ParamLogState {
	open: boolean;
	data: API.IParamLog[];
}

export default withMobileDialog<ParamDialogProps>()(withTranslation()(
	class ParamLog extends React.Component<ParamDialogProps & WithTranslation, ParamLogState> {

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
			const { path, onClose, t } = this.props;
			const { data, open } = this.state;

			return (
				<Dialog
					open={open}
					onClose={onClose}
					maxWidth="md"
					fullScreen={this.props.fullScreen}
				>
					<ParamDialogTitle path={path}>{t('param.menu.history')}</ParamDialogTitle>
					<DialogContent>
						{data.map(row => <LogCard key={row.version} {...row}/>)}
					</DialogContent>
					<DialogActions>
						<Button color="primary" onClick={onClose}>{t('button.close')}</Button>
					</DialogActions>
				</Dialog>
			);
		}

	}
));
