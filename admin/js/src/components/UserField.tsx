import * as React from 'react';
import { Omit } from '@material-ui/core';

import { getUsers } from '../api';
import Autocomplete, { AutocompleteProps, AutocompleteOption } from './Autocomplete';

export default class UserField extends React.Component<Omit<AutocompleteProps, 'loadOptions'>> {

	cache: { [K: string]: AutocompleteOption[] } = {};

	loadOptions = async (value: string) => {
		if (value.length < 2) {
			return [];
		}
		const cached = this.cache[value];
		if (cached !== undefined) {
			return cached;
		}
		const users = await getUsers(value);
		const options = users.map(user => ({ label: user, value: user }));
		this.cache[value] = options;
		return options;
	}

	render() {
		return <Autocomplete {...this.props} loadOptions={this.loadOptions}/>;
	}

}
