import * as React from 'react';
import { Theme, createStyles, WithStyles, withStyles } from '@material-ui/core/styles';

import { NonNullValueProps, EditNonnullValueProps } from './common';
import TextValuePreview from './text/TextValuePreview';
import ServerListValueEditBase, { ServerListEntry } from './ServerListValueEditBase';

const viewStyles = (theme: Theme) => createStyles({
	root: {
		listStyleType: 'none',
		margin: 0,
		padding: 0,
	},
	sameHost: {
		color: 'transparent',
	},
});

const ServerListValueView = (props: NonNullValueProps & WithStyles<typeof viewStyles>) => {
	const list = props.value.split(',');
	let lastHost: string;
	return (
		<ul className={props.classes.root}>
			{list.map((v, i) => {
				const [host, port] = v.split(':');
				if (host === lastHost) {
					return <li key={i}><span className={props.classes.sameHost}>{host}</span>:{port}</li>;
				} else {
					lastHost = host;
					return <li key={i}>{v}</li>;
				}
			})}
		</ul>
	);
};

const split = (value: string) => {
	const list: ServerListEntry[] = [];
	for (const item of value.split(',')) {
		const m = item.match(/^(\[[^\]]*\]?|.*?)(?::([^:]*))?$/);
		const host = m ? m[1] : item;
		const port = m && m[2] !== undefined ? m[2] : undefined;
		if (port === undefined) {
			list.push({ host, ports: [] });
		} else if (list.length > 0 && list[list.length - 1].host === host && list[list.length - 1].ports.length > 0) {
			list[list.length - 1].ports.push(port);
		} else {
			list.push({ host, ports: [port] });
		}
	}
	return list;
};

const join = (list: ServerListEntry[]) => {
	const items: string[] = [];
	for (const item of list) {
		if (item.ports.length > 0) {
			for (const port of item.ports) {
				items.push(item.host + ':' + port);
			}
		} else {
			items.push(item.host);
		}
	}
	return items.join(',');
};

const ServerListValueEdit = (props: EditNonnullValueProps) => (
	<ServerListValueEditBase split={split} join={join} {...props}/>
);

const serverValue = {
	preview: TextValuePreview,
	view: withStyles(viewStyles)(ServerListValueView),
	edit: ServerListValueEdit,
};
export default serverValue;
