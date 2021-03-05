import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { makeStyles, Theme, useTheme } from '@material-ui/core/styles';
import Dialog from '@material-ui/core/Dialog';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogActions from '@material-ui/core/DialogActions';
import TextField from '@material-ui/core/TextField';
import Button from '@material-ui/core/Button';
import IconButton from '@material-ui/core/IconButton';
import useMediaQuery from '@material-ui/core/useMediaQuery';

import EditIcon from '@material-ui/icons/Edit';

import { ParamDialogProps, ValueProps } from './common';
import ParamDialogTitle from './ParamDialogTitle';
import { ValueView } from './value';
import * as API from '../api';
import axios from 'axios';
import Editor, { EditorProps } from './Editor';
import ErrorContext from './ErrorContext';
import NoAccess from './NoAccess';


const useStyles = makeStyles((theme: Theme) => ({
	current: {
		opacity: theme.palette.action.disabledOpacity,
	},
	textBlock: {
		display: 'flex',
		alignItems: 'center',
	},
	text: {
		flex: 'auto',
	},
	edit: {
		flex: 'none',
		margin: '-12px 0',
	},
}));

interface RollbackDialogProps extends ParamDialogProps, ValueProps {
	version: number;
	onChange: (param: API.IParam) => void;
}

export default function RollbackDialog(props: RollbackDialogProps) {
	const classes = useStyles();
	const { t } = useTranslation();

	const theme = useTheme();
	const fullScreen = useMediaQuery(theme.breakpoints.down('sm'));

	const genericOnError = React.useContext(ErrorContext);
	const [ current, setCurrent ] = React.useState<API.IParam>();
	const [ comment, setComment ] = React.useState('');
	const [ editor, setEditor ] = React.useState<JSX.Element>();

	React.useEffect(() => {
		const cts = axios.CancelToken.source();
		API.getParam(props.path, undefined, undefined, { cancelToken: cts.token })
			.then(data => {
				setCurrent(data);
				props.onLoaded();
			})
			.catch(error => {
				props.onError(error);
				props.onClose();
			});
		return () => cts.cancel('Operation canceled by the user.');
	}, [props]);
	if (current === undefined) {
		return null;
	}

	const changeDone = (param: API.IParam) => {
		props.onChange(param);
		props.onClose();
	};

	const handleEdit = (edit: Required<Pick<EditorProps, 'path' | 'type' | 'value'>>) => setEditor(
		<Editor
			path={edit.path}
			version={current.version}
			type={edit.type}
			value={edit.value}
			summary={current.summary}
			description={current.description}
			onChange={changeDone}
			onError={genericOnError}
			onClose={() => setEditor(undefined)}
		/>
	);
	const currentEditProps = { path: current.path, type: current.mime, value: current.data };

	const handleRollback = () => {
		API.postParam(props.path, {
			version: current.version,
			mime: props.type,
			data: props.value,
			comment,
		})
			.then(changeDone)
			.catch(props.onError);
	};

	return (
		<Dialog open onClose={props.onClose} fullScreen={fullScreen}>
			<ParamDialogTitle path={props.path}>{t('log.rollback.rollback')}</ParamDialogTitle>
			<DialogContent>
				<DialogContentText component="div" className={classes.current}>
					<div className={classes.textBlock}>
						<div className={classes.text}>{t('log.rollback.current', { version: current.version})}</div>
						<div className={classes.edit}><IconButton onClick={() => handleEdit(currentEditProps)}><EditIcon/></IconButton></div>
					</div>
					{current.rw === null ? <NoAccess/> : (
						<ValueView type={current.mime} value={current.data}/>
					)}
				</DialogContentText>
				<DialogContentText component="div">
					<div className={classes.textBlock}>
						<div className={classes.text}>{t('log.rollback.confirmation', { param: props.path, version: props.version })}</div>
						<div className={classes.edit}><IconButton onClick={() => handleEdit(props)}><EditIcon/></IconButton></div>
					</div>
					<ValueView type={props.type} value={props.value}/>
				</DialogContentText>
				<TextField
					label={t('param.comment')}
					variant="outlined"
					margin="dense"
					fullWidth
					required
					autoFocus
					value={comment}
					onChange={event => setComment(event.target.value)}
				/>
			</DialogContent>
			<DialogActions>
				<Button color="primary" onClick={props.onClose}>{t('button.cancel')}</Button>
				<Button color="primary" onClick={handleRollback} disabled={comment === ''}>{t('log.rollback.rollback')}</Button>
			</DialogActions>
			{editor}
		</Dialog>
	);
}
