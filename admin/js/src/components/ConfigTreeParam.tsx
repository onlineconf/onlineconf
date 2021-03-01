import * as React from 'react';
import clsx from 'clsx';
import * as CSS from 'csstype';
import { Omit } from '@material-ui/types';
import { Theme, makeStyles, createStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';
import IconButton from '@material-ui/core/IconButton';
import Typography from '@material-ui/core/Typography';
import { SvgIconProps } from '@material-ui/core/SvgIcon';

import DefaultIcon from '@material-ui/icons/Settings';
import RootIcon from '@material-ui/icons/Language';
import TopLevelIcon from '@material-ui/icons/Domain';
import NullIcon from '@material-ui/icons/Block';
import FolderIcon from '@material-ui/icons/FolderOpen';
import TextIcon from '@material-ui/icons/Subject';
import StructIcon from '@material-ui/icons/DeviceHub';
import SymlinkIcon from '@material-ui/icons/Launch';
import CaseIcon from '@material-ui/icons/Help';
import TemplateIcon from '@material-ui/icons/LocalAtm';
import ListIcon from '@material-ui/icons/List';
import LockOpen from '@material-ui/icons/LockOpen';
import MoreHoriz from '@material-ui/icons/MoreHoriz';
import Notifications from '@material-ui/icons/Notifications';

import { ParamType } from '../api';
import { IParamNode } from './common';
import { ValuePreview } from './value';
import IconButtonProgress from './IconButtonProgress';
import ButtonProgress from './ButtonProgress';
import ParamMenu, { ParamMenuProps } from './ParamMenu';
import NoAccess from './NoAccess';

export const spacingUnit = 2;
export const buttonSize = 24;
export const iconButtonPadding = (buttonSize - 24) / 2;

const iconByType: { [P in ParamType]: React.ComponentType<SvgIconProps> } = {
	'application/x-null': NullIcon,
	'text/plain': TextIcon,
	'application/json': StructIcon,
	'application/x-yaml': StructIcon,
	'application/x-symlink': SymlinkIcon,
	'application/x-case': CaseIcon,
	'application/x-template': TemplateIcon,
	'application/x-list': ListIcon,
	'application/x-server': ListIcon,
	'application/x-server2': ListIcon,
};

const usePreviewStyles = makeStyles((theme: Theme) => {
	const overflow: CSS.Properties = {
		overflow: 'hidden',
		textOverflow: 'ellipsis',
		whiteSpace: 'nowrap',
	};

	return createStyles({
		root: {
			display: 'flex',
			alignItems: 'center',
		},
		matched: {
			fontWeight: 'bolder',
		},
		icon: {
			flex: 'none',
		},
		name: {
			flex: 'auto',
			marginLeft: spacingUnit,
			marginRight: spacingUnit,
			...overflow
		},
		summary: {
			color: 'gray',
			'&::before': {
				content: '" ("',
			},
			'&::after': {
				content: '")"',
			},
		},
		valueContainer: {
			flex: 'none',
			width: '5em',
			display: 'flex',
			alignItems: 'center',
			'@media (min-width: 400px)': { width: '10em' },
			'@media (min-width: 500px)': { width: '15em' },
			'@media (min-width: 600px)': { width: '20em' },
			'@media (min-width: 960px)': { width: '25em' },
			'@media (min-width: 1280px)': { width: '40em' },
			'@media (min-width: 1920px)': { width: '50em' },
		},
		value: {
			flex: 'auto',
			marginLeft: spacingUnit,
			marginRight: spacingUnit,
			...overflow
		},
		version: {
			width: '2em',
			...overflow
		},
		mtime: {
			width: '10em',
			...overflow
		},
		logButtons: {
			flex: 'none',
			'@media (max-width: 800px)': {
				display: 'none',
			},
		},
		button: {
			minWidth: buttonSize,
			minHeight: buttonSize,
			padding: `0 ${spacingUnit}px`,
		},
		iconButton: {
			padding: iconButtonPadding,
		},
	});
});

interface ConfigTreeParamPreviewProps {
	param: IParamNode;
	userIsRoot: boolean;
	onMenuOpen: () => void;
	onEdit: () => void;
	onNotification: () => void;
	onAccess: () => void;
	onLog: () => void;
	onValuePopoverOpen: React.MouseEventHandler<HTMLElement>;
	onValuePopoverClose: () => void;
}

function ConfigTreeParamPreview(props: ConfigTreeParamPreviewProps) {
	const { param, onMenuOpen: onViewOpen, onLog, onAccess, onNotification, onValuePopoverOpen, onValuePopoverClose } = props;
	const classes = usePreviewStyles();
	let Icon = iconByType[param.mime] || DefaultIcon;

	if (param.mime === 'application/x-null') {
		if (param.path === '/') {
			Icon = RootIcon;
		} else if (param.path === '/' + param.name) {
			Icon = TopLevelIcon;
		} else if (param.num_children > 0) {
			Icon = FolderIcon;
		}
	}

	const className = clsx(classes.root, { [classes.matched]: param.match });

	return (
		<Typography component="div" variant="body2" className={className}>
			<Icon className={classes.icon} color="action"/>
			<div className={classes.name}>
				{param.name}
				{param.summary !== '' && <span className={classes.summary}>{param.summary}</span>}
			</div>
			<div className={classes.valueContainer}>
				<div className={classes.value}>
					{param.rw === null ? <NoAccess/> : (
						<span onMouseEnter={onValuePopoverOpen} onMouseLeave={onValuePopoverClose}>
							<ValuePreview type={param.mime} value={param.data}/>
						</span>
					)}
				</div>
				{param.notification_modified && (
					<IconButton onClick={onNotification} disabled={param.rw !== true} className={classes.iconButton}><Notifications/></IconButton>
				)}
				{param.access_modified && (
					<IconButtonProgress size={buttonSize} loading={param.accessLoading}>
						<IconButton onClick={onAccess} disabled={param.rw !== true && !props.userIsRoot} className={classes.iconButton}><LockOpen/></IconButton>
					</IconButtonProgress>
				)}
			</div>
			<ButtonProgress loading={param.logLoading} className={classes.logButtons}>
				<Button onClick={onLog} classes={{ root: classes.button, label: classes.version }}>{param.version}</Button>
				<Button onClick={onLog} classes={{ root: classes.button, label: classes.mtime }}>{param.mtime}</Button>
			</ButtonProgress>
			<IconButton onClick={onViewOpen} className={classes.iconButton}><MoreHoriz/></IconButton>
		</Typography>
	);
}

interface ConfigTreeParamProps extends Omit<ParamMenuProps, 'onClose' | 'anchorEl'> {
	param: IParamNode;
	userIsRoot: boolean;
	menu?: string;
	menuAnchorX?: number;
	onMenuOpen: () => void;
	onMenuClose: () => void;
	onValuePopoverOpen: React.MouseEventHandler<HTMLElement>;
	onValuePopoverClose: () => void;
}

export default class ConfigTreeParam extends React.Component<ConfigTreeParamProps> {

	anchorEl = React.createRef<HTMLDivElement>();

	render() {
		const { onEdit, onLog, onAddChild, onAccess, onNotification, onValuePopoverOpen, onValuePopoverClose, param } = this.props;

		return (
			<div ref={this.anchorEl}>
				{this.props.menu === param.path && (
					<ParamMenu
						anchorEl={this.anchorEl.current}
						anchorX={this.props.menuAnchorX}
						param={param}
						userIsRoot={this.props.userIsRoot}
						onClose={this.props.onMenuClose}
						onView={this.props.onView}
						onEdit={onEdit}
						onDescribe={this.props.onDescribe}
						onNotification={onNotification}
						onAccess={onAccess}
						onLog={onLog}
						onAddChild={onAddChild}
						onReload={this.props.onReload}
						onDelete={this.props.onDelete}
						onMove={this.props.onMove}
					/>
				)}
				<ConfigTreeParamPreview
					param={param}
					userIsRoot={this.props.userIsRoot}
					onMenuOpen={this.props.onMenuOpen}
					onEdit={onEdit}
					onNotification={onNotification}
					onAccess={onAccess}
					onLog={onLog}
					onValuePopoverOpen={onValuePopoverOpen}
					onValuePopoverClose={onValuePopoverClose}
				/>
			</div>
		);
	}

}
