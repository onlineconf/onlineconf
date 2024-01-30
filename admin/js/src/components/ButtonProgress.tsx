import * as React from 'react';
import clsx from 'clsx';
import { WithStyles } from '@mui/styles';
import createStyles from '@mui/styles/createStyles';
import withStyles from '@mui/styles/withStyles';
import LinearProgress from '@mui/material/LinearProgress';

const styles = createStyles({
	wrapper: {
		position: 'relative',
	},
	progress: {
		marginTop: -5,
		zIndex: 1,
		borderRadius: 2,
	},
	over: {
		position: 'absolute',
		top: 0,
		left: 0,
		right: 0,
		bottom: -5,
		zIndex: 1,
	},
});

interface ButtonProgressProps {
	size?: number | 'inherit';
	loading?: boolean;
	className?: string;
	children: React.ReactNode;
}

const ButtonProgress = (props: ButtonProgressProps & WithStyles<typeof styles>) => (
	<div className={clsx(props.className, props.classes.wrapper)}>
		{props.children}
		{props.loading && (
			<React.Fragment>
				<div className={props.classes.over} />
				<LinearProgress className={props.classes.progress} />
			</React.Fragment>
		)}
	</div>
);

export default withStyles(styles)(ButtonProgress);
