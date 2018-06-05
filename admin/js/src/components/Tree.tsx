import * as React from 'react';
import * as classNames from 'classnames';
import { StandardProps, createStyles } from '@material-ui/core';
import { withStyles, WithStyles, Theme } from '@material-ui/core/styles';

const styles = (theme: Theme) => createStyles({
	root: {
		listStyle: 'none',
		margin: 0,
		padding: 0,
	},
	padding: {
		paddingTop: theme.spacing.unit,
		paddingBottom: theme.spacing.unit,
	},
});

type TreeClassKey = keyof ReturnType<typeof styles>;

interface TreeProps extends StandardProps<React.HTMLAttributes<HTMLUListElement>, TreeClassKey> {
	component?: React.ReactType<TreeProps>;
	disablePadding?: boolean;
}

function Tree(props: TreeProps & WithStyles<TreeClassKey>) {
	const { children, classes, className: classNameProp, component, disablePadding, ...rest } = props;
	const className = classNames(classes.root, { [classes.padding]: !disablePadding }, classNameProp);
	const Component = component || 'ul';

	return (
		<Component className={className} {...rest}>
			{children}
		</Component>
	);
}

export default withStyles(styles)(Tree);
