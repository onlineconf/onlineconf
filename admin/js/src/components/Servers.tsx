import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { Theme } from '@mui/material/styles';
import { WithStyles } from '@mui/styles';
import withStyles from '@mui/styles/withStyles';
import createStyles from '@mui/styles/createStyles';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Table from '@mui/material/Table';
import TableHead from '@mui/material/TableHead';
import TableBody from '@mui/material/TableBody';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import TableSortLabel from '@mui/material/TableSortLabel';

import * as API from '../api';
import { smartCompare } from './common';
import WhoAmIContext from './WhoAmIContext';

import RemoveIcon from '@mui/icons-material/RemoveCircle';

const styles = (theme: Theme) => createStyles({
	root: {
		overflow: 'auto',
	},
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
	onError: (error: unknown) => void;
}

interface ServersState {
	servers: API.Server[];
	orderBy: ServerColumns;
	order: 'asc' | 'desc';
	confirmDeleteServer?: API.Server;
}

class Servers extends React.Component<ServersProps & WithStyles<typeof styles> & WithTranslation, ServersState> {

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
		return (event: React.MouseEvent<HTMLElement>) => {
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
			sort = (a, b) => {
				const av = a[orderBy];
				const bv = b[orderBy];
				return av === bv ? sortByHost(a, b) : bv !== null && (av === null || av < bv) ? -1 : 1;
			};
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
		const { t } = this.props;
		return (
			<Dialog open onClose={handleCloseDialog}>
				<DialogTitle>{t('server.delete.title')}</DialogTitle>
				<DialogContent>
					<DialogContentText>{t('server.delete.message', { server: server.host })}</DialogContentText>
				</DialogContent>
				<DialogActions>
					<Button color="primary" onClick={handleCloseDialog}>{t('button.cancel')}</Button>
					<Button color="primary" onClick={() => this.deleteServerNow(server.host)}>{t('button.ok')}</Button>
				</DialogActions>
			</Dialog>
		);
	}

	static contextType = WhoAmIContext;

	render() {
		const { classes, t } = this.props;
		const { servers, orderBy, order } = this.state;
		return (
			<div className={classes.root}>
				<Table size="small">
					<TableHead>
						<TableRow className={classes.head}>
							<TableCell><TableSortLabel active={orderBy === 'host'} direction={order} onClick={this.createSortHandler('host')}>{t('server.host')}</TableSortLabel></TableCell>
							<TableCell padding="none"/>
							<TableCell><TableSortLabel active={orderBy === 'mtime'} direction={order} onClick={this.createSortHandler('mtime')}>{t('server.mtime')}</TableSortLabel></TableCell>
							<TableCell><TableSortLabel active={orderBy === 'online'} direction={order} onClick={this.createSortHandler('online')}>{t('server.online')}</TableSortLabel></TableCell>
							<TableCell><TableSortLabel active={orderBy === 'package'} direction={order} onClick={this.createSortHandler('package')}>{t('server.package')}</TableSortLabel></TableCell>
						</TableRow>
					</TableHead>
					<TableBody>
						{servers.map(server => {
							return (
								<TableRow key={server.host} className={classes.row}>
									<TableCell>{server.host}</TableCell>
									<TableCell padding="none">
										{server.online_alert && this.context.userIsRoot && (
											<IconButton
												onClick={() => this.deleteServer(server)}
												className={classes.delete}
												size="large"><RemoveIcon/></IconButton>
										)}
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
			</div>
		);
	}

}

export default withTranslation()(withStyles(styles)(Servers));
