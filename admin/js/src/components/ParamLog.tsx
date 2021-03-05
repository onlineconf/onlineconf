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
import LogList from './LogList';

interface ParamLogProps extends ParamDialogProps {
	onChange: (param: API.IParam) => void;
}

interface ParamLogState {
	open: boolean;
	data: API.IParamLog[];
}

export default withMobileDialog<ParamLogProps>()(withTranslation()(
	class ParamLog extends React.Component<ParamLogProps & WithTranslation, ParamLogState> {

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
				this.setState({ open: false });
				this.load();
			}
		}

		componentWillUnmount() {
			this.cancel();
		}

		private async load() {
			const { onLoaded, onError } = this.props;
			try {
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

		handleChange = (param: API.IParam) => {
			this.load();
			this.props.onChange(param);
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
						<LogList data={data} onChange={this.handleChange}/>
					</DialogContent>
					<DialogActions>
						<Button color="primary" onClick={onClose}>{t('button.close')}</Button>
					</DialogActions>
				</Dialog>
			);
		}

	}
));
