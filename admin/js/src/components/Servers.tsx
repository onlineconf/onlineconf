import * as React from 'react';
import { Table, TableHead, TableBody, TableRow, TableCell, TableSortLabel, withStyles, WithStyles, IconButton, createStyles, DialogTitle, DialogContent, Dialog, DialogContentText, DialogActions, Button, Theme } from '@material-ui/core';

import * as API from '../api';
import { smartCompare } from './common';

import RemoveIcon from '@material-ui/icons/RemoveCircle';

const styles = (theme: Theme) => createStyles({
	alert: {
		color: theme.palette.error.dark,
	},
	head: {
		height: 40,
	},
	row: {
		height: 32,
	},
	delete: {
		padding: 2,
	},
});

type ServerColumns = 'host' | 'mtime' | 'online' | 'package';

interface ServersProps {
	onError: (error: Error) => void;
}

interface ServersState {
	servers: API.Server[];
	orderBy: ServerColumns;
	order: 'asc' | 'desc';
	confirmDeleteServer?: API.Server;
}

class Servers extends React.Component<ServersProps & WithStyles<typeof styles>, ServersState> {

	state: ServersState = {
		servers: [],
		orderBy: 'host',
		order: 'asc',
	};

	componentDidMount() {
		this.load();
	}

	private async load() {
		try {
			const servers = await API.getServers();
			this.setState(({ orderBy, order }) => this.sort({ servers, orderBy, order }));
		} catch (error) {
			this.props.onError(error);
		}
	}

	private createSortHandler(column: ServerColumns) {
		return (event: React.MouseEvent<{}>) => {
			this.setState(({ servers, orderBy, order }) => {
				if (orderBy === column) {
					order = order === 'asc' ? 'desc' : 'asc';
				} else {
					orderBy = column;
				}

				return this.sort({ servers, orderBy, order });
			});
		};
	}

	private sort({ servers, orderBy, order }: ServersState) {
		let sort = (a: API.Server, b: API.Server) => smartCompare(a.host, b.host);
		if (orderBy !== 'host') {
			const sortByHost = sort;
			sort = (a, b) => (a[orderBy] === b[orderBy] ? sortByHost(a, b) : a[orderBy] < b[orderBy] ? -1 : 1);
		}
		if (order === 'desc') {
			const sortAsc = sort;
			sort = (a, b) => sortAsc(b, a);
		}
		servers.sort(sort);
		return { servers, orderBy, order };
	}

	private deleteServer(server: API.Server) {
		if (new Date(server.online) < new Date(Date.now() - 2592000000)) {
			this.deleteServerNow(server.host);
		} else {
			this.setState({ confirmDeleteServer: server });
		}
	}

	private async deleteServerNow(host: string) {
		try {
			await API.deleteServer(host);
			this.setState(({ servers }) => ({
				servers: servers.filter(s => s.host !== host),
				confirmDeleteServer: undefined,
			}));
		} catch (error) {
			this.props.onError(error);
		}
	}

	private renderConfirmDeleteServer() {
		const server = this.state.confirmDeleteServer;
		if (server === undefined) {
			return null;
		}
		const handleCloseDialog = () => {
			this.setState({ confirmDeleteServer: undefined });
		};
		return (
			<Dialog open onClose={handleCloseDialog}>
				<DialogTitle>Delete server?</DialogTitle>
				<DialogContent>
					<DialogContentText>Delete {server.host} from monitoring?</DialogContentText>
				</DialogContent>
				<DialogActions>
					<Button color="primary" onClick={handleCloseDialog}>Cancel</Button>
					<Button color="primary" onClick={() => this.deleteServerNow(server.host)}>OK</Button>
				</DialogActions>
			</Dialog>
		);
	}

	render() {
		const { classes } = this.props;
		const { servers, orderBy, order } = this.state;
		return (
			<React.Fragment>
				<Table padding="dense">
					<TableHead>
						<TableRow className={classes.head}>
							<TableCell><TableSortLabel active={orderBy === 'host'} direction={order} onClick={this.createSortHandler('host')}>Server</TableSortLabel></TableCell>
							<TableCell padding="none"/>
							<TableCell><TableSortLabel active={orderBy === 'mtime'} direction={order} onClick={this.createSortHandler('mtime')}>Modifiation</TableSortLabel></TableCell>
							<TableCell><TableSortLabel active={orderBy === 'online'} direction={order} onClick={this.createSortHandler('online')}>Online</TableSortLabel></TableCell>
							<TableCell><TableSortLabel active={orderBy === 'package'} direction={order} onClick={this.createSortHandler('package')}>Version</TableSortLabel></TableCell>
						</TableRow>
					</TableHead>
					<TableBody>
						{servers.map(server => {
							return (
								<TableRow key={server.host} className={classes.row}>
									<TableCell>{server.host}</TableCell>
									<TableCell padding="none">
										{server.online_alert && <IconButton onClick={() => this.deleteServer(server)} className={classes.delete}><RemoveIcon/></IconButton>}
									</TableCell>
									<TableCell className={server.mtime_alert ? classes.alert : undefined}>{server.mtime}</TableCell>
									<TableCell className={server.online_alert ? classes.alert : undefined}>{server.online}</TableCell>
									<TableCell>{server.package}</TableCell>
								</TableRow>
							);
						})}
					</TableBody>
				</Table>
				{this.renderConfirmDeleteServer()}
			</React.Fragment>
		);
	}

}

export default withStyles(styles)(Servers);
