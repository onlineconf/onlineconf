import * as React from 'react';
import { Theme, createStyles, WithStyles, withStyles } from '@material-ui/core/styles';

import { NonNullValueProps } from '../common';
import ListValueEdit from './ListValueEdit';
import TextValuePreview from '../text/TextValuePreview';

const viewStyles = (theme: Theme) => createStyles({
	root: {
		listStyleType: 'none',
		margin: 0,
		padding: 0,
	},
});

const ListValueView = (props: NonNullValueProps & WithStyles<typeof viewStyles>) => {
	const list = props.value.split(',');
	return (
		<ul className={props.classes.root}>
			{list.map((v, i) => <li key={i}>{v}</li>)}
		</ul>
	);
};

const listValue = {
	preview: TextValuePreview,
	view: withStyles(viewStyles)(ListValueView),
	edit: ListValueEdit,
};
export default listValue;
