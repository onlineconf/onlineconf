import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
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

const Viewer = ({ param, classes, t, ...props }: ViewerProps & WithStyles<typeof styles> & WithTranslation) => (
	<Dialog open onClose={props.onClose} fullScreen={props.fullScreen}>
		<DialogTitle className={classes.title}>
			{param.path}
			{param.summary !== '' && <Typography color="textSecondary">{param.summary}</Typography>}
		</DialogTitle>
		<DialogContent>
			{param.description !== '' && <Typography paragraph>{param.description}</Typography>}
			<Typography component="div" className={classes.value}><ValueView type={param.mime} value={param.data} accessible={param.rw !== null}/></Typography>
			<div className={classes.versionBox}>
				<Typography className={classes.version}>v.{param.version}</Typography>
				<Typography color="textSecondary">{param.mtime}</Typography>
			</div>
		</DialogContent>
		<DialogActions>
			<Button onClick={props.onDescribe}>{t('param.menu.describe')}</Button>
			<Button onClick={props.onEdit}>{t('param.menu.edit')}</Button>
			<Button color="primary" onClick={props.onClose}>{t('button.close')}</Button>
		</DialogActions>
	</Dialog>
);

export default withStyles(styles)(withMobileDialog<ViewerProps>()(withTranslation()(Viewer)));
