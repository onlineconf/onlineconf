import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { Theme } from '@mui/material/styles';
import makeStyles from '@mui/styles/makeStyles';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import Collapse from '@mui/material/Collapse';
import Fade from '@mui/material/Fade';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';

import CommentIcon from '@mui/icons-material/Comment';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import UndoIcon from '@mui/icons-material/Undo';

import * as API from '../api';
import { ValuePreview, ValueView } from './value';
import symlink from './value/symlink';
import Avatar from './Avatar';
import RollbackDialog from './RollbackDialog';
import ButtonProgress from './ButtonProgress';
import ErrorContext from './ErrorContext';
import NoAccess from './NoAccess';
import ParamIcon from './ParamIcon';

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
	versionRow: {
		display: 'flex',
	},
	versionBlock: {
		flex: 'none',
		maxWidth: '100%',
	},
	flexBlock: {
		display: 'flex',
		alignItems: 'center',
		columnGap: theme.spacing(0.5),
	},
	flexBlockText: {
		flex: 'auto',
		overflow: 'hidden',
		textOverflow: 'ellipsis',
		whiteSpace: 'nowrap',
		color: theme.palette.text.primary,
	},
	version: {
		backgroundColor: theme.palette.text.secondary,
		color: theme.palette.background.default,
		borderRadius: 6,
		minWidth: 20,
		lineHeight: 1,
		padding: 2,
		textAlign: 'center',
	},
	headerComment: {
		marginLeft: theme.spacing(0.5),
		color: theme.palette.text.secondary,
		overflow: 'hidden',
		textOverflow: 'ellipsis',
		whiteSpace: 'nowrap',
	},
	deleted: {
		color: theme.palette.text.secondary,
		letterSpacing: '1em',
		fontVariant: 'small-caps',
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
	onChange: (param: API.IParam | null) => void;
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
			deleted={props.deleted}
			onLoaded={() => setRollbackLoading(false)}
			onChange={props.onChange}
			onError={onError}
			onClose={() => {
				setRollback(undefined);
				setRollbackLoading(false);
			}}
		/>);
		setRollbackLoading(true);
	}, [onError, props.data, props.deleted, props.mime, props.onChange, props.path, props.version]);

	return (
		<ListItem disableGutters className={classes.root}>
			<CardHeader
				title={
					<React.Fragment>
						{props.showPath && <div className={classes.path}><symlink.view type="application/x-symlink" value={props.path}/></div>}
						<div className={classes.versionRow}>
							<div className={`${classes.flexBlock} ${classes.versionBlock}`}>
								<div className={classes.version}>{props.version}</div>
								<div className={classes.flexBlockText}>{props.mtime}</div>
							</div>
							<Fade in={!expanded}>
								<div className={classes.headerComment}>
									<CommentIcon className={classes.commentIcon}/>
									{props.comment}
								</div>
							</Fade>
						</div>
					</React.Fragment>
				}
				subheader={
					<Collapse in={!expanded}>
						{props.deleted ? <div className={classes.deleted}>{t('log.deleted')}</div> : (
							<div className={classes.flexBlock}>
								<ParamIcon type={props.mime} fontSize="small" color="action"/>
								<div className={classes.flexBlockText}>
									{props.rw === null ? <NoAccess/> : <ValuePreview type={props.mime} value={props.data}/>}
								</div>
							</div>
						)}
					</Collapse>
				}
				avatar={<Avatar username={props.author}/>}
				action={
					<IconButton onClick={handleExpand} size="large">
						{expanded ? <ExpandLessIcon/> : <ExpandMoreIcon/>}
					</IconButton>
				}
				disableTypography
				classes={headerClasses}
			/>
			<Collapse in={expanded} mountOnEnter unmountOnExit>
				<CardContent className={classes.content}>
					{
						props.deleted
							? <span className={classes.deleted}>{t('log.deleted')}</span>
							: props.rw === null
								? <NoAccess/>
								: <ValueView type={props.mime} value={props.data}/>
					}
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
	onChange: (param: API.IParam | null) => void;
}

export default function LogList(props: LogListProps) {
	const classes = useStyles();
	return (
		<List className={classes.root} disablePadding>
			{props.data.map(row => <LogItem key={`${row.path} ${row.version}`} {...row} showPath={props.showPath} onChange={props.onChange}/>)}
		</List>
	);
}
