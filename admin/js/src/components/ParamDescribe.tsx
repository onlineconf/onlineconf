import * as React from 'react';
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

export default class ParamDescribe extends React.Component<ParamDescribeProps, ParamDescribeState> {

	constructor(props: ParamDescribeProps) {
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
		return (
			<Dialog open onClose={this.props.onClose} PaperProps={{ component: 'form' as any, onSubmit: this.handleSubmit }}>
				<ParamDialogTitle path={this.props.path}>Describe</ParamDialogTitle>
				<DialogContent>
					<SummaryDescriptionFields summary={this.state.summary} description={this.state.description} onChange={this.handleChange}/>
				</DialogContent>
				<DialogActions>
					<Button color="primary" onClick={this.props.onClose}>Cancel</Button>
					<Button color="primary" type="submit">OK</Button>
				</DialogActions>
			</Dialog>
		);
	}

}
