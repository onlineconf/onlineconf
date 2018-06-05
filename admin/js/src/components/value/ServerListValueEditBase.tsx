import * as React from 'react';
import { createStyles, WithStyles, withStyles, IconButton, TextField, Theme, InputAdornment } from '@material-ui/core';

import AddIcon from '@material-ui/icons/AddCircle';
import RemoveIcon from '@material-ui/icons/RemoveCircle';

import { EditNonnullValueProps } from '../common';

const styles = (theme: Theme) => createStyles({
	list: {
		margin: `${theme.spacing.unit}px 0`,
		display: 'grid',
		gridTemplateColumns: '[action] 36px [host] 1fr [port] 110px [port-action] 36px',
		justifyItems: 'center',
	},
	action: {
		gridColumn: 'action',
		padding: 6,
	},
	host: {
		gridColumn: 'host',
	},
	port: {
		gridColumn: 'port',
	},
	portAction: {
		gridColumn: 'port-action',
		padding: 6,
	},
});

export interface ServerListEntry {
	host: string;
	ports: string[];
}

interface ServerListValueEditBaseProps extends EditNonnullValueProps {
	split: (value: string) => ServerListEntry[];
	join: (list: ServerListEntry[]) => string;
}

export default withStyles(styles)(
	class ServerListValueEditBase extends React.Component<ServerListValueEditBaseProps & WithStyles<typeof styles>> {

		private handleAddHost = () => {
			this.props.onChange({ target: { value: this.props.value + ',' } });
				let list = this.getList();
				list.push({ host: '', ports: [] });
				this.setList(list);
		}

		private createRemoveHostHandler(id: number) {
			return () => {
				let list = this.getList();
				list.splice(id, 1);
				this.setList(list);
			};
		}

		private createChangeHostHandler(id: number) {
			return (event: React.ChangeEvent<HTMLInputElement>) => {
				let list = this.getList();
				list[id].host = event.target.value.replace(/\s+/g, '');
				this.setList(list);
			};
		}

		private createAddPortHandler(hostId: number) {
			return () => {
				let list = this.getList();
				list[hostId].ports.push('');
				this.setList(list);
			};
		}

		private createRemovePortHandler(hostId: number, portId: number) {
			return () => {
				let list = this.getList();
				list[hostId].ports.splice(portId, 1);
				this.setList(list);
			};
		}

		private createChangePortHandler(hostId: number, portId: number) {
			return (event: React.ChangeEvent<HTMLInputElement>) => {
				let list = this.getList();
				list[hostId].ports[portId] = event.target.value.replace(/[^\d;,]+/g, '');
				this.setList(list);
			};
		}

		private getList(): ServerListEntry[] {
			return this.props.split(this.props.value);
		}

		private setList(list: ServerListEntry[]) {
			this.props.onChange({ target: { value: this.props.join(list) } });
		}

		render() {
			const list = this.getList();
			return (
				<div className={this.props.classes.list}>
					{list.map((item, i) => (
						<React.Fragment>
							<IconButton onClick={this.createRemoveHostHandler(i)} className={this.props.classes.action}><RemoveIcon/></IconButton>
							<TextField fullWidth margin="none" value={item.host} onChange={this.createChangeHostHandler(i)} autoFocus={item.host === ''} className={this.props.classes.host}/>
							{item.ports.map((port, j) => (
								<React.Fragment>
								<TextField fullWidth margin="none" value={port} onChange={this.createChangePortHandler(i, j)} autoFocus={port === '' && item.host !== ''}
									className={this.props.classes.port}
									InputProps={{ startAdornment: <InputAdornment position="start">:</InputAdornment> }}
								/>
								<IconButton className={this.props.classes.portAction} onClick={this.createRemovePortHandler(i, j)}><RemoveIcon/></IconButton>
								</React.Fragment>
							))}
							<IconButton className={this.props.classes.portAction} onClick={this.createAddPortHandler(i)}><AddIcon /></IconButton>
						</React.Fragment>
					))}
					<IconButton className={this.props.classes.action} onClick={this.handleAddHost}><AddIcon /></IconButton>
				</div>
			);
		}

	}
);
