import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { Theme } from '@mui/material/styles';
import makeStyles from '@mui/styles/makeStyles';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Chip from '@mui/material/Chip';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import IconButton from '@mui/material/IconButton';

import AddIcon from '@mui/icons-material/AddCircle';

import * as api from '../api';
import Avatar from './Avatar';
import UserField from './UserField';
import WhoAmIContext from './WhoAmIContext';

const useStyles = makeStyles((theme: Theme) => ({
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
}));

interface AccessListProps {
	access: { [group: string]: string[] };
	onCreateGroup(): void;
	onDeleteGroup(group: string): void;
	onAddUser(group: string): void;
	onRemoveUser(group: string, user: string): void;
}

function AccessList(props: AccessListProps) {
	const classes = useStyles();
	const { userIsRoot } = React.useContext(WhoAmIContext);
	return (
		<List>
			{Object.keys(props.access).sort().map(group => {
				return (
					<ListItem key={group} divider className={classes.group}>
						<Chip
							color="primary"
							label={group}
							className={classes.chip}
							onDelete={userIsRoot && props.access[group].length === 0 ? () => props.onDeleteGroup(group) : undefined}
						/>
						{props.access[group].map(user => (
							<Chip
								key={user}
								label={user}
								avatar={<Avatar username={user} disableTooltip/>}
								className={classes.chip}
								onDelete={userIsRoot ? () => props.onRemoveUser(group, user) : undefined}
							/>
						))}
						{userIsRoot && (
							<IconButton
								className={classes.add}
								onClick={() => props.onAddUser(group)}
								size="large"><AddIcon/></IconButton>
						)}
					</ListItem>
				);
			})}
			{userIsRoot && (
				<ListItem className={classes.group}>
					<IconButton className={classes.add} onClick={props.onCreateGroup} size="large"><AddIcon/></IconButton>
				</ListItem>
			)}
		</List>
	);
}

interface CreateGroupDialogProps {
	onSubmit: (group: string) => void;
	onClosed: () => void;
}

function CreateGroupDialog(props: CreateGroupDialogProps) {
	const { t } = useTranslation();
	const [ open, setOpen ] = React.useState(true);
	const [ group, setGroup ] = React.useState('');
	return (
		<Dialog open={open} onClose={() => setOpen(false)} TransitionProps={{ onExited: props.onClosed }}>
			<DialogTitle>{t('access.createGroup')}</DialogTitle>
			<DialogContent>
				<TextField placeholder={t('access.group')} value={group} onChange={event => setGroup(event.target.value)} autoFocus variant="outlined" margin="dense" fullWidth/>
			</DialogContent>
			<DialogActions>
				<Button color="primary" onClick={() => setOpen(false)}>{t('button.cancel')}</Button>
				<Button color="primary" onClick={() => props.onSubmit(group)} disabled={group.length === 0}>{t('button.ok')}</Button>
			</DialogActions>
		</Dialog>
	);
}

interface AddUserDialogProps {
	group: string;
	onSubmit: (group: string, user: string) => void;
	onClosed: () => void;
}

function AddUserDialog(props: AddUserDialogProps) {
	const { t } = useTranslation();
	const [ open, setOpen ] = React.useState(true);
	const [ user, setUser ] = React.useState('');
	return (
		<Dialog open={open} onClose={() => setOpen(false)} TransitionProps={{ onExited: props.onClosed}}>
			<DialogTitle>{t('access.addUser', { group: props.group })}</DialogTitle>
			<DialogContent>
				<UserField placeholder={t('access.user')} value={user} onChange={setUser} autoFocus variant="outlined" margin="dense" fullWidth/>
			</DialogContent>
			<DialogActions>
				<Button color="primary" onClick={() => setOpen(false)}>{t('button.cancel')}</Button>
				<Button color="primary" onClick={() => props.onSubmit(props.group, user)} disabled={user.length === 0}>{t('button.ok')}</Button>
			</DialogActions>
		</Dialog>
	);
}

interface AccessProps {
	onError: (error: unknown) => void;
}

interface AccessState {
	access: { [group: string]: string[] };
	dialog?: JSX.Element;
}

export default class Access extends React.Component<AccessProps, AccessState> {

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

	private showCreateGroupDialog = () => {
		this.setState({
			dialog: <CreateGroupDialog onSubmit={this.createGroup} onClosed={this.handleDialogClosed}/>
		});
	};

	private createGroup = async (group: string) => {
		try {
			await api.createGroup(group);
			this.setState(({ access }) => ({
				access: { [group]: [], ...access },
				dialog: undefined,
			}));
		} catch (error) {
			this.props.onError(error);
		}
	};

	private deleteGroup = async (group: string) => {
		try {
			await api.deleteGroup(group);
			this.setState(({ access }) => {
				delete(access[group]);
				return { access };
			});
		} catch (error) {
			this.props.onError(error);
		}
	};

	private showAddUserDialog = (group: string) => {
		this.setState({
			dialog: <AddUserDialog group={group} onSubmit={this.addUser} onClosed={this.handleDialogClosed}/>
		});
	};

	private addUser = async (group: string, user: string) => {
		try {
			await api.addUser(group, user);
			this.setState(({ access }) => ({
				access: { ...access, [group]: [ ...access[group], user ].sort() },
				dialog: undefined,
			}));
		} catch (error) {
			this.props.onError(error);
		}
	};

	private removeUser = async (group: string, user: string) => {
		try {
			await api.removeUser(group, user);
			this.setState(({ access }) => ({
				access: { ...access, [group]: access[group].filter(u => u !== user) },
			}));
		} catch (error) {
			this.props.onError(error);
		}
	};

	private handleDialogClosed = () => {
		this.setState({ dialog: undefined });
	};

	static contextType = WhoAmIContext;

	render() {
		return (
			<React.Fragment>
				<AccessList
					access={this.state.access}
					onCreateGroup={this.showCreateGroupDialog}
					onDeleteGroup={this.deleteGroup}
					onAddUser={this.showAddUserDialog}
					onRemoveUser={this.removeUser}
				/>
				{this.state.dialog}
			</React.Fragment>
		);
	}

}
