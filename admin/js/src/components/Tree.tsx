import * as React from 'react';
import clsx from 'clsx';
import { makeStyles, Theme } from '@material-ui/core/styles';

const useStyles = makeStyles((theme: Theme) => ({
	root: {
		listStyle: 'none',
		margin: 0,
		padding: 0,
	},
	padding: {
		paddingTop: theme.spacing(1),
		paddingBottom: theme.spacing(1),
	},
}));

interface TreeProps extends React.HTMLAttributes<HTMLUListElement> {
	component?: React.ElementType;
	disablePadding?: boolean;
}

export default function Tree(props: TreeProps) {
	const { children, className: classNameProp, component, disablePadding, ...rest } = props;
	const classes = useStyles();
	const className = clsx(classes.root, { [classes.padding]: !disablePadding }, classNameProp);
	const Component = component || 'ul';

	return (
		<Component className={className} {...rest}>
			{children}
		</Component>
	);
}
