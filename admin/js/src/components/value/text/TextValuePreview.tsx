import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { NonNullValueProps } from '../common';

const useStyles = makeStyles((theme: Theme) => ({
	root: {
		color: theme.palette.text.primary,
	},
}), { name: 'TextValuePreview' });

export default function TextValuePreview(props: NonNullValueProps) {
	const classes = useStyles();
	return <span className={classes.root}>{props.value}</span>;
}
