import * as React from 'react';
import clsx from 'clsx';
import { createStyles, withStyles, WithStyles } from '@material-ui/core/styles';
import CircularProgress from '@material-ui/core/CircularProgress';

const styles = createStyles({
	wrapper: {
		position: 'relative',
	},
	progress: {
		position: 'absolute',
		top: 0,
		left: 0,
		zIndex: 1,
	},
	inheritSize: {
		width: 'inherit',
		height: 'inherit',
	},
});

interface IconButtonProgressProps {
	size?: number | 'inherit';
	loading?: boolean;
	className?: string;
	children: React.ReactNode;
}

const IconButtonProgress = (props: IconButtonProgressProps & WithStyles<typeof styles>) => {
	const progress = props.size === 'inherit'
		? <CircularProgress className={clsx(props.classes.progress, props.classes.inheritSize)} style={{ width: undefined, height: undefined }}/>
		: <CircularProgress className={props.classes.progress} size={props.size || 48}/>;
	return (
		<div className={clsx(props.className, props.classes.wrapper)}>
			{props.children}
			{props.loading && progress}
		</div>
	);
};

export default withStyles(styles)(IconButtonProgress);
