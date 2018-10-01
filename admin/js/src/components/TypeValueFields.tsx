import * as React from 'react';
import { TextField, MenuItem } from '@material-ui/core';

import { typeNames, Case } from './common';
import { ParamType } from '../api';
import ValueEdit from './ValueEdit';

interface TypeValueFieldsProps {
	type: ParamType;
	value: string | null;
	onChange: (state: { type?: ParamType, value?: string | null }) => void;
	onError: (error: Error) => void;
}

export default class TypeValueFields extends React.Component<TypeValueFieldsProps> {

	handleTypeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		const type = event.target.value as ParamType;
		if (type !== this.props.type) {
			const state: { type: ParamType, value?: string | null } = { type };
			if (type === 'application/x-case') {
				state.value = JSON.stringify([{ mime: this.props.type, value: this.props.value }]);
			} else if (this.props.type === 'application/x-case') {
				const cases: Case[] = JSON.parse(this.props.value || '[]');
				state.value = cases.filter(c => (c.server === undefined && c.group === undefined && c.datacenter === undefined))[0].value;
			} else if (type === 'application/x-null' && this.props.value !== null) {
				state.value = null;
			} else if (this.props.type === 'application/x-null' && this.props.value === null) {
				state.value = '';
			}
			this.props.onChange(state);
		}
	}

	handleValueChange = (event: { target: { value: string }}) => {
		this.props.onChange({ value: event.target.value });
	}

	render() {
		return (
			<React.Fragment>
				<TextField
					id="mime"
					select
					label="Data type"
					value={this.props.type}
					margin="dense"
					fullWidth
					onChange={this.handleTypeChange}
				>
					{Object.keys(typeNames).map((type: ParamType) => (
						<MenuItem key={type} value={type}>{typeNames[type]}</MenuItem>
					))}
				</TextField>
				<ValueEdit type={this.props.type} value={this.props.value} onChange={this.handleValueChange} onError={this.props.onError} />
			</React.Fragment>
		);
	}

}
