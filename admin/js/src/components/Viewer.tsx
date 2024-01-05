import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { createStyles, withStyles, WithStyles } from '@mui/styles';
import { Theme } from '@mui/material/styles';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';

import { IParamNode } from './common';
import { ValueView } from './value';
import NoAccess from './NoAccess';

const styles = (theme: Theme) => createStyles({
	title: {
		overflowWrap: 'break-word',
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
	<Dialog open onClose={props.onClose} maxWidth={false} fullScreen={props.fullScreen}>
		<DialogTitle className={classes.title}>
			{param.path}
			{param.summary !== '' && <Typography variant="body2" color="textSecondary">{param.summary}</Typography>}
		</DialogTitle>
		<DialogContent>
			{param.description !== '' && <Typography paragraph variant="body2">{param.description}</Typography>}
			{param.rw === null ? <NoAccess/> : (
				<ValueView type={param.mime} value={param.data}/>
			)}
			<div className={classes.versionBox}>
				<Typography variant="body2" className={classes.version}>v.{param.version}</Typography>
				<Typography variant="body2" color="textSecondary">{param.mtime}</Typography>
			</div>
		</DialogContent>
		<DialogActions>
			<Button onClick={props.onDescribe}>{t('param.menu.describe')}</Button>
			<Button onClick={props.onEdit} disabled={param.rw !== true}>{t('param.menu.edit')}</Button>
			<Button color="primary" onClick={props.onClose}>{t('button.close')}</Button>
		</DialogActions>
	</Dialog>
);

export default withStyles(styles)(withTranslation()(Viewer));
