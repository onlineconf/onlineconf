import * as React from 'react';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import { Theme, makeStyles } from '@material-ui/core/styles';
import TextField from '@material-ui/core/TextField';
import Checkbox from '@material-ui/core/Checkbox';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import Button from '@material-ui/core/Button';

import * as API from '../api';
import UserField from './UserField';
import PathField from './PathField';
import ButtonProgress from './ButtonProgress';
import LogList from './LogList';

const useStyles = makeStyles((theme: Theme) => ({
	filter: {
		display: 'flex',
		flexWrap: 'wrap',
		alignItems: 'center',
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
}));

interface GlobalLogFilterProps {
	loading: boolean;
	onSubmit: (filter: API.GlobalLogFilter) => void;
}

function GlobalLogFilter(props: GlobalLogFilterProps) {
	const { t } = useTranslation();
	const classes = useStyles();
	const [ author, setAuthor ] = React.useState('');
	const [ branch, setBranch ] = React.useState('');
	const [ from, setFrom ] = React.useState('');
	const [ till, setTill ] = React.useState('');
	const [ all, setAll ] = React.useState(false);
	return (
		<div className={classes.filter}>
			<div className={classes.subgroup}>
				<UserField
					label={t('log.author')}
					value={author}
					onChange={value => setAuthor(value)}
					variant="filled"
					margin="dense"
					className={classes.field}
				/>
				<PathField
					label={t('log.branch')}
					value={branch}
					onChange={value => setBranch(value)}
					variant="filled"
					margin="dense"
					className={classes.field}
				/>
			</div>
			<div className={classes.subgroup}>
				<TextField
					label={t('log.from')}
					value={from}
					onChange={event => setFrom(event.target.value)}
					variant="filled"
					margin="dense"
					className={classes.field}
				/>
				<TextField
					label={t('log.till')}
					value={till}
					onChange={event => setTill(event.target.value)}
					variant="filled"
					margin="dense"
					className={classes.field}
				/>
			</div>
			<div className={classes.loadGroup}>
				<FormControlLabel
					value="all"
					label={t('log.all')}
					control={<Checkbox />}
					checked={all}
					onChange={(event, checked) => setAll(checked)}
					className={classes.all} />
				<ButtonProgress loading={props.loading} className={classes.load}>
					<Button variant="contained" onClick={() => props.onSubmit({ author, branch, from, till, all })}>Load</Button>
				</ButtonProgress>
			</div>
		</div>
	);
}

interface GlobalLogProps {
	onLoaded: () => void;
	onError: (errors: Error) => void;
}

export default function GlobalLog(props: GlobalLogProps) {
	const [ filter, setFilter ] = React.useState<API.GlobalLogFilter>({
		author: '',
		branch: '',
		from: '',
		till: '',
		all: false,
	});
	const [ loading, setLoading ] = React.useState(false);
	const [ data, setData ] = React.useState<API.IParamLog[]>([]);
	React.useEffect(() => {
		setLoading(true);
		const cts = axios.CancelToken.source();
		API.getGlobalLog(filter, { cancelToken: cts.token })
			.then(data => {
				setData(data);
				setLoading(false);
				props.onLoaded();
			})
			.catch(error => {
				setData([]);
				setLoading(false);
				props.onError(error);
			});
		return () => cts.cancel('Operation canceled by the user.');
	}, [filter, props]);
	return (
		<React.Fragment>
			<GlobalLogFilter onSubmit={setFilter} loading={loading}/>
			<LogList data={data} showPath onChange={() => setFilter({ ...filter })}/>
		</React.Fragment>
	);
}
