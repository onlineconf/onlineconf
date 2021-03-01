import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { Theme, createStyles, withStyles, WithStyles } from '@material-ui/core/styles';
import Dialog from '@material-ui/core/Dialog';
import DialogTitle from '@material-ui/core/DialogTitle';
import DialogContent from '@material-ui/core/DialogContent';
import DialogActions from '@material-ui/core/DialogActions';
import Button from '@material-ui/core/Button';
import Typography from '@material-ui/core/Typography';
import withMobileDialog from '@material-ui/core/withMobileDialog';

import { IParamNode } from './common';
import { ValueView } from './value';

const styles = (theme: Theme) => createStyles({
	title: {
		overflowWrap: 'break-word',
	},
	value: {
		overflow: 'auto',
	},
	versionBox: {
		marginTop: theme.spacing(0.5),
		borderTop: `1px solid ${theme.palette.divider}`,
		display: 'flex',
	},
	version: {
		flex: 'auto',
		marginRight: theme.spacing(2),
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
			{param.summary !== '' && <Typography variant="body2" color="textSecondary">{param.summary}</Typography>}
		</DialogTitle>
		<DialogContent>
			{param.description !== '' && <Typography paragraph variant="body2">{param.description}</Typography>}
			<Typography component="div" variant="body2" className={classes.value}><ValueView type={param.mime} value={param.data} accessible={param.rw !== null}/></Typography>
			<div className={classes.versionBox}>
				<Typography variant="body2" className={classes.version}>v.{param.version}</Typography>
				<Typography variant="body2" color="textSecondary">{param.mtime}</Typography>
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
