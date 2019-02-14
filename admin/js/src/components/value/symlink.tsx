import * as React from 'react';
import { EditNonnullValueProps, NonNullValueProps } from '../common';
import PathField from '../PathField';
import { Theme, createStyles, WithStyles, withStyles } from '@material-ui/core';

const viewStyles = (theme: Theme) => createStyles({
	root: {
		color: theme.onlineconf.palette.symlink,
		textDecoration: 'none',
	},
});

const SymlinkValueView = (props: NonNullValueProps & WithStyles<typeof viewStyles>) => (
	<a className={props.classes.root} href={'/#' + props.value} onClick={event => event.stopPropagation()}>{props.value}</a>
);

const SymlinkValueEdit = (props: EditNonnullValueProps) => (
	<PathField {...props} label="Value" symlink="resolve" fullWidth variant="outlined" margin="dense"/>
);

const SymlinkValueViewStyled = withStyles(viewStyles)(SymlinkValueView);

export default {
	preview: SymlinkValueViewStyled,
	view: SymlinkValueViewStyled,
	edit: SymlinkValueEdit,
};
