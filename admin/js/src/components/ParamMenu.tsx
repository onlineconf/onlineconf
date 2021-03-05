import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { Theme, makeStyles } from '@material-ui/core/styles';
import Menu, { MenuProps } from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';

import NotesIcon from '@material-ui/icons/Notes';
import EditIcon from '@material-ui/icons/Edit';
import HistoryIcon from '@material-ui/icons/History';
import LockOpenIcon from '@material-ui/icons/LockOpen';
import NotificationsIcon from '@material-ui/icons/Notifications';
import PlaylistAddIcon from '@material-ui/icons/PlaylistAdd';
import RefreshIcon from '@material-ui/icons/Refresh';
import DeleteIcon from '@material-ui/icons/Delete';
import LowPriorityIcon from '@material-ui/icons/LowPriority';
import InsertCommentIcon from '@material-ui/icons/InsertComment';

import { IParamNode } from './common';
import WhoAmIContext from './WhoAmIContext';
import IconButtonProgress from './IconButtonProgress';

const useStyles = makeStyles((theme: Theme) => ({
	progress: {
		margin: -6,
		padding: 6,
		display: 'inline-flex',
	}
}), { name: 'ParamMenu' });

export interface ParamMenuProps extends Pick<MenuProps, 'anchorEl'> {
	anchorX?: number;
	param: IParamNode;
	onClose: () => void;
	onView: () => void;
	onEdit: () => void;
	onNotification: () => void;
	onAccess: () => void;
	onLog: () => void;
	onAddChild: () => void;
	onReload: () => void;
	onDelete: () => void;
	onMove: () => void;
	onDescribe: () => void;
}

export default function ParamMenu(props: ParamMenuProps) {
	const classes = useStyles();
	const { t } = useTranslation();
	const { userIsRoot } = React.useContext(WhoAmIContext);
	return (
		<Menu
			open
			onClose={props.onClose}
			anchorEl={props.anchorEl}
			anchorOrigin={{ horizontal: props.anchorX || 'right', vertical: 'top' }}
		>
			<MenuItem onClick={props.onReload} divider>
				<ListItemIcon>
					<IconButtonProgress loading={props.param.state === 'loading'} size={36} className={classes.progress}>
						<RefreshIcon/>
					</IconButtonProgress>
				</ListItemIcon>
				<ListItemText primary={t('param.menu.reload')}/>
			</MenuItem>

			<MenuItem onClick={props.onView}>
				<ListItemIcon><NotesIcon/></ListItemIcon>
				<ListItemText primary={t('param.menu.view')}/>
			</MenuItem>
			<MenuItem onClick={props.onLog} divider>
				<ListItemIcon>
					<IconButtonProgress loading={props.param.logLoading} size={36} className={classes.progress}>
						<HistoryIcon/>
					</IconButtonProgress>
				</ListItemIcon>
				<ListItemText primary={t('param.menu.history')}/>
			</MenuItem>

			<MenuItem onClick={props.onEdit} disabled={props.param.rw !== true}>
				<ListItemIcon><EditIcon/></ListItemIcon>
				<ListItemText primary={t('param.menu.edit')}/>
			</MenuItem>
			<MenuItem onClick={props.onDescribe} disabled={props.param.rw !== true}>
				<ListItemIcon><InsertCommentIcon/></ListItemIcon>
				<ListItemText primary={t('param.menu.describe')}/>
			</MenuItem>
			<MenuItem onClick={props.onMove} disabled={props.param.rw !== true}>
				<ListItemIcon><LowPriorityIcon/></ListItemIcon>
				<ListItemText primary={t('param.menu.move')}/>
			</MenuItem>
			<MenuItem onClick={props.onDelete} disabled={props.param.rw !== true || props.param.num_children !== 0} divider>
				<ListItemIcon><DeleteIcon/></ListItemIcon>
				<ListItemText primary={t('param.menu.delete')}/>
			</MenuItem>

			<MenuItem onClick={props.onAccess} disabled={props.param.rw !== true && !userIsRoot}>
				<ListItemIcon>
					<IconButtonProgress loading={props.param.accessLoading} size={36} className={classes.progress}>
						<LockOpenIcon/>
					</IconButtonProgress>
				</ListItemIcon>
				<ListItemText primary={t('param.menu.access')}/>
			</MenuItem>
			<MenuItem onClick={props.onNotification} disabled={props.param.rw !== true} divider>
				<ListItemIcon><NotificationsIcon/></ListItemIcon>
				<ListItemText primary={t('param.menu.notifications')}/>
			</MenuItem>

			<MenuItem onClick={props.onAddChild} disabled={props.param.rw !== true || props.param.mime === 'application/x-symlink'}>
				<ListItemIcon><PlaylistAddIcon/></ListItemIcon>
				<ListItemText primary={t('param.menu.create')}/>
			</MenuItem>
		</Menu>
	);
}
