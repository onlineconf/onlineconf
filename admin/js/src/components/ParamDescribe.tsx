import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { Dialog, DialogContent, DialogActions, Button } from '@material-ui/core';

import { postParam, IParam } from '../api';
import ParamDialogTitle from './ParamDialogTitle';
import SummaryDescriptionFields from './SummaryDescriptionFields';

interface ParamDescribeProps {
	path: string;
	summary: string;
	description: string;
	onDescribed: (param: IParam) => void;
	onError: (error: Error) => void;
	onClose: () => void;
}

interface ParamDescribeState {
	summary: string;
	description: string;
}

class ParamDescribe extends React.Component<ParamDescribeProps & WithTranslation, ParamDescribeState> {

	constructor(props: ParamDescribeProps & WithTranslation) {
		super(props);
		this.state = {
			summary: props.summary,
			description: props.description,
		};
	}

	private handleChange = (newState: ParamDescribeState) => {
		this.setState(newState);
	}

	private handleSubmit = async (event: React.FormEvent) => {
		const { onDescribed, onError } = this.props;
		event.preventDefault();
		try {
			const param = await postParam(this.props.path, this.state);
			onDescribed(param);
		} catch (error) {
			onError(error);
		}
	}

	render() {
		const { t } = this.props;
		return (
			<Dialog open onClose={this.props.onClose} PaperProps={{ component: 'form' as any, onSubmit: this.handleSubmit }}>
				<ParamDialogTitle path={this.props.path}>{t('param.menu.describe')}</ParamDialogTitle>
				<DialogContent>
					<SummaryDescriptionFields summary={this.state.summary} description={this.state.description} onChange={this.handleChange}/>
				</DialogContent>
				<DialogActions>
					<Button color="primary" onClick={this.props.onClose}>{t('button.cancel')}</Button>
					<Button color="primary" type="submit">{t('button.ok')}</Button>
				</DialogActions>
			</Dialog>
		);
	}

}

export default withTranslation()(ParamDescribe);
