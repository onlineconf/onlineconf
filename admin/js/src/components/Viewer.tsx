import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { Theme } from '@mui/material/styles';
import { WithStyles } from '@mui/styles';
import createStyles from '@mui/styles/createStyles';
import withStyles from '@mui/styles/withStyles';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import { useTheme, useMediaQuery } from '@mui/material';

import { IParamNode } from './common';
import { ValueView } from './value';
import { RevealButton, useReveal } from './value/encrypted';
import NoAccess from './NoAccess';
import ErrorContext from './ErrorContext';

const styles = (theme: Theme) => createStyles({
	title: {
		overflowWrap: 'break-word',
	},
	valueRow: {
		display: 'flex',
		alignItems: 'flex-start',
		columnGap: theme.spacing(1),
	},
	value: {
		flex: 'auto',
		minWidth: 0,
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

const Viewer = ({ param, classes, t, ...props }: ViewerProps & WithStyles<typeof styles> & WithTranslation) => {
	const onError = React.useContext(ErrorContext);
	const { revealed, loading, reveal, hide } = useReveal(param.path, param.version, onError);
	return (
		<Dialog open onClose={props.onClose} maxWidth={false} fullScreen={props.fullScreen}>
			<DialogTitle className={classes.title}>
				{param.path}
				{param.summary !== '' && <Typography variant="body2" color="textSecondary">{param.summary}</Typography>}
			</DialogTitle>
			<DialogContent>
				{param.description !== '' && <Typography paragraph variant="body2">{param.description}</Typography>}
				{param.rw === null ? <NoAccess/> : (
					<div className={classes.valueRow}>
						<ValueView type={param.mime} value={revealed ? revealed.data : param.data} encrypted={param.encrypted} className={classes.value}/>
						{param.can_reveal && <RevealButton revealed={revealed !== undefined} loading={loading} onReveal={reveal} onHide={hide} size="medium"/>}
					</div>
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
};

/* eslint-disable react/display-name */
const withMobileDialog = () => <P extends ViewerProps >(WrappedComponent: React.ComponentType<P>) => (props: any) => {
	const theme = useTheme();
	const fullScreen = useMediaQuery(theme.breakpoints.down('sm'));
	return <WrappedComponent {...props} width="lg" fullScreen={fullScreen} />;
};

export default withStyles(styles)(withMobileDialog()(withTranslation()(Viewer)));
