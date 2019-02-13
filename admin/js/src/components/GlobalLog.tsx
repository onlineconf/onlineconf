import * as React from 'react';
import axios, { CancelTokenSource } from 'axios';
import { TextField, withStyles, createStyles, WithStyles, Theme, Checkbox, FormControlLabel, Button } from '@material-ui/core';

import * as API from '../api';
import PathField from './PathField';
import ButtonProgress from './ButtonProgress';
import LogCard from './LogCard';

const styles = (theme: Theme) => createStyles({
	filter: {
		display: 'flex',
		flexWrap: 'wrap',
		alignItems: 'center',
		borderBottom: `1px solid ${theme.palette.divider}`,
		padding: `${2 * theme.spacing.unit}px ${theme.spacing.unit}px`,
	},
	field: {
		flex: 'auto',
		marginLeft: theme.spacing.unit,
		marginRight: theme.spacing.unit,
	},
	subgroup: {
		flex: '100 1 auto',
		display: 'flex',
		flexWrap: 'wrap',
		alignItems: 'center',
	},
	loadGroup: {
		flex: '1 0 auto',
		display: 'flex',
		alignItems: 'center',
		justifyContent: 'space-between',
	},
	all: {
		marginLeft: -6,
	},
	load: {
		margin: theme.spacing.unit,
	},
	body: {
		padding: 2 * theme.spacing.unit,
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
	all: boolean;
	loading: boolean;
	data: API.IParamLog[];
}

class GlobalLog extends React.Component<GlobalLogProps & WithStyles<typeof styles>, GlobalLogState> {

	state: GlobalLogState = {
		author: '',
		branch: '',
		from: '',
		till: '',
		all: false,
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
		const filter: API.GlobalLogFilter = {
			author: this.state.author,
			branch: this.state.branch,
			from: this.state.from,
			till: this.state.till,
			all: this.state.all,
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
					<div className={this.props.classes.subgroup}>
						<TextField
							label="Author"
							value={this.state.author}
							onChange={event => this.setState({ author: event.target.value })}
							variant="filled"
							margin="dense"
							className={this.props.classes.field}
						/>
						<PathField
							label="Branch"
							value={this.state.branch}
							onChange={event => this.setState({ branch: event.target.value })}
							variant="filled"
							margin="dense"
							className={this.props.classes.field}
						/>
					</div>
					<div className={this.props.classes.subgroup}>
						<TextField
							label="From"
							value={this.state.from}
							onChange={event => this.setState({ from: event.target.value })}
							variant="filled"
							margin="dense"
							className={this.props.classes.field}
						/>
						<TextField
							label="Till"
							value={this.state.till}
							onChange={event => this.setState({ till: event.target.value })}
							variant="filled"
							margin="dense"
							className={this.props.classes.field}
						/>
					</div>
					<div className={this.props.classes.loadGroup}>
						<FormControlLabel
							value="all"
							label="All"
							control={<Checkbox />}
							checked={this.state.all}
							onChange={(event, checked) => this.setState({ all: checked })}
							className={this.props.classes.all} />
						<ButtonProgress loading={this.state.loading} className={this.props.classes.load}>
							<Button variant="contained" onClick={() => this.load()}>Load</Button>
						</ButtonProgress>
					</div>
				</div>
				<div className={this.props.classes.body}>
					{this.state.data.map(row => <LogCard key={`${row.path} ${row.version}`} {...row} showPath/>)}
				</div>
			</div>
		);
	}

}

export default withStyles(styles)(GlobalLog);
