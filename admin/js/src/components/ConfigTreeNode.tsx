import * as React from 'react';
import { Theme, withStyles, WithStyles, createStyles } from '@material-ui/core/styles';

import { IParamNode, smartCompare } from './common';
import ConfigTreeParam, { buttonSize, spacingUnit, iconButtonPadding } from './ConfigTreeParam';
import TreeNode from './TreeNode';

const styles = (theme: Theme) => createStyles({
	internalContainer: {
		padding: spacingUnit,
	},
	arrow: {
		margin: spacingUnit,
		padding: iconButtonPadding,
		minWidth: buttonSize,
	},
	children: {
		paddingLeft: buttonSize + 2 * spacingUnit,
	},
});

interface ConfigTreeNodeProps {
	param: IParamNode;
	menu?: string;
	onOpen: (path: string) => void;
	onClose: (path: string) => void;
	onSelect: (param: IParamNode) => void;
	onMenuOpen: (path: string) => void;
	onMenuClose: (path: string) => void;
	onView: (param: IParamNode) => void;
	onEdit: (param: IParamNode) => void;
	onMove: (param: IParamNode) => void;
	onDescribe: (param: IParamNode) => void;
	onNotification: (param: IParamNode) => void;
	onAccess: (path: string) => void;
	onLog: (path: string) => void;
	onAddChild: (param: IParamNode) => void;
	onReload: (param: IParamNode) => void;
	onDelete: (param: IParamNode) => void;
	onValuePopoverOpen: (event: React.MouseEvent<{}>, param: IParamNode) => void;
	onValuePopoverClose: () => void;
}

let ConfigTreeNode: React.ComponentType<ConfigTreeNodeProps>;
ConfigTreeNode = (props: ConfigTreeNodeProps & WithStyles<typeof styles>) => {
	const { param, ...rest } = props;
	const { onOpen, onClose, onSelect, onEdit, onLog, onAddChild, onAccess, onNotification, onValuePopoverOpen, onValuePopoverClose, classes } = rest;

	if (param.hidden) {
		return null;
	}

	return (
		<TreeNode
			state={param.forceOpen ? 'open' : param.state}
			selected={param.selected}
			onOpen={() => onOpen(param.path)}
			onClose={() => onClose(param.path)}
			onClick={event => { if (!param.selected) { onSelect(param); } }}
			onDoubleClick={event => onEdit(param)}
			item={<ConfigTreeParam
				param={param}
				menu={props.menu}
				onMenuOpen={() => props.onMenuOpen(param.path)}
				onMenuClose={() => props.onMenuClose(param.path)}
				onView={() => props.onView(param)}
				onEdit={() => onEdit(param)}
				onDescribe={() => props.onDescribe(param)}
				onNotification={() => onNotification(param)}
				onAccess={() => onAccess(param.path)}
				onLog={() => onLog(param.path)}
				onAddChild={() => onAddChild(param)}
				onReload={() => props.onReload(param)}
				onDelete={() => props.onDelete(param)}
				onMove={() => props.onMove(param)}
				onValuePopoverOpen={event => onValuePopoverOpen(event, param)}
				onValuePopoverClose={onValuePopoverClose}
			/>}
			classes={classes}
			divider
		>
			{param.children && Object.keys(param.children).sort(smartCompare).map(name => {
				const child = { ...param!.children![name] };
				if (!(param.forceOpen || param.state === 'open')) {
					child.children = undefined;
				}
				return <ConfigTreeNode key={name} param={child} {...rest}/>;
			})}
		</TreeNode>
	);
};

export default ConfigTreeNode = withStyles(styles)(ConfigTreeNode);
