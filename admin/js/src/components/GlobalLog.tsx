import * as React from 'react';
import axios, { CancelTokenSource } from 'axios';
import { withTranslation, WithTranslation } from 'react-i18next';
import { withStyles, createStyles, WithStyles, Theme } from '@material-ui/core/styles';
import TextField from '@material-ui/core/TextField';
import Checkbox from '@material-ui/core/Checkbox';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import Button from '@material-ui/core/Button';

import * as API from '../api';
import UserField from './UserField';
import PathField from './PathField';
import ButtonProgress from './ButtonProgress';
import LogCard from './LogCard';

const styles = (theme: Theme) => createStyles({
	filter: {
		display: 'flex',
		flexWrap: 'wrap',
		alignItems: 'center',
		borderBottom: `1px solid ${theme.palette.divider}`,
		padding: theme.spacing(1),
	},
	field: {
		flex: 'auto',
		marginLeft: theme.spacing(1),
		marginRight: theme.spacing(1),
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
		margin: theme.spacing(1),
	},
	body: {
		padding: theme.spacing(1, 2),
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

class GlobalLog extends React.Component<GlobalLogProps & WithStyles<typeof styles> & WithTranslation, GlobalLogState> {

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
		const { t } = this.props;
		return (
			<div>
				<div className={this.props.classes.filter}>
					<div className={this.props.classes.subgroup}>
						<UserField
							label={t('log.author')}
							value={this.state.author}
							onChange={value => this.setState({ author: value })}
							variant="filled"
							margin="dense"
							className={this.props.classes.field}
						/>
						<PathField
							label={t('log.branch')}
							value={this.state.branch}
							onChange={value => this.setState({ branch: value })}
							variant="filled"
							margin="dense"
							className={this.props.classes.field}
						/>
					</div>
					<div className={this.props.classes.subgroup}>
						<TextField
							label={t('log.from')}
							value={this.state.from}
							onChange={event => this.setState({ from: event.target.value })}
							variant="filled"
							margin="dense"
							className={this.props.classes.field}
						/>
						<TextField
							label={t('log.till')}
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
							label={t('log.all')}
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

export default withTranslation()(withStyles(styles)(GlobalLog));
