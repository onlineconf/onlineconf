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
	deleted: boolean;
	onChange: (param: API.IParam | null) => void;
}

export default function RollbackDialog(props: RollbackDialogProps) {
	const classes = useStyles();
	const { t } = useTranslation();

	const theme = useTheme();
	const fullScreen = useMediaQuery(theme.breakpoints.down('sm'));

	const genericOnError = React.useContext(ErrorContext);
	const [ current, setCurrent ] = React.useState<API.IParam | null>();
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
				if (error.response?.status === 404) {
					setCurrent(null);
					props.onLoaded();
				} else {
					props.onError(error);
					props.onClose();
				}
			});
		return () => cts.cancel('Operation canceled by the user.');
	}, [props]);
	if (current === undefined) {
		return null;
	}

	const changeDone = (param: API.IParam | null) => {
		props.onChange(param);
		props.onClose();
	};

	const handleEdit = current !== null
		? (edit: Required<Pick<EditorProps, 'path' | 'type' | 'value'>>) => setEditor(
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
		)
		: undefined;

	const handleRollback = () => {
		if (props.deleted) {
			if (current === null) return;
			API.deleteParam(props.path, {
				version: current.version,
				comment,
			})
				.then(() => changeDone(null))
				.catch(props.onError);
		} else {
			const ver = current !== null ? { version: current.version } : {};
			API.postParam(props.path, {
				...ver,
				mime: props.type,
				data: props.value,
				comment,
			})
				.then(changeDone)
				.catch(props.onError);
		}
	};

	return (
		<Dialog open onClose={props.onClose} fullScreen={fullScreen}>
			<ParamDialogTitle path={props.path}>{t('log.rollback.rollback')}</ParamDialogTitle>
			<DialogContent>
				{current !== null && (
					<DialogContentText component="div" className={classes.current}>
						<div className={classes.textBlock}>
							<div className={classes.text}>{t('log.rollback.current', { version: current.version})}</div>
							{handleEdit && (
								<div className={classes.edit}>
									<IconButton onClick={() => handleEdit({ path: current.path, type: current.mime, value: current.data })}>
										<EditIcon/>
									</IconButton>
								</div>
							)}
						</div>
						{current.rw === null ? <NoAccess/> : (
							<ValueView type={current.mime} value={current.data}/>
						)}
					</DialogContentText>
				)}
				<DialogContentText component="div">
					{props.deleted ? t('param.delete.confirm', { param: props.path }) : (
						<React.Fragment>
							<div className={classes.textBlock}>
								<div className={classes.text}>{t('log.rollback.confirmation', { param: props.path, version: props.version })}</div>
								{handleEdit && (
									<div className={classes.edit}><IconButton onClick={() => handleEdit(props)}><EditIcon/></IconButton></div>
								)}
							</div>
							<ValueView type={props.type} value={props.value}/>
						</React.Fragment>
					)}
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
