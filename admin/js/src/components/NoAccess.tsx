import * as React from 'react';
import { WithStyles, createStyles, Theme, withStyles } from '@material-ui/core';

const styles = (theme: Theme) => createStyles({
	root: {
		color: theme.onlineconf.palette.noAccess,
		letterSpacing: '1em',
		fontVariant: 'small-caps',
	},
});

const NoAccess = (props: WithStyles<typeof styles>) => (
	<span className={props.classes.root}>нет доступа</span>
);

export default withStyles(styles)(NoAccess);
