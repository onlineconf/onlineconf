import * as React from 'react';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import { Theme, useTheme } from '@mui/material/styles';
import makeStyles from '@mui/styles/makeStyles';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import LinearProgress from '@mui/material/LinearProgress';
import Button from '@mui/material/Button';
import useMediaQuery from '@mui/material/useMediaQuery';

import * as API from '../api';
import { ParamDialogProps } from './common';
import ParamDialogTitle from './ParamDialogTitle';
import LogList from './LogList';
import ErrorContext from './ErrorContext';

const useStyles = makeStyles((theme: Theme) => ({
	progressBlock: {
		height: 4,
	}
}));

interface ParamLogProps extends ParamDialogProps {
	onChange: (param: API.IParam | null) => void;
}

export default function ParamLog(props: ParamLogProps) {
	const { t } = useTranslation();

	const theme = useTheme();
	const fullScreen = useMediaQuery(theme.breakpoints.down('md'));

	const [ open, setOpen ] = React.useState(false);
	const [ lastID, setLastID ] = React.useState<number>();
	const [ active, setActive ] = React.useState(0);
	const [ data, setData ] = React.useState<API.IParamLog[]>([]);
	const [ more, setMore ] = React.useState(false);

	const { path } = props;
	const nextPageHandlers = {
		onLoaded: React.useCallback(() => { return }, []),
		onError: React.useContext(ErrorContext),
	};
	const { onLoaded, onError } = lastID === undefined ? props : nextPageHandlers;
	const load = () => {
		setActive(a => a + 1);
		const cts = axios.CancelToken.source();
		API.getParamLog(path, lastID, { cancelToken: cts.token })
			.then(data => {
				if (lastID !== undefined) {
					setData(currentData => [ ...currentData, ...data ]);
				} else {
					setData(data);
				}
				setMore(data.length === API.logLimit);
				setOpen(true);
				setActive(a => a - 1);
				onLoaded();
			})
			.catch(error => {
				setActive(a => a - 1);
				if (axios.isCancel(error)) {
					onLoaded();
				} else {
					onError(error);
				}
			});
		return () => cts.cancel('Operation canceled by the user.');
	};
	React.useEffect(load, [path, lastID, onError, onLoaded]);
	const loading = active !== 0;

	const handleClose = () => setOpen(false);

	const handleChange = (param: API.IParam | null) => {
		load();
		props.onChange(param);
	};

	const handleScroll: React.UIEventHandler<HTMLDivElement> = React.useCallback(event => {
		if (!more || data.length === 0) return;
		const content = event.currentTarget;
		if (content.scrollHeight - content.scrollTop - content.clientHeight < 100) {
			setLastID(data[data.length - 1].id);
		}
	}, [data, more]);

	const classes = useStyles();
	return (
		<Dialog
			open={open}
			onClose={handleClose}
			maxWidth="md"
			fullScreen={fullScreen}
			TransitionProps={{
				onExited: props.onClose
			}}>
			<ParamDialogTitle path={path}>{t('param.menu.history')}</ParamDialogTitle>
			<DialogContent onScroll={handleScroll}>
				<LogList data={data} onChange={handleChange}/>
				{(more || loading) && (
					<div className={classes.progressBlock}>
						{loading && <LinearProgress/>}
					</div>
				)}
			</DialogContent>
			<DialogActions>
				<Button color="primary" onClick={handleClose}>{t('button.close')}</Button>
			</DialogActions>
		</Dialog>
	);
}
