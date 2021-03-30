import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { createStyles, WithStyles, withStyles, Theme } from '@material-ui/core/styles';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import Chip from '@material-ui/core/Chip';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import IconButton from '@material-ui/core/IconButton';

import AddIcon from '@material-ui/icons/AddCircle';

import * as api from '../api';
import Avatar from './Avatar';
import UserField from './UserField';
import WhoAmIContext from './WhoAmIContext';

const styles = (theme: Theme) => createStyles({
	group: {
		display: 'flex',
		flexWrap: 'wrap',
		padding: theme.spacing(0.5, 1),
	},
	chip: {
		margin: theme.spacing(0.5),
	},
	add: {
		padding: 4,
	},
});

interface AccessProps {
	onError: (error: Error) => void;
}

interface AccessState {
	access: { [group: string]: string[] };
	group?: string;
	user?: string;
}

class Access extends React.Component<AccessProps & WithStyles<typeof styles> & WithTranslation, AccessState> {

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
		const { t } = this.props;
		return (
			<Dialog open onClose={handleClose}>
				<DialogTitle>{t('access.createGroup')}</DialogTitle>
				<DialogContent>
					<TextField placeholder={t('access.group')} value={this.state.group} onChange={handleChange} autoFocus variant="outlined" margin="dense" fullWidth/>
				</DialogContent>
				<DialogActions>
					<Button color="primary" onClick={handleClose}>{t('button.cancel')}</Button>
					<Button color="primary" onClick={handleCreate} disabled={this.state.group.length === 0}>{t('button.ok')}</Button>
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
		if (this.state.user === undefined) {
			return;
		}
		const handleChange = (value: string) => {
			this.setState({ user: value });
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
		const { t } = this.props;
		return (
			<Dialog open onClose={handleClose}>
				<DialogTitle>{t('access.addUser', { group: this.state.group })}</DialogTitle>
				<DialogContent>
					<UserField placeholder={t('access.user')} value={this.state.user} onChange={handleChange} autoFocus variant="outlined" margin="dense" fullWidth/>
				</DialogContent>
				<DialogActions>
					<Button color="primary" onClick={handleClose}>{t('button.cancel')}</Button>
					<Button color="primary" onClick={handleAdd} disabled={this.state.user!.length === 0}>{t('button.ok')}</Button>
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

	static contextType = WhoAmIContext;

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
									onDelete={this.context.userIsRoot && access[group].length === 0 ? () => this.deleteGroup(group) : undefined}
								/>
								{access[group].map(user => (
									<Chip
										key={user}
										label={user}
										avatar={<Avatar username={user} disableTooltip/>}
										className={classes.chip}
										onDelete={this.context.userIsRoot ? () => this.removeUser(group, user) : undefined}
									/>
								))}
								{this.context.userIsRoot && (
									<IconButton className={classes.add} onClick={() => this.showAddUserDialog(group)}><AddIcon/></IconButton>
								)}
							</ListItem>
						);
					})}
					{this.context.userIsRoot && (
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

export default withTranslation()(withStyles(styles)(Access));
