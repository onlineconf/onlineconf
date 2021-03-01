import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { WithStyles, createStyles, Theme, withStyles } from '@material-ui/core/styles';

const styles = (theme: Theme) => createStyles({
	root: {
		color: theme.onlineconf.palette.noAccess,
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
