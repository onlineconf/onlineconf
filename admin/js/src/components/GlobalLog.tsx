import * as React from 'react';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import { Theme, makeStyles } from '@material-ui/core/styles';
import TextField from '@material-ui/core/TextField';
import Checkbox from '@material-ui/core/Checkbox';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import Button from '@material-ui/core/Button';
import LinearProgress from '@material-ui/core/LinearProgress';

import * as API from '../api';
import UserField from './UserField';
import PathField from './PathField';
import LogList from './LogList';

const useFilterStyles = makeStyles((theme: Theme) => ({
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
	const classes = useFilterStyles();
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
					variant="outlined"
					margin="dense"
					className={classes.field}
				/>
				<PathField
					label={t('log.branch')}
					value={branch}
					onChange={value => setBranch(value)}
					variant="outlined"
					margin="dense"
					className={classes.field}
				/>
			</div>
			<div className={classes.subgroup}>
				<TextField
					label={t('log.from')}
					value={from}
					onChange={event => setFrom(event.target.value)}
					variant="outlined"
					margin="dense"
					className={classes.field}
				/>
				<TextField
					label={t('log.till')}
					value={till}
					onChange={event => setTill(event.target.value)}
					variant="outlined"
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
					className={classes.all}
				/>
				<Button
					variant="contained"
					color="primary"
					onClick={() => props.onSubmit({ author, branch, from, till, all })}
					className={classes.load}
				>
					{t('log.load')}
				</Button>
			</div>
		</div>
	);
}

const useStyles = makeStyles((theme: Theme) => ({
	progressBlock: {
		height: 4,
	}
}));

interface GlobalLogProps {
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
	const [ lastID, setLastID ] = React.useState<number>();
	const [ active, setActive ] = React.useState(0);
	const [ data, setData ] = React.useState<API.IParamLog[]>([]);
	const [ more, setMore ] = React.useState(false);

	const { onError } = props;
	React.useEffect(() => {
		setActive(a => a + 1);
		const cts = axios.CancelToken.source();
		API.getGlobalLog(filter, lastID, { cancelToken: cts.token })
			.then(data => {
				if (lastID !== undefined) {
					setData(currentData => [ ...currentData, ...data ]);
				} else {
					setData(data);
				}
				setMore(data.length === API.logLimit);
				setActive(a => a - 1);
			})
			.catch(error => {
				setActive(a => a - 1);
				if (!axios.isCancel(error)) {
					onError(error);
				}
			});
		return () => cts.cancel('Operation canceled by the user.');
	}, [filter, lastID, onError]);
	const loading = active !== 0;

	const ref = React.useRef<HTMLDivElement>(null);
	React.useEffect(() => {
		const handleScroll = () => {
			if (!more || data.length === 0 || ref.current === null) return;
			const offsetBottom = ref.current.offsetHeight + ref.current.offsetTop;
			const windowBottom = window.pageYOffset + window.innerHeight;
			if (offsetBottom - windowBottom < 100) {
				setLastID(data[data.length - 1].id);
			}
		};
		window.addEventListener('scroll', handleScroll);
		return () => window.removeEventListener('scroll', handleScroll);
	});

	const load = (filter: API.GlobalLogFilter) => {
		setLastID(undefined);
		setFilter(filter);
		setData([]);
	};

	const handleChange = React.useCallback(() => load({ ...filter }), [filter]);

	const classes = useStyles();
	return (
		<div ref={ref}>
			<GlobalLogFilter onSubmit={load} loading={loading}/>
			<LogList data={data} showPath onChange={handleChange}/>
			{(more || loading) && (
				<div className={classes.progressBlock}>
					{loading && <LinearProgress/>}
				</div>
			)}
		</div>
	);
}
