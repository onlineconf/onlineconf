import * as React from 'react';
import { Theme, createStyles, WithStyles, withStyles } from '@material-ui/core/styles';

import { NonNullValueProps, EditNonnullValueProps } from './common';
import TextValuePreview from './text/TextValuePreview';
import ServerListValueEditBase, { ServerListEntry } from './ServerListValueEditBase';

const viewStyles = (theme: Theme) => createStyles({
	server: {
		display: 'grid',
		gridTemplateColumns: '[host] auto [port] 1fr',
	},
	host: {
		gridColumn: 'host',
	},
	port: {
		gridColumn: 'port',
		'&::before': {
			content: '":"',
		}
	},
});

const Server2ListValueView = (props: NonNullValueProps & WithStyles<typeof viewStyles>) => {
	return (
		<div>
			{props.value.split(';').map((v, i) => {
				const [host, ports] = v.split(':');

				return (
					<div key={i} className={props.classes.server}>
						<div className={props.classes.host}>{host}</div>
						{ports.split(',').map((p, j) => <div key={j} className={props.classes.port}>{p}</div>)}
					</div>
				);
			})}
		</div>
	);
};

const split = (value: string) => {
	return value.split(';').map(item => {
		const m = item.match(/^(\[[^\]]*\]?|.*?)(?::([^:]*))?$/);
		const host = m ? m[1] : item;
		const ports = m && m[2] !== undefined ? m[2].split(',') : [];
		return { host, ports };
	});
};

const join = (list: ServerListEntry[]) => (
	list.map(item => (
		item.host + (item.ports.length > 0 ? ':' + item.ports.join(',') : '')
	)).join(';')
);

const Server2ListValueEdit = (props: EditNonnullValueProps) => (
	<ServerListValueEditBase split={split} join={join} {...props}/>
);

const server2Value = {
	preview: TextValuePreview,
	view: withStyles(viewStyles)(Server2ListValueView),
	edit: Server2ListValueEdit,
};
export default server2Value;
