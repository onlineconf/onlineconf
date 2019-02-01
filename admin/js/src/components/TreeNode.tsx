import * as React from 'react';
import * as classNames from 'classnames';
import scrollIntoView from 'scroll-into-view-if-needed';
import { withStyles, WithStyles, Theme, createStyles } from '@material-ui/core/styles';
import Collapse from '@material-ui/core/Collapse';

import Tree from './Tree';
import TreeNodeArrow from './TreeNodeArrow';

export type TreeNodeState = 'open' | 'closed' | 'leaf' | 'loading';

const styles = (theme: Theme) => createStyles({
	root: {},
	externalContainer: {
		display: 'flex',
		justifyContent: 'flex-start',
		alignItems: 'center',
		textDecoration: 'none',
		width: '100%',
		boxSizing: 'border-box',
		textAlign: 'left',
	},
	arrow: {
		flex: 'none',
		margin: theme.spacing.unit,
		padding: 12,
		minWidth: 48,
	},
	internalContainer: {
		flex: 'auto',
		overflow: 'hidden',
		padding: theme.spacing.unit,
		cursor: 'pointer',
		transition: theme.transitions.create('background-color', {
			duration: theme.transitions.duration.shortest,
		}),
		'&:hover': {
			backgroundColor: theme.palette.action.hover,
		},
	},
	divider: {
		borderTop: `1px solid ${theme.palette.divider}`,
		backgroundClip: 'padding-box',
	},
	children: {
		paddingLeft: 48 + theme.spacing.unit,
	},
	selected: {
		'& > $externalContainer > $internalContainer': {
			backgroundColor: theme.palette.action.selected,
		},
	},
});

interface TreeNodeProps {
	component?: React.ReactType;
	className?: string;
	item: React.ReactNode;
	children?: React.ReactNode;
	state: TreeNodeState;
	selected?: boolean;
	disabled?: boolean;
	divider?: boolean;
	onOpen: () => void;
	onClose: () => void;
	onClick?: React.MouseEventHandler<{}>;
	onDoubleClick?: React.MouseEventHandler<{}>;
	onContextMenu?: React.MouseEventHandler<{}>;
}

class TreeNode extends React.Component<TreeNodeProps & WithStyles<typeof styles>> {

	ecRef = React.createRef<HTMLDivElement>();

	componentDidMount() {
		if (this.props.selected) {
			scrollIntoView(this.ecRef.current!, { scrollMode: 'if-needed' });
		}
	}

	componentDidUpdate(prevProps: TreeNodeProps) {
		if (this.props.selected && !prevProps.selected) {
			scrollIntoView(this.ecRef.current!, { scrollMode: 'if-needed' });
		}
	}

	render() {
		const { item, children, state, onOpen, onClose, onClick, onDoubleClick, onContextMenu, component, classes, className: classNameProp, selected, disabled, divider, ...rest } = this.props;

		const Component: any = component || 'li';
		const className = classNames(classes.root, { [classes.selected]: selected, [classes.divider]: divider }, classNameProp);
		const componentProps = { className, disabled, ...rest };

		return (
			<Component {...componentProps}>
				<div className={classes.externalContainer} ref={this.ecRef}>
					<TreeNodeArrow state={state} onOpen={onOpen} onClose={onClose} className={classes.arrow} />
					<div className={classes.internalContainer} onClick={onClick} onDoubleClick={onDoubleClick} onContextMenu={onContextMenu}>
						{item}
					</div>
				</div>
				{state !== 'leaf' && (
					<Collapse in={state === 'open'} timeout="auto" unmountOnExit>
						<Tree disablePadding className={classes.children}>
							{children}
						</Tree>
					</Collapse>
				)}
			</Component>
		);
	}

}

export default withStyles(styles)(TreeNode);
