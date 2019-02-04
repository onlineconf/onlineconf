import * as React from 'react';
import { withStyles, createStyles, WithStyles, Theme } from '@material-ui/core';

const styles = (theme: Theme) => createStyles({
	root: {
		borderStyle: 'solid',
		borderRadius: theme.shape.borderRadius,
		borderWidth: 1,
		borderColor: theme.palette.type === 'light' ? 'rgba(0, 0, 0, 0.23)' : 'rgba(255, 255, 255, 0.23)',
		margin: '3px 0 4px 0',
		padding: 0,
	},
	legend: {
		cursor: 'default',
		textAlign: 'left',
		margin: '0 8px',
		padding: '0 4px',
		lineHeight: '11px',
		fontFamily: theme.typography.fontFamily,
		color: theme.palette.text.secondary,
		fontSize: theme.typography.pxToRem(16 * 0.75),
	},
});

const ValueOutline = (props: { children?: React.ReactNode } & WithStyles<typeof styles>) => (
	<fieldset className={props.classes.root}>
		<legend className={props.classes.legend}>Value</legend>
		{props.children}
	</fieldset>
);

export default withStyles(styles)(ValueOutline);
