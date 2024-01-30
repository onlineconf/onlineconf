import * as React from 'react';
import { Theme } from '@mui/material/styles';
import makeStyles from '@mui/styles/makeStyles';
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
