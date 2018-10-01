import * as React from 'react';
import { createStyles, WithStyles, withStyles, IconButton, TextField, Theme } from '@material-ui/core';

import AddIcon from '@material-ui/icons/AddCircle';
import RemoveIcon from '@material-ui/icons/RemoveCircle';

import { EditNonnullValueProps } from '../../common';

const styles = (theme: Theme) => createStyles({
	list: {
		padding: 0,
		margin: `${theme.spacing.unit}px 0`,
	},
	item: {
		display: 'flex',
	},
});

class ListValueEdit extends React.Component<EditNonnullValueProps & WithStyles<typeof styles>> {

	handleAddItem = () => {
		this.props.onChange({ target: { value: this.props.value + ',' } });
	}

	createRemoveItemHandler(id: number) {
		return () => {
			const list = this.props.value.split(',');
			list.splice(id, 1);
			this.props.onChange({ target: { value: list.join(',') } });
		};
	}

	createChangeItemHandler(id: number) {
		return (event: React.ChangeEvent<HTMLInputElement>) => {
			const list = this.props.value.split(',');
			list[id] = event.target.value;
			this.props.onChange({ target: { value: list.join(',') } });
		};
	}

	render() {
		const list = this.props.value.split(',');
		return (
			<ul className={this.props.classes.list}>
				{list.map((item, i) => (
					<li className={this.props.classes.item}>
						<IconButton onClick={this.createRemoveItemHandler(i)}><RemoveIcon/></IconButton>
						<TextField margin="dense" fullWidth value={item} onChange={this.createChangeItemHandler(i)} autoFocus={item === ''}/>
					</li>
				))}
				<li className={this.props.classes.item}>
					<IconButton onClick={this.handleAddItem}><AddIcon /></IconButton>
				</li>
			</ul>
		);
	}

}

export default withStyles(styles)(ListValueEdit);
