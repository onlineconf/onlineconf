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
	userIsRoot: boolean;
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

	if (param.hidden) {
		return null;
	}

	return (
		<TreeNode
			state={param.forceOpen ? 'open' : param.state}
			selected={param.selected}
			onOpen={() => props.onOpen(param.path)}
			onClose={() => props.onClose(param.path)}
			onClick={() => { if (!param.selected) { props.onSelect(param); } }}
			onDoubleClick={() => param.rw === true ? props.onEdit(param) : props.onView(param)}
			onContextMenu={event => {
				event.preventDefault();
				if (!param.selected) {
					props.onSelect(param);
				}
				props.onMenuOpen(param.path);
			}}
			item={<ConfigTreeParam
				param={param}
				userIsRoot={props.userIsRoot}
				menu={props.menu}
				onMenuOpen={() => props.onMenuOpen(param.path)}
				onMenuClose={() => props.onMenuClose(param.path)}
				onView={() => props.onView(param)}
				onEdit={() => props.onEdit(param)}
				onDescribe={() => props.onDescribe(param)}
				onNotification={() => props.onNotification(param)}
				onAccess={() => props.onAccess(param.path)}
				onLog={() => props.onLog(param.path)}
				onAddChild={() => props.onAddChild(param)}
				onReload={() => props.onReload(param)}
				onDelete={() => props.onDelete(param)}
				onMove={() => props.onMove(param)}
				onValuePopoverOpen={event => props.onValuePopoverOpen(event, param)}
				onValuePopoverClose={props.onValuePopoverClose}
			/>}
			classes={props.classes}
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
