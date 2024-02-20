import * as React from 'react';
import { Theme } from '@mui/material/styles';
import { WithStyles } from '@mui/styles';
import createStyles from '@mui/styles/createStyles';
import withStyles from '@mui/styles/withStyles';
import IconButton from '@mui/material/IconButton';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';

import AddIcon from '@mui/icons-material/AddCircle';
import RemoveIcon from '@mui/icons-material/RemoveCircle';

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
		};

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
								<IconButton
									onClick={this.createRemoveHostHandler(i)}
									className={this.props.classes.action}
									size="large"><RemoveIcon/></IconButton>
								<TextField
									variant="standard"
									fullWidth
									margin="none"
									value={item.host}
									onChange={this.createChangeHostHandler(i)}
									autoFocus={item.host === ''}
									className={this.props.classes.host} />
								{item.ports.map((port, j) => (
									<React.Fragment key={j}>
										<TextField
											variant="standard"
											fullWidth
											margin="none"
											value={port}
											onChange={this.createChangePortHandler(i, j)}
											autoFocus={port === '' && item.host !== ''}
											className={this.props.classes.port}
											InputProps={{ startAdornment: <InputAdornment position="start">:</InputAdornment> }} />
										<IconButton
											className={this.props.classes.portAction}
											onClick={this.createRemovePortHandler(i, j)}
											size="large"><RemoveIcon/></IconButton>
									</React.Fragment>
								))}
								<IconButton
									className={this.props.classes.portAction}
									onClick={this.createAddPortHandler(i)}
									size="large"><AddIcon /></IconButton>
							</React.Fragment>
						))}
						<IconButton
							className={this.props.classes.action}
							onClick={this.handleAddHost}
							size="large"><AddIcon /></IconButton>
					</div>
				</ValueOutline>
			);
		}

	}
);
