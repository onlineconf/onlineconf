import * as React from 'react';
import { Chip, List, ListItem, createStyles, WithStyles, withStyles, Theme, IconButton, Dialog, DialogTitle, DialogContent, DialogActions, Button, TextField } from '@material-ui/core';

import AddIcon from '@material-ui/icons/AddCircle';

import * as api from '../api';

const styles = (theme: Theme) => createStyles({
	group: {
		display: 'flex',
		flexWrap: 'wrap',
		padding: `${theme.spacing.unit / 2}px ${theme.spacing.unit}px`,
	},
	chip: {
		margin: theme.spacing.unit / 2,
	},
	add: {
		padding: 4,
	},
});

interface AccessProps {
	userIsRoot: boolean;
	onError: (error: Error) => void;
}

interface AccessState {
	access: { [group: string]: string[] };
	group?: string;
	user?: string;
}

class Access extends React.Component<AccessProps & WithStyles<typeof styles>, AccessState> {

	state: AccessState = {
		access: {},
	};

	componentDidMount() {
		this.load();
	}

	private async load() {
		try {
			this.setState({ access: await api.getAccess() });
		} catch (error) {
			this.props.onError(error);
		}
	}

	private showCreateGroupDialog() {
		this.setState({ group: '' });
	}

	private renderCreateGroupDialog() {
		if (this.state.group === undefined) {
			return null;
		}
		const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
			this.setState({ group: event.target.value });
		};
		const handleClose = () => {
			this.setState({ group: undefined });
		};
		const handleCreate = async () => {
			const group = this.state.group;
			if (group === undefined) {
				return;
			}
			try {
				await api.createGroup(group);
				this.setState(({ access }) => ({
					access: { [group]: [], ...access },
					group: undefined,
				}));
			} catch (error) {
				this.props.onError(error);
			}
		};
		return (
			<Dialog open onClose={handleClose}>
				<DialogTitle>Create group</DialogTitle>
				<DialogContent>
					<TextField placeholder="Group" value={this.state.group} onChange={handleChange} autoFocus variant="outlined" margin="dense" fullWidth/>
				</DialogContent>
				<DialogActions>
					<Button color="primary" onClick={handleClose}>Cancel</Button>
					<Button color="primary" onClick={handleCreate} disabled={this.state.group.length === 0}>OK</Button>
				</DialogActions>
			</Dialog>
		);
	}

	private async deleteGroup(group: string) {
		try {
			await api.deleteGroup(group);
			this.setState(({ access }) => {
				delete(access[group]);
				return { access };
			});
		} catch (error) {
			this.props.onError(error);
		}
	}

	private showAddUserDialog(group: string) {
		this.setState({ group, user: '' });
	}

	private renderAddUserDialog() {
		const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
			this.setState({ user: event.target.value });
		};
		const handleClose = () => {
			this.setState({ user: undefined, group: undefined });
		};
		const handleAdd = async () => {
			const { user, group } = this.state;
			if (user === undefined || group === undefined) {
				return;
			}
			try {
				await api.addUser(group, user);
				this.setState(({ access }) => ({
					access: { ...access, [group]: [ ...access[group], user ].sort() },
					group: undefined,
					user: undefined,
				}));
			} catch (error) {
				this.props.onError(error);
			}
		};
		return (
			<Dialog open onClose={handleClose}>
				<DialogTitle>Add user to group "{this.state.group}"</DialogTitle>
				<DialogContent>
					<TextField placeholder="User" value={this.state.user} onChange={handleChange} autoFocus variant="outlined" margin="dense" fullWidth/>
				</DialogContent>
				<DialogActions>
					<Button color="primary" onClick={handleClose}>Cancel</Button>
					<Button color="primary" onClick={handleAdd} disabled={this.state.user!.length === 0}>OK</Button>
				</DialogActions>
			</Dialog>
		);
	}

	private async removeUser(group: string, user: string) {
		try {
			await api.removeUser(group, user);
			this.setState(({ access }) => ({
				access: { ...access, [group]: access[group].filter(u => u !== user) },
			}));
		} catch (error) {
			this.props.onError(error);
		}
	}

	private renderDialog() {
		if (this.state.group === undefined) {
			return null;
		} else if (this.state.user === undefined) {
			return this.renderCreateGroupDialog();
		} else {
			return this.renderAddUserDialog();
		}
	}

	render() {
		const { classes } = this.props;
		const { access } = this.state;

		return (
			<React.Fragment>
				<List>
					{Object.keys(access).sort().map(group => {
						return (
							<ListItem key={group} divider className={classes.group}>
								<Chip
									color="primary"
									label={group}
									className={classes.chip}
									onDelete={this.props.userIsRoot && access[group].length === 0 ? () => this.deleteGroup(group) : undefined}
								/>
								{access[group].map(user => (
									<Chip
										key={user}
										label={user}
										className={classes.chip}
										onDelete={this.props.userIsRoot ? () => this.removeUser(group, user) : undefined}
									/>
								))}
								{this.props.userIsRoot && (
									<IconButton className={classes.add} onClick={() => this.showAddUserDialog(group)}><AddIcon/></IconButton>
								)}
							</ListItem>
						);
					})}
					{this.props.userIsRoot && (
						<ListItem className={classes.group}>
							<IconButton className={classes.add} onClick={() => this.showCreateGroupDialog()}><AddIcon/></IconButton>
						</ListItem>
					)}
				</List>
				{this.renderDialog()}
			</React.Fragment>
		);
	}

}

export default withStyles(styles)(Access);
