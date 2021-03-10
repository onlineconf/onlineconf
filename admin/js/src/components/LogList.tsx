import * as React from 'react';
import clsx from 'clsx';
import { useTranslation } from 'react-i18next';
import { makeStyles, Theme } from '@material-ui/core/styles';
import CardHeader from '@material-ui/core/CardHeader';
import CardContent from '@material-ui/core/CardContent';
import CardActions from '@material-ui/core/CardActions';
import Collapse from '@material-ui/core/Collapse';
import Fade from '@material-ui/core/Fade';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import Button from '@material-ui/core/Button';
import IconButton from '@material-ui/core/IconButton';
import Typography from '@material-ui/core/Typography';

import CommentIcon from '@material-ui/icons/Comment';
import ExpandLessIcon from '@material-ui/icons/ExpandLess';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import UndoIcon from '@material-ui/icons/Undo';

import * as API from '../api';
import { ValuePreview, ValueView } from './value';
import symlink from './value/symlink';
import Avatar from './Avatar';
import RollbackDialog from './RollbackDialog';
import ButtonProgress from './ButtonProgress';
import ErrorContext from './ErrorContext';
import NoAccess from './NoAccess';

const useItemStyles = makeStyles((theme: Theme) => ({
	root: {
		display: 'block',
		padding: 0,
		borderTop: `1px solid ${theme.palette.divider}`,
	},
	headerContent: {
		minWidth: 0,
	},
	path: {
		wordBreak: 'break-all',
	},
	versionLine: {
		overflow: 'hidden',
		textOverflow: 'ellipsis',
		whiteSpace: 'nowrap',
	},
	headerComment: {
		marginLeft: theme.spacing(1),
		color: theme.palette.text.secondary,
	},
	version: {
		backgroundColor: theme.palette.text.secondary,
		color: theme.palette.background.default,
		padding: '0 2px',
		borderRadius: 4,
	},
	preview: {
		overflow: 'hidden',
		textOverflow: 'ellipsis',
		whiteSpace: 'nowrap',
		color: theme.palette.text.primary,
	},
	deleted: {
		textDecoration: '3px solid line-through rgba(255, 0, 0, 0.5)',
	},
	content: {
		paddingTop: 0,
		paddingBottom: 0,
		'&:last-child': {
			paddingBottom: theme.spacing(2),
		},
	},
	commentBlock: {
		marginTop: theme.spacing(1),
	},
	commentIcon: {
		fontSize: 'inherit',
		marginRight: theme.spacing(0.5),
	},
}), { name: 'LogItem' });

interface LogItemProps extends API.IParamLog {
	showPath?: boolean;
	onChange: (param: API.IParam) => void;
}

function LogItem(props: LogItemProps) {
	const classes = useItemStyles();
	const { t } = useTranslation();
	const onError = React.useContext(ErrorContext);
	const [ expanded, setExpanded ] = React.useState(false);
	const [ rollback, setRollback ] = React.useState<JSX.Element>();
	const [ rollbackLoading, setRollbackLoading ] = React.useState(false);
	const handleExpand = React.useCallback(() => setExpanded(expanded => !expanded), []);
	const headerClasses = React.useMemo(() => ({ content: classes.headerContent }), [classes.headerContent]);

	const showRollbackDialog = React.useCallback(() => {
		setRollback(<RollbackDialog
			path={props.path}
			type={props.mime}
			value={props.data}
			version={props.version}
			onLoaded={() => setRollbackLoading(false)}
			onChange={props.onChange}
			onError={onError}
			onClose={() => {
				setRollback(undefined);
				setRollbackLoading(false);
			}}
		/>);
		setRollbackLoading(true);
	}, [onError, props.data, props.mime, props.onChange, props.path, props.version]);

	return (
		<ListItem disableGutters className={classes.root}>
			<CardHeader
				title={
					<React.Fragment>
						{props.showPath && <span className={classes.path}><symlink.view type="application/x-symlink" value={props.path}/></span>}
						<div className={classes.versionLine}>
							<span className={classes.version}>{props.version}</span>
							{' ' + props.mtime + ' '}
							<Fade in={!expanded}>
								<span className={classes.headerComment}>
									<CommentIcon className={classes.commentIcon}/>
									{props.comment}
								</span>
							</Fade>
						</div>
					</React.Fragment>
				}
				subheader={
					<Collapse in={!expanded}>
						{props.rw === null ? <NoAccess/> : (
							<div className={clsx(classes.preview, props.deleted && classes.deleted)}>
								<ValuePreview type={props.mime} value={props.data}/>
							</div>
						)}
					</Collapse>
				}
				avatar={<Avatar username={props.author}/>}
				action={
					<IconButton onClick={handleExpand}>
						{expanded ? <ExpandLessIcon/> : <ExpandMoreIcon/>}
					</IconButton>
				}
				classes={headerClasses}
			/>
			<Collapse in={expanded} mountOnEnter unmountOnExit>
				<CardContent className={classes.content}>
					{props.rw === null ? <NoAccess/> : (
						<ValueView type={props.mime} value={props.data} className={clsx(props.deleted && classes.deleted)}/>
					)}
					<Typography variant="body2" color="textSecondary" component="div" className={classes.commentBlock}>
						<CommentIcon className={classes.commentIcon}/>
						{props.comment}
					</Typography>
				</CardContent>
				{!props.same && (
					<CardActions>
						<ButtonProgress loading={rollbackLoading}>
							<Button color="primary" disabled={props.rw !== true} endIcon={<UndoIcon/>} onClick={showRollbackDialog}>{t('log.rollback.rollback')}</Button>
						</ButtonProgress>
					</CardActions>
				)}
				{rollback}
			</Collapse>
		</ListItem>
	);
}

const useStyles = makeStyles((theme: Theme) => ({
	root: {
		borderBottom: `1px solid ${theme.palette.divider}`,
	},
}), { name: 'LogList' });

interface LogListProps {
	data: API.IParamLog[];
	showPath?: boolean;
	onChange: (param: API.IParam) => void;
}

export default function LogList(props: LogListProps) {
	const classes = useStyles();
	return (
		<List className={classes.root} disablePadding>
			{props.data.map(row => <LogItem key={`${row.path} ${row.version}`} {...row} showPath={props.showPath} onChange={props.onChange}/>)}
		</List>
	);
}
