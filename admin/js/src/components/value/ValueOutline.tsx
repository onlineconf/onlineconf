import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { Theme } from '@mui/material/styles';

import { WithStyles } from '@mui/styles';
import withStyles from '@mui/styles/withStyles';
import createStyles from '@mui/styles/createStyles';

const styles = (theme: Theme) => createStyles({
	root: {
		borderStyle: 'solid',
		borderRadius: theme.shape.borderRadius,
		borderWidth: 1,
		borderColor: theme.palette.mode === 'light' ? 'rgba(0, 0, 0, 0.23)' : 'rgba(255, 255, 255, 0.23)',
		margin: '3px 0 4px 0',
		padding: 0,
		minInlineSize: 'initial',
	},
	legend: {
		cursor: 'default',
		textAlign: 'left',
		margin: '0 8px',
		padding: '0 4px',
		lineHeight: '11px',
		fontFamily: theme.typography.fontFamily,
		color: theme.palette.text.secondary,
		fontSize: theme.typography.pxToRem(16 * 0.75),
	},
});

function ValueOutline(props: { children?: React.ReactNode } & WithStyles<typeof styles>) {
	const { t } = useTranslation();
	return (
		<fieldset className={props.classes.root}>
			<legend className={props.classes.legend}>{t('param.value')}</legend>
			{props.children}
		</fieldset>
	);
}

export default withStyles(styles)(ValueOutline);
