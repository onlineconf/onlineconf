import * as React from 'react';
import clsx from 'clsx';
import { StandardProps, createStyles } from '@material-ui/core';
import { withStyles, WithStyles, Theme } from '@material-ui/core/styles';

const styles = (theme: Theme) => createStyles({
	root: {
		listStyle: 'none',
		margin: 0,
		padding: 0,
	},
	padding: {
		paddingTop: theme.spacing(1),
		paddingBottom: theme.spacing(1),
	},
});

type TreeClassKey = keyof ReturnType<typeof styles>;

interface TreeProps extends StandardProps<React.HTMLAttributes<HTMLUListElement>, TreeClassKey> {
	component?: React.ElementType;
	disablePadding?: boolean;
}

function Tree(props: TreeProps & WithStyles<TreeClassKey>) {
	const { children, classes, className: classNameProp, component, disablePadding, ...rest } = props;
	const className = clsx(classes.root, { [classes.padding]: !disablePadding }, classNameProp);
	const Component = component || 'ul';

	return (
		<Component className={className} {...rest}>
			{children}
		</Component>
	);
}

export default withStyles(styles)(Tree);
