import * as React from 'react';
import { Theme, createStyles, WithStyles, withStyles } from '@material-ui/core/styles';

import { NonNullValueProps } from './common';
import TextValueView from './text/TextValueView';
import TextValueEdit from './text/TextValueEdit';

const previewStyles = (theme: Theme) => createStyles({
	variable: {
		textDecoration: 'none',
		color: '#0055aa',
	},
});

const TemplateValuePreview = (props: NonNullValueProps & WithStyles<typeof previewStyles>) => {
	const tokens = props.value.split(/(\$\{.*?\})/g);
	return (
		<span>
			{tokens.map((token: string, i: number) => {
				if (i % 2 === 0) {
					return <span key={i}>{token}</span>;
				} else {
					const m = tokens[i].match(/^\$\{(\/.*)\}$/);
					return m ? <a key={i} href={'#' + m[1]} className={props.classes.variable} onClick={event => event.stopPropagation()}>{m[0]}</a>
						: <span key={i} className={props.classes.variable}>{token}</span>;
				}
			})}
		</span>
	);
};

const templateValue = {
	preview: withStyles(previewStyles)(TemplateValuePreview),
	view: TextValueView,
	edit: TextValueEdit,
};
export default templateValue;
