import * as React from 'react';
import { createStyles, WithStyles, withStyles, Theme } from '@material-ui/core/styles';
import IconButton from '@material-ui/core/IconButton';
import TextField from '@material-ui/core/TextField';

import AddIcon from '@material-ui/icons/AddCircle';
import RemoveIcon from '@material-ui/icons/RemoveCircle';

import { EditNonnullValueProps } from '../common';
import ValueOutline from '../ValueOutline';

const styles = (theme: Theme) => createStyles({
	list: {
		padding: 0,
		margin: '2px 14px 2px 2px',
	},
	item: {
		display: 'flex',
	},
});

class ListValueEdit extends React.Component<EditNonnullValueProps & WithStyles<typeof styles>> {

	handleAddItem = () => {
		this.props.onChange(this.props.value + ',');
	}

	createRemoveItemHandler(id: number) {
		return () => {
			const list = this.props.value.split(',');
			list.splice(id, 1);
			this.props.onChange(list.join(','));
		};
	}

	createChangeItemHandler(id: number) {
		return (event: React.ChangeEvent<HTMLInputElement>) => {
			const list = this.props.value.split(',');
			list[id] = event.target.value;
			this.props.onChange(list.join(','));
		};
	}

	render() {
		const list = this.props.value.split(',');
		return (
			<ValueOutline>
				<ul className={this.props.classes.list}>
					{list.map((item, i) => (
						<li key={i} className={this.props.classes.item}>
							<IconButton onClick={this.createRemoveItemHandler(i)}><RemoveIcon/></IconButton>
							<TextField margin="dense" fullWidth value={item} onChange={this.createChangeItemHandler(i)} autoFocus={item === ''}/>
						</li>
					))}
					<li className={this.props.classes.item}>
						<IconButton onClick={this.handleAddItem}><AddIcon /></IconButton>
					</li>
				</ul>
			</ValueOutline>
		);
	}

}

export default withStyles(styles)(ListValueEdit);
