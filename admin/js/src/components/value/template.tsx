import * as React from 'react';
import { Theme, makeStyles } from '@material-ui/core/styles';

import { NonNullValueProps } from './common';
import TextValueView from './text/TextValueView';
import TextValueEdit from './text/TextValueEdit';

const usePreviewStyles = makeStyles((theme: Theme) => ({
	root: {
		color: theme.palette.text.primary,
	},
	variable: {
		textDecoration: 'none',
		color: '#0055aa',
	},
}));

function TemplateValuePreview(props: NonNullValueProps) {
	const classes = usePreviewStyles();
	const tokens = props.value.split(/(\$\{.*?\})/g);
	return (
		<span className={classes.root}>
			{tokens.map((token: string, i: number) => {
				if (i % 2 === 0) {
					return <span key={i}>{token}</span>;
				} else {
					const m = tokens[i].match(/^\$\{(\/.*)\}$/);
					return m ? <a key={i} href={'#' + m[1]} className={classes.variable} onClick={event => event.stopPropagation()}>{m[0]}</a>
						: <span key={i} className={classes.variable}>{token}</span>;
				}
			})}
		</span>
	);
}

const templateValue = {
	preview: TemplateValuePreview,
	view: TextValueView,
	edit: TextValueEdit,
};
export default templateValue;
