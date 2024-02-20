import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { Theme } from '@mui/material/styles';
import { WithStyles } from '@mui/styles';
import createStyles from '@mui/styles/createStyles';
import withStyles from '@mui/styles/withStyles';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import Typography from '@mui/material/Typography';
import { useTheme, useMediaQuery } from '@mui/material';

import { ValueProps } from './common';
import { Notification, postParam, IParam, ParamModify } from '../api';
import NotificationControl from './NotificationControl';
import TypeValueFields from './TypeValueFields';
import SummaryDescriptionFields from './SummaryDescriptionFields';

const styles = (theme: Theme) => createStyles({
	paper: {
		minWidth: theme.breakpoints.values.sm,
	},
	content: {
		display: 'flex',
		flexDirection: 'column',
	},
	pathPrefix: {
		marginRight: 0,
		color: theme.palette.text.secondary,
		display: 'block',
		height: 'auto',
		overflow: 'hidden',
		textOverflow: 'ellipsis',
		whiteSpace: 'nowrap',
		direction: 'rtl',
	},
	pathInput: {
		width: 'auto',
		flex: 'auto',
	},
});

export interface EditorProps extends Partial<ValueProps> {
	path: string;
	version?: number;
	create?: boolean;
	summary?: string;
	fullScreen?: boolean;
	description?: string;
	notification?: Notification;
	onClose: () => void;
	onChange: (param: IParam) => void;
	onError: (error: unknown) => void;
}

interface EditorState extends ValueProps {
	open: boolean;
	name: string;
	summary: string;
	description: string;
	notification?: Notification | null;
	comment: string;
}

class Editor extends React.Component<EditorProps & WithStyles<typeof styles> & WithTranslation, EditorState> {

	constructor(props: EditorProps & WithStyles<typeof styles> & WithTranslation) {
		super(props);
		this.state = {
			open: true,
			name: '',
			type: props.type !== undefined ? props.type : 'application/x-null',
			value: props.value !== undefined ? props.value : null,
			summary: props.summary !== undefined ? props.summary : '',
			description: props.description !== undefined ? props.description : '',
			notification: props.notification !== undefined ? null : undefined,
			comment: '',
		};
	}

	handleNameChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		this.setState({ name: event.target.value });
	};

	handleSummaryChange = (summary: string) => {
		this.setState({ summary });
	};

	handleDescriptionChange = (description: string) => {
		this.setState({ description });
	};

	handleTypeValueChange = (newState: Pick<EditorState, 'type' | 'value'>) => {
		this.setState(newState);
	};

	handleNotificationChange = (notification: Notification | null) => {
		this.setState({ notification });
	};

	handleCommentChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		this.setState({ comment: event.target.value });
	};

	handleClose = () => {
		this.setState({ open: false });
	};

	handleSave = async () => {
		const { onChange, onError } = this.props;
		try {
			let path = this.props.path;
			let params: ParamModify;
			if (this.props.create) {
				if (path === '/') {
					path += this.state.name;
				} else {
					path += '/' + this.state.name;
				}
				params = {
					summary: this.state.summary,
					description: this.state.description,
					notification: this.state.notification,
				};
			} else {
				params = {
					version: this.props.version,
				};
			}
			const param = await postParam(path, {
				...params,
				mime: this.state.type,
				data: this.state.value,
				comment: this.state.comment,
			});
			onChange(param);
			this.handleClose();
		} catch (error) {
			onError(error);
		}
	};

	render() {
		const { t } = this.props;
		return (
			<Dialog
				open={this.state.open}
				onClose={this.handleClose}
				maxWidth={false}
				fullScreen={this.props.fullScreen}
				classes={this.props.fullScreen ? undefined : { paper: this.props.classes.paper }}
				TransitionProps={{
					onExited: this.props.onClose
				}}>
				{this.props.create ? <DialogTitle>{t('param.menu.create')}</DialogTitle> : (
					<DialogTitle>
						{this.props.path}
						{this.props.summary !== '' && <Typography variant="body2" color="textSecondary">{this.props.summary}</Typography>}
					</DialogTitle>
				)}
				<DialogContent className={this.props.classes.content}>
					{this.props.create ? (
						<React.Fragment>
							<TextField
								label={t('param.path')}
								autoFocus
								variant="outlined"
								fullWidth
								margin="dense"
								required
								value={this.state.name}
								onChange={this.handleNameChange}
								InputProps={{
									startAdornment: <InputAdornment position="start" disableTypography className={this.props.classes.pathPrefix}>
										{this.props.path.endsWith('/') ? this.props.path : this.props.path + '/'}
									</InputAdornment>
								}}
								inputProps={{ className: this.props.classes.pathInput }}
							/>
							<SummaryDescriptionFields
								summary={this.state.summary}
								description={this.state.description}
								onSummaryChange={this.handleSummaryChange}
								onDescriptionChange={this.handleDescriptionChange}
							/>
						</React.Fragment>
					) : this.props.description !== '' &&
						<Typography paragraph variant="body2">{this.props.description}</Typography>
					}
					<TypeValueFields type={this.state.type} value={this.state.value} onChange={this.handleTypeValueChange} onError={this.props.onError} />
					{this.state.notification !== undefined && (
						<NotificationControl
							compact
							overridden={this.state.notification !== null}
							value={this.state.notification !== null ? this.state.notification : this.props.notification!}
							onChange={this.handleNotificationChange}
						/>
					)}
					<TextField label={t('param.comment')} variant="outlined" margin="dense" fullWidth required value={this.state.comment} onChange={this.handleCommentChange} />
				</DialogContent>
				<DialogActions>
					<Button color="primary" onClick={this.handleClose}>{t('button.cancel')}</Button>
					<Button color="primary" onClick={this.handleSave} disabled={this.state.comment === ''}>{t('button.save')}</Button>
				</DialogActions>
			</Dialog>
		);
	}

}

/* eslint-disable react/display-name */
const withMobileDialog = () => <P extends EditorProps >(WrappedComponent: React.ComponentType<P>) => (props: any) => {
	const theme = useTheme();
	const fullScreen = useMediaQuery(theme.breakpoints.down('sm'));
	return <WrappedComponent {...props} width="lg" fullScreen={fullScreen} />;
};

export default withStyles(styles)(withTranslation()(withMobileDialog()(Editor)));
