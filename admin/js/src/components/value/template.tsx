import * as React from 'react';

import { NonNullValueProps } from '../common';
import TextValueView from './text/TextValueView';
import TextValueEdit from './text/TextValueEdit';

const TemplateValuePreview = (props: NonNullValueProps) => {
	const tokens = props.value.split(/(\$\{.*?\})/g);
	return (
		<span>
			{tokens.map((token: string, i: number) => {
				if (i % 2 === 0) {
					return <span key={i}>{token}</span>;
				} else {
					const m = tokens[i].match(/^\$\{(\/.*)\}$/);
					return m ? <a key={i} href={'#' + m[1]} className="template" onClick={event => event.stopPropagation()}>{m[0]}</a>
						: <span key={i} className="template">{token}</span>;
				}
			})}
		</span>
	);
};

export default {
	preview: TemplateValuePreview,
	view: TextValueView,
	edit: TextValueEdit,
};
