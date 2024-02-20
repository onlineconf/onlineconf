import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { Theme } from '@mui/material/styles';

import { WithStyles } from '@mui/styles';
import createStyles from '@mui/styles/createStyles';
import withStyles from '@mui/styles/withStyles';

const styles = (theme: Theme) => createStyles({
	root: {
		color: theme.palette.onlineconf.noAccess,
		letterSpacing: '1em',
		fontVariant: 'small-caps',
	},
});

function NoAccess(props: WithStyles<typeof styles>) {
	const { t } = useTranslation();
	return (
		<span className={props.classes.root}>{t('param.noAccess')}</span>
	);
}

export default withStyles(styles)(NoAccess);
