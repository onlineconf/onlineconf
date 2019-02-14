import * as React from 'react';
import { Theme, createStyles, withStyles, WithStyles, Menu, MenuItem, ListItemIcon, ListItemText } from '@material-ui/core';

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
import IconButtonProgress from './IconButtonProgress';

const styles = (theme: Theme) => createStyles({
	progress: {
		margin: -6,
		padding: 6,
		display: 'inline-flex',
	}
});

export interface ParamMenuProps {
	anchorEl: HTMLElement | ((elem: HTMLElement) => HTMLElement) | null;
	anchorX?: number;
	param: IParamNode;
	userIsRoot: boolean;
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

const ParamMenu = (props: ParamMenuProps & WithStyles<typeof styles>) => (
	<Menu
		open
		onClose={props.onClose}
		anchorEl={props.anchorEl}
		anchorOrigin={{ horizontal: props.anchorX || 'right', vertical: 'top' }}
	>
		<MenuItem onClick={props.onReload} divider>
			<ListItemIcon>
				<IconButtonProgress loading={props.param.state === 'loading'} size={36} className={props.classes.progress}>
					<RefreshIcon/>
				</IconButtonProgress>
			</ListItemIcon>
			<ListItemText primary="Reload"/>
		</MenuItem>

		<MenuItem onClick={props.onView}>
			<ListItemIcon><NotesIcon/></ListItemIcon>
			<ListItemText primary="View"/>
		</MenuItem>
		<MenuItem onClick={props.onLog} divider>
			<ListItemIcon>
				<IconButtonProgress loading={props.param.logLoading} size={36} className={props.classes.progress}>
					<HistoryIcon/>
				</IconButtonProgress>
			</ListItemIcon>
			<ListItemText primary="History"/>
		</MenuItem>

		<MenuItem onClick={props.onEdit} disabled={props.param.rw !== true}>
			<ListItemIcon><EditIcon/></ListItemIcon>
			<ListItemText primary="Edit"/>
		</MenuItem>
		<MenuItem onClick={props.onDescribe} disabled={props.param.rw !== true}>
			<ListItemIcon><InsertCommentIcon/></ListItemIcon>
			<ListItemText primary="Describe"/>
		</MenuItem>
		<MenuItem onClick={props.onMove} disabled={props.param.rw !== true}>
			<ListItemIcon><LowPriorityIcon/></ListItemIcon>
			<ListItemText primary="Move"/>
		</MenuItem>
		<MenuItem onClick={props.onDelete} disabled={props.param.rw !== true || props.param.num_children !== 0} divider>
			<ListItemIcon><DeleteIcon/></ListItemIcon>
			<ListItemText primary="Delete"/>
		</MenuItem>

		<MenuItem onClick={props.onAccess} disabled={props.param.rw !== true && !props.userIsRoot}>
			<ListItemIcon>
				<IconButtonProgress loading={props.param.accessLoading} size={36} className={props.classes.progress}>
					<LockOpenIcon/>
				</IconButtonProgress>
			</ListItemIcon>
			<ListItemText primary="Access"/>
		</MenuItem>
		<MenuItem onClick={props.onNotification} disabled={props.param.rw !== true} divider>
			<ListItemIcon><NotificationsIcon/></ListItemIcon>
			<ListItemText primary="Notifications"/>
		</MenuItem>

		<MenuItem onClick={props.onAddChild} disabled={props.param.rw !== true || props.param.mime === 'application/x-symlink'}>
			<ListItemIcon><PlaylistAddIcon/></ListItemIcon>
			<ListItemText primary="Create"/>
		</MenuItem>
	</Menu>
);

export default withStyles(styles)(ParamMenu);
