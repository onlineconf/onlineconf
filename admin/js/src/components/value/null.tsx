import * as React from 'react';
import { WithStyles, createStyles, Theme, withStyles } from '@material-ui/core';

import { ValueProps, EditValueProps } from '../common';

const NullValuePreview = (props: ValueProps) => null;

const viewStyles = (theme: Theme) => createStyles({
	root: {
		color: theme.onlineconf.palette.null,
	},
});

const NullValueView = (props: ValueProps & WithStyles<typeof viewStyles>) => (
	<span className={props.classes.root}>NULL</span>
);

const NullValueEdit = (props: EditValueProps) => null;

export default {
	preview: NullValuePreview,
	view: withStyles(viewStyles)(NullValueView),
	edit: NullValueEdit,
};
