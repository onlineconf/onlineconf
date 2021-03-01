import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';

import { NullValueProps } from './common';

export function NullValuePreview(props: NullValueProps) {
	return null;
}

const useViewStyles = makeStyles((theme: Theme) => ({
	root: {
		color: theme.onlineconf.palette.null,
	},
}));

export function NullValueView(props: NullValueProps) {
	const classes = useViewStyles();
	return <span className={classes.root}>NULL</span>;
}

export function NullValueEdit(props: NullValueProps) {
	return null;
}

const nullValue = {
	preview: NullValuePreview,
	view: NullValueView,
	edit: NullValueEdit,
};
export default nullValue;
