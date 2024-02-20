import * as React from 'react';
import { WithStyles } from '@mui/styles';
import createStyles from '@mui/styles/createStyles';
import withStyles from '@mui/styles/withStyles';
import DialogTitle from '@mui/material/DialogTitle';
import Typography from '@mui/material/Typography';

const styles = createStyles({
	path: {
		overflowWrap: 'break-word',
	},
});

interface ParamDialogTitleProps {
	children: React.ReactNode;
	path: string;
}

const ParamDialogTitle = (props: ParamDialogTitleProps & WithStyles<typeof styles>) => (
	<DialogTitle>
		{props.children}
		<Typography variant="body2" color="textSecondary" className={props.classes.path}>{props.path}</Typography>
	</DialogTitle>
);

export default withStyles(styles)(ParamDialogTitle);
