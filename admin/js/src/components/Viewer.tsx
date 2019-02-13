import * as React from 'react';
import { Dialog, DialogTitle, DialogContent, DialogActions, Button, Typography, Theme, createStyles, withStyles, WithStyles, withMobileDialog } from '@material-ui/core';

import { IParamNode } from './common';
import ValueView from './ValueView';

const styles = (theme: Theme) => createStyles({
	title: {
		overflowWrap: 'break-word',
	},
	value: {
		overflow: 'auto',
	},
	versionBox: {
		marginTop: theme.spacing.unit / 2,
		borderTop: `1px solid ${theme.palette.divider}`,
		display: 'flex',
	},
	version: {
		flex: 'auto',
		marginRight: 2 * theme.spacing.unit,
	},
});

interface ViewerProps {
	param: IParamNode;
	fullScreen?: boolean;
	onDescribe: () => void;
	onEdit: () => void;
	onClose: () => void;
}

const Viewer = ({ param, classes, ...props }: ViewerProps & WithStyles<typeof styles>) => (
	<Dialog open onClose={props.onClose} fullScreen={props.fullScreen}>
		<DialogTitle className={classes.title}>
			{param.path}
			{param.summary !== '' && <Typography variant="body1" color="textSecondary">{param.summary}</Typography>}
		</DialogTitle>
		<DialogContent>
			{param.description !== '' && <Typography variant="body1" paragraph>{param.description}</Typography>}
			<Typography variant="body1" component="div" className={classes.value}><ValueView type={param.mime} value={param.data} accessible={param.rw !== null}/></Typography>
			<div className={classes.versionBox}>
				<Typography variant="body1" className={classes.version}>v.{param.version}</Typography>
				<Typography variant="body1" color="textSecondary">{param.mtime}</Typography>
			</div>
		</DialogContent>
		<DialogActions>
			<Button onClick={props.onDescribe}>Describe</Button>
			<Button onClick={props.onEdit}>Edit</Button>
			<Button color="primary" onClick={props.onClose}>Close</Button>
		</DialogActions>
	</Dialog>
);

export default withStyles(styles)(withMobileDialog<ViewerProps>()(Viewer));
