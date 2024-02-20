import * as React from 'react';
import { Theme } from '@mui/material/styles';

import makeStyles from '@mui/styles/makeStyles';

import { NullValueProps } from './common';

export function NullValuePreview(props: NullValueProps) {
	return null;
}

const useViewStyles = makeStyles((theme: Theme) => ({
	root: {
		color: theme.palette.onlineconf.null,
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
