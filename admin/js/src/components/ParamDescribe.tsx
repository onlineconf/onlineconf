import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import Button from '@material-ui/core/Button';

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

	private handleSummaryChange = (summary: string) => {
		this.setState({ summary });
	}

	private handleDescriptionChange = (description: string) => {
		this.setState({ description });
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
			<Dialog open onClose={this.props.onClose}>
				<form onSubmit={this.handleSubmit}>
					<ParamDialogTitle path={this.props.path}>{t('param.menu.describe')}</ParamDialogTitle>
					<DialogContent>
						<SummaryDescriptionFields
							summary={this.state.summary}
							description={this.state.description}
							onSummaryChange={this.handleSummaryChange}
							onDescriptionChange={this.handleDescriptionChange}
						/>
					</DialogContent>
					<DialogActions>
						<Button color="primary" onClick={this.props.onClose}>{t('button.cancel')}</Button>
						<Button color="primary" type="submit">{t('button.ok')}</Button>
					</DialogActions>
				</form>
			</Dialog>
		);
	}

}

export default withTranslation()(ParamDescribe);
