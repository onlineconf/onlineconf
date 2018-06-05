import * as React from 'react';
import * as classNames from 'classnames';
import { withStyles, WithStyles, Theme, createStyles } from '@material-ui/core/styles';
import IconButton from '@material-ui/core/IconButton';

import ChevronRightIcon from '@material-ui/icons/ChevronRight';

import IconButtonProgress from './IconButtonProgress';
import { TreeNodeState } from './TreeNode';

const arrowStyles = (theme: Theme) => createStyles({
	button: {
		padding: 'inherit',
	},
	icon: {
		transition: theme.transitions.create('transform', {
			duration: theme.transitions.duration.shorter,
		}),
	},
	closedIcon: {
		transform: 'rotate(0deg)',
	},
	openIcon: {
		transform: 'rotate(90deg)',
	},
});

interface TreeNodeArrowProps {
	state: TreeNodeState;
	onOpen: () => void;
	onClose: () => void;
	className: string;
}

const TreeNodeArrow = (props: TreeNodeArrowProps & WithStyles<typeof arrowStyles>) => {
	const { state, classes, onOpen, onClose, className } = props;

	if (state === 'leaf') {
		return <span className={className}/>;
	}

	const iconClassName = classNames(classes.icon, {
		[classes.openIcon]: state === 'open',
		[classes.closedIcon]: state !== 'open',
	});

	return (
		<IconButtonProgress size={24} loading={state === 'loading'} className={className}>
			<IconButton
				onClick={state === 'open' ? onClose : state === 'closed' ? onOpen : undefined}
				className={classes.button}
			>
				<ChevronRightIcon className={iconClassName}/>
			</IconButton>
		</IconButtonProgress>
	);
};

export default withStyles(arrowStyles)(TreeNodeArrow);
