import * as React from 'react';
import clsx from 'clsx';
import * as CSS from 'csstype';
import { DistributiveOmit } from '@mui/types';
import { Theme } from '@mui/material/styles';
import makeStyles from '@mui/styles/makeStyles';
import createStyles from '@mui/styles/createStyles';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';

import LockOpen from '@mui/icons-material/LockOpen';
import MoreHoriz from '@mui/icons-material/MoreHoriz';
import Notifications from '@mui/icons-material/Notifications';

import { IParamNode } from './common';
import { ValuePreview } from './value';
import IconButtonProgress from './IconButtonProgress';
import ButtonProgress from './ButtonProgress';
import ParamMenu, { ParamMenuProps } from './ParamMenu';
import NoAccess from './NoAccess';
import WhoAmIContext from './WhoAmIContext';
import ParamIcon from './ParamIcon';

export const spacingUnit = 2;
export const buttonSize = 24;
export const iconButtonPadding = (buttonSize - 24) / 2;

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
			display: 'flex',
			alignItems: 'center',
			overflow: 'hidden',
			justifyContent: 'flex-end',
			width: 'calc(min(50vw - var(--left-menu-width, 0px) / 2, calc(100vw - (var(--left-menu-width, 0px) + (var(--max-depth, 0) + 1) * 28px + 10ch))) - 28px)',
		},
		value: {
			flex: 'auto',
			marginLeft: spacingUnit,
			marginRight: spacingUnit,
			...overflow
		},
		extraButtons: {
			'@media (max-width: 400px)': {
				display: 'none',
			},
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
}, { name: 'ConfigTreeParamPreview' });

interface ConfigTreeParamPreviewProps {
	param: IParamNode;
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
	const { userIsRoot } = React.useContext(WhoAmIContext);

	const className = clsx(classes.root, { [classes.matched]: param.match });

	return (
		<Typography component="div" variant="body2" className={className}>
			<ParamIcon type={param.mime} path={param.path} numChildren={param.num_children} className={classes.icon} color="action"/>
			<div className={classes.name}>
				{param.name}
				{param.summary !== '' && <span className={classes.summary}>{param.summary}</span>}
			</div>
			<div className={clsx(classes.valueContainer, 'tree-value-column')}>
				<div className={classes.value}>
					{param.rw === null ? <NoAccess/> : (
						<span onMouseEnter={onValuePopoverOpen} onMouseLeave={onValuePopoverClose}>
							<ValuePreview type={param.mime} value={param.data}/>
						</span>
					)}
				</div>
				{param.notification_modified && (
					<IconButton
						onClick={onNotification}
						disabled={param.rw !== true}
						className={clsx(classes.iconButton, classes.extraButtons)}
						size="large"><Notifications/></IconButton>
				)}
				{param.access_modified && (
					<IconButtonProgress size={buttonSize} loading={param.accessLoading} className={classes.extraButtons}>
						<IconButton
							onClick={onAccess}
							disabled={param.rw !== true && !userIsRoot}
							className={classes.iconButton}
							size="large"><LockOpen/></IconButton>
					</IconButtonProgress>
				)}
				<ButtonProgress loading={param.logLoading} className={classes.logButtons}>
					<Button onClick={onLog} classes={{ root: classes.button}}>{param.version}</Button>
					<Button onClick={onLog} classes={{ root: classes.button}}>{param.mtime}</Button>
				</ButtonProgress>
			</div>
			<IconButton onClick={onViewOpen} className={classes.iconButton} size="large"><MoreHoriz/></IconButton>
		</Typography>
	);
}

interface ConfigTreeParamProps extends DistributiveOmit<ParamMenuProps, 'onClose' | 'anchorEl'> {
	param: IParamNode;
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
