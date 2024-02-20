import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { Theme } from '@mui/material/styles';
import { WithStyles } from '@mui/styles';
import createStyles from '@mui/styles/createStyles';
import withStyles from '@mui/styles/withStyles';
import { EditNonnullValueProps, NonNullValueProps } from './common';
import PathField from '../PathField';

const viewStyles = (theme: Theme) => createStyles({
	root: {
		color: theme.palette.onlineconf.symlink,
		textDecorationLine: 'none',
		textDecorationColor: 'inherit',
	},
});

const SymlinkValueView = (props: NonNullValueProps & WithStyles<typeof viewStyles>) => (
	<a className={props.classes.root} href={'/#' + props.value} onClick={event => event.stopPropagation()}>{props.value}</a>
);

const SymlinkValueEdit = (props: EditNonnullValueProps) => {
	const { t } = useTranslation();
	return (
		<PathField {...props} label={t('param.value')} symlink="resolve" fullWidth variant="outlined" margin="dense"/>
	);
};

const SymlinkValueViewStyled = withStyles(viewStyles)(SymlinkValueView);

const symlinkValue = {
	preview: SymlinkValueViewStyled,
	view: SymlinkValueViewStyled,
	edit: SymlinkValueEdit,
};
export default symlinkValue;
