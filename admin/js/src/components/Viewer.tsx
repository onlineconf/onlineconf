import * as React from 'react';
import { Dialog, DialogTitle, DialogContent, DialogActions, Button, Typography, Theme, createStyles, withStyles, WithStyles, withMobileDialog } from '@material-ui/core';

import { typeNames, IParamNode } from './common';
import ValueView from './ValueView';

const styles = (theme: Theme) => createStyles({
	title: {
		overflowWrap: 'break-word',
	},
	value: {
		borderTop: `1px solid ${theme.palette.divider}`,
		borderBottom: `1px solid ${theme.palette.divider}`,
		paddingTop: theme.spacing.unit,
		paddingBottom: theme.spacing.unit,
	},
	versionBox: {
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
			{(param.mime !== 'application/x-null' || param.data !== null) && <Typography variant="caption">{typeNames[param.mime]}</Typography>}
			<Typography variant="body1" component="div" className={classes.value}><ValueView type={param.mime} value={param.data} /></Typography>
			<div className={classes.versionBox}>
				<Typography variant="body1" className={classes.version}>Version: {param.version}</Typography>
				<Typography variant="body1" color="textSecondary">{param.mtime}</Typography>
			</div>
		</DialogContent>
		<DialogActions>
			<Button>Describe</Button>
			<Button onClick={props.onEdit}>Edit</Button>
			<Button color="primary" onClick={props.onClose}>Close</Button>
		</DialogActions>
	</Dialog>
);

export default withStyles(styles)(withMobileDialog<ViewerProps>()(Viewer));
