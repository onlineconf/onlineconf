import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import MenuItem from '@material-ui/core/MenuItem';
import TextField from '@material-ui/core/TextField';

import { types } from './common';
import { Case } from './value/case/common';
import { ParamType } from '../api';
import { ValueEdit } from './value';

interface TypeValueFieldsProps {
	type: ParamType;
	value: string | null;
	onChange: (state: { type: ParamType, value: string | null }) => void;
	onError: (error: Error) => void;
}

class TypeValueFields extends React.Component<TypeValueFieldsProps & WithTranslation> {

	handleTypeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		const type = event.target.value as ParamType;
		if (type !== this.props.type) {
			const state: { type: ParamType, value: string | null } = { type, value: this.props.value };
			if (type === 'application/x-null') {
				state.value = null;
			} else if (type === 'application/x-case') {
				state.value = JSON.stringify([{ mime: this.props.type, value: this.props.value }]);
			} else if (this.props.value === null) {
				state.value = '';
			} else if (this.props.type === 'application/x-case') {
				const cases: Case[] = JSON.parse(this.props.value || '[]');
				const defaultCase = cases.filter(c => (c.server === undefined && c.group === undefined && c.datacenter === undefined && c.service === undefined))[0];
				state.value = defaultCase !== undefined && defaultCase.value !== null ? defaultCase.value : '';
			}
			this.props.onChange(state);
		}
	}

	handleValueChange = (value: string | null) => {
		this.props.onChange({ type: this.props.type, value });
	}

	render() {
		return (
			<React.Fragment>
				<TextField
					id="mime"
					select
					label={this.props.t('param.type')}
					value={this.props.type}
					variant="outlined"
					margin="dense"
					fullWidth
					onChange={this.handleTypeChange}
				>
					{types.map((type: ParamType) => (
						<MenuItem key={type} value={type}>{this.props.t(`param.types.${type}`)}</MenuItem>
					))}
				</TextField>
				<ValueEdit type={this.props.type} value={this.props.value} onChange={this.handleValueChange} onError={this.props.onError} />
			</React.Fragment>
		);
	}

}

export default withTranslation()(TypeValueFields);
