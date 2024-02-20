import * as React from 'react';
import clsx from 'clsx';
import { Theme } from '@mui/material/styles';
import { WithStyles } from '@mui/styles';
import withStyles from '@mui/styles/withStyles';
import createStyles from '@mui/styles/createStyles';
import IconButton from '@mui/material/IconButton';

import ChevronRightIcon from '@mui/icons-material/ChevronRight';

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

	const iconClassName = clsx(classes.icon, {
		[classes.openIcon]: state === 'open',
		[classes.closedIcon]: state !== 'open',
	});

	return (
		<IconButtonProgress size={24} loading={state === 'loading'} className={className}>
			<IconButton
				onClick={state === 'open' ? onClose : state === 'closed' ? onOpen : undefined}
				className={classes.button}
				size="large">
				<ChevronRightIcon className={iconClassName}/>
			</IconButton>
		</IconButtonProgress>
	);
};

export default withStyles(arrowStyles)(TreeNodeArrow);
