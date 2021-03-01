import * as React from 'react';
import { createStyles, withStyles, WithStyles } from '@material-ui/core/styles';
import DialogTitle from '@material-ui/core/DialogTitle';
import Typography from '@material-ui/core/Typography';

const styles = createStyles({
	path: {
		overflowWrap: 'break-word',
	},
});

interface ParamDialogTitleProps {
	children: string;
	path: string;
}

const ParamDialogTitle = (props: ParamDialogTitleProps & WithStyles<typeof styles>) => (
	<DialogTitle>
		{props.children}
		<Typography variant="body2" color="textSecondary" className={props.classes.path}>{props.path}</Typography>
	</DialogTitle>
);

export default withStyles(styles)(ParamDialogTitle);
