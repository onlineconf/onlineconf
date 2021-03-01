import * as React from 'react';
import { createStyles, withStyles, WithStyles, Theme } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';

import CommentIcon from '@material-ui/icons/Comment';

import * as API from '../api';
import { ValueView } from './value';
import symlink from './value/symlink';

const styles = (theme: Theme) => createStyles({
	root: {
		marginTop: theme.spacing(1),
		marginBottom: theme.spacing(1),
		borderWidth: 1,
		borderStyle: 'solid',
		borderColor: theme.palette.divider,
		borderRadius: theme.shape.borderRadius,
		backgroundColor: theme.palette.background.paper,
		boxShadow: theme.shadows[1],
		overflow: 'hidden',
	},
	path: {
		padding: '0 4px',
	},
	header: {
		borderBottomWidth: 1,
		borderBottomStyle: 'solid',
		borderBottomColor: theme.palette.divider,
		display: 'flex',
		flexWrap: 'wrap',
		justifyContent: 'space-between',
		alignItems: 'center',
		'& > div': {
			padding: '0 4px',
		},
	},
	version: {
		minWidth: 0,
		textTransform: 'none',
		padding: '0 4px',
		borderRadius: 0,
		boxShadow: 'none',
	},
	author: {
		flex: 'auto',
		textAlign: 'right',
	},
	value: {
		padding: '0 4px',
	},
	comment: {
		display: 'flex',
		borderTopWidth: 1,
		borderTopStyle: 'solid',
		borderTopColor: theme.palette.divider,
		color: theme.palette.text.secondary,
	},
	commentIcon: {
		display: 'block',
		fontSize: 16,
	},
	commentText: {
		flex: 'auto',
		padding: '0 4px 0 2px',
	},
	rollback: {
		marginLeft: 4,
	},
});

interface LogCardProps extends API.IParamLog {
	showPath?: boolean;
}

const LogCard = (props: LogCardProps & WithStyles<typeof styles>) => (
	<Typography component="div" variant="body2" className={props.classes.root}>
		{props.showPath && <div className={props.classes.path}><symlink.view type="application/x-symlink" value={props.path}/></div>}
		<div className={props.classes.header}>
			<Button variant="contained" size="small" disabled={props.rw !== true} className={props.classes.version}>v.{props.version}</Button>
			<div>{props.mtime}</div>
			<div className={props.classes.author}>{props.author}</div>
		</div>
		<div className={props.classes.value}>
			<ValueView type={props.mime} value={props.data} accessible={props.rw !== null}/>
		</div>
		<div className={props.classes.comment}>
			<CommentIcon className={props.classes.commentIcon}/>
			<div className={props.classes.commentText}>
				{props.comment}
			</div>
		</div>
	</Typography>
);

export default withStyles(styles)(LogCard);
