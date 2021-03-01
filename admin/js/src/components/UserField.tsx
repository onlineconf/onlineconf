import * as React from 'react';
import { Omit } from '@material-ui/types';
import ListItemAvatar from '@material-ui/core/ListItemAvatar';
import ListItemText from '@material-ui/core/ListItemText';

import { getUsers } from '../api';
import Autocomplete, { AutocompleteProps, AutocompleteOption, AutocompleteItemProps } from './Autocomplete';
import Avatar from './Avatar';

function UserFieldItem(props: AutocompleteItemProps) {
	return (
		<React.Fragment>
			<ListItemAvatar>
				<Avatar username={props.children}/>
			</ListItemAvatar>
			<ListItemText>{props.children}</ListItemText>
		</React.Fragment>
	);
}

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
		return <Autocomplete {...this.props} loadOptions={this.loadOptions} itemComponent={UserFieldItem}/>;
	}

}
