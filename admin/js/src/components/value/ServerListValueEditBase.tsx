import * as React from 'react';
import { createStyles, WithStyles, withStyles, Theme } from '@material-ui/core/styles';
import IconButton from '@material-ui/core/IconButton';
import TextField from '@material-ui/core/TextField';
import InputAdornment from '@material-ui/core/InputAdornment';

import AddIcon from '@material-ui/icons/AddCircle';
import RemoveIcon from '@material-ui/icons/RemoveCircle';

import { EditNonnullValueProps } from './common';
import ValueOutline from './ValueOutline';

const styles = (theme: Theme) => createStyles({
	list: {
		margin: 2,
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
			this.props.onChange(this.props.value + ',');
			const list = this.getList();
			list.push({ host: '', ports: [] });
			this.setList(list);
		}

		private createRemoveHostHandler(id: number) {
			return () => {
				const list = this.getList();
				list.splice(id, 1);
				this.setList(list);
			};
		}

		private createChangeHostHandler(id: number) {
			return (event: React.ChangeEvent<HTMLInputElement>) => {
				const list = this.getList();
				list[id].host = event.target.value.replace(/\s+/g, '');
				this.setList(list);
			};
		}

		private createAddPortHandler(hostId: number) {
			return () => {
				const list = this.getList();
				list[hostId].ports.push('');
				this.setList(list);
			};
		}

		private createRemovePortHandler(hostId: number, portId: number) {
			return () => {
				const list = this.getList();
				list[hostId].ports.splice(portId, 1);
				this.setList(list);
			};
		}

		private createChangePortHandler(hostId: number, portId: number) {
			return (event: React.ChangeEvent<HTMLInputElement>) => {
				const list = this.getList();
				list[hostId].ports[portId] = event.target.value.replace(/[^\d;,]+/g, '');
				this.setList(list);
			};
		}

		private getList(): ServerListEntry[] {
			return this.props.split(this.props.value);
		}

		private setList(list: ServerListEntry[]) {
			this.props.onChange(this.props.join(list));
		}

		render() {
			const list = this.getList();
			return (
				<ValueOutline>
					<div className={this.props.classes.list}>
						{list.map((item, i) => (
							<React.Fragment key={i}>
								<IconButton onClick={this.createRemoveHostHandler(i)} className={this.props.classes.action}><RemoveIcon/></IconButton>
								<TextField fullWidth margin="none" value={item.host} onChange={this.createChangeHostHandler(i)} autoFocus={item.host === ''} className={this.props.classes.host}/>
								{item.ports.map((port, j) => (
									<React.Fragment key={j}>
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
				</ValueOutline>
			);
		}

	}
);
