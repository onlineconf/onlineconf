import * as React from 'react';
import axios, { CancelTokenSource } from 'axios';
import Table from '@material-ui/core/Table';
import TableHead from '@material-ui/core/TableHead';
import TableBody from '@material-ui/core/TableBody';
import TableRow from '@material-ui/core/TableRow';
import TableCell from '@material-ui/core/TableCell';
import { TextField, withStyles, createStyles, WithStyles, Theme, IconButton } from '@material-ui/core';

import RefreshIcon from '@material-ui/icons/Refresh';

import * as API from '../api';
import IconButtonProgress from './IconButtonProgress';
import PathField from './PathField';
import ValueView from './ValueView';

const styles = (theme: Theme) => createStyles({
	filter: {
		display: 'flex',
		flexWrap: 'wrap',
		borderBottom: `1px solid ${theme.palette.divider}`,
		padding: `${2 * theme.spacing.unit}px ${theme.spacing.unit}px`,
	},
	field: {
		flex: 'auto',
		marginLeft: theme.spacing.unit,
		marginRight: theme.spacing.unit,
	},
	load: {
		marginLeft: theme.spacing.unit,
		marginRight: theme.spacing.unit,
	},
	head: {
		height: 40,
	},
	row: {
		height: 32,
	},
});

interface GlobalLogProps {
	onLoaded: () => void;
	onError: (errors: Error) => void;
}

interface GlobalLogState {
	author: string;
	branch: string;
	from: string;
	till: string;
	loading: boolean;
	data: API.IParamLog[];
}

class GlobalLog extends React.Component<GlobalLogProps & WithStyles<typeof styles>, GlobalLogState> {

	state: GlobalLogState = {
		author: '',
		branch: '',
		from: '',
		till: '',
		loading: false,
		data: [],
	};

	private cts?: CancelTokenSource;

	componentDidMount() {
		this.load();
	}

	componentWillUnmount() {
		this.cancel();
	}

	private async load() {
		const { onLoaded, onError } = this.props;
		const filter: API.GlobalLogParams = {
			author: this.state.author !== '' ? this.state.author : undefined,
			branch: this.state.branch !== '' ? this.state.branch : undefined,
			from: this.state.from !== '' ? this.state.from : undefined,
			till: this.state.till !== '' ? this.state.till : undefined,
		};
		try {
			this.setState({ loading: true });
			this.cts = axios.CancelToken.source();
			const data = await API.getGlobalLog(filter, { cancelToken: this.cts.token });
			this.setState({ data, loading: false });
			onLoaded();
		} catch (error) {
			this.setState({ loading: false });
			if (axios.isCancel(error)) {
				onLoaded();
			} else {
				onError(error);
			}
		}
	}

	private cancel() {
		if (this.cts) {
			this.cts.cancel('Operation canceled by the user.');
			this.cts = undefined;
		}
	}

	render() {
		return (
			<div>
				<div className={this.props.classes.filter}>
					<TextField label="Author" value={this.state.author} onChange={event => this.setState({ author: event.target.value })} className={this.props.classes.field} />
					<PathField label="Branch" value={this.state.branch} onChange={event => this.setState({ branch: event.target.value })} className={this.props.classes.field} />
					<TextField label="From" value={this.state.from} onChange={event => this.setState({ from: event.target.value })} className={this.props.classes.field} />
					<TextField label="Till" value={this.state.till} onChange={event => this.setState({ till: event.target.value })} className={this.props.classes.field} />
					<IconButtonProgress loading={this.state.loading} className={this.props.classes.load}>
						<IconButton onClick={() => this.load()}><RefreshIcon/></IconButton>
					</IconButtonProgress>
				</div>
				<Table padding="dense">
					<TableHead>
						<TableRow className={this.props.classes.head}>
							<TableCell>Time</TableCell>
							<TableCell>Param</TableCell>
							<TableCell>Value</TableCell>
							<TableCell>Author</TableCell>
							<TableCell>Comment</TableCell>
						</TableRow>
					</TableHead>
					<TableBody>
						{this.state.data.map(row => {
							return (
								<TableRow key={`${row.path} ${row.version}`} className={this.props.classes.row}>
									<TableCell>{row.mtime}</TableCell>
									<TableCell>{row.path}</TableCell>
									<TableCell><ValueView type={row.mime} value={row.data} accessible={row.rw !== null}/></TableCell>
									<TableCell>{row.author}</TableCell>
									<TableCell>{row.comment}</TableCell>
								</TableRow>
							);
						})}
					</TableBody>
				</Table>
			</div>
		);
	}

}

export default withStyles(styles)(GlobalLog);
