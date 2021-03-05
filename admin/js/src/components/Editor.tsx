import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { createStyles, withStyles, WithStyles, Theme } from '@material-ui/core/styles';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import InputAdornment from '@material-ui/core/InputAdornment';
import Typography from '@material-ui/core/Typography';
import withMobileDialog from '@material-ui/core/withMobileDialog';

import { ValueProps } from './common';
import { Notification, postParam, IParam, ParamModify } from '../api';
import NotificationControl from './NotificationControl';
import TypeValueFields from './TypeValueFields';
import SummaryDescriptionFields from './SummaryDescriptionFields';

const styles = (theme: Theme) => createStyles({
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
	description?: string;
	notification?: Notification;
	fullScreen?: boolean;
	onClose: () => void;
	onChange: (param: IParam) => void;
	onError: (error: Error) => void;
}

interface EditorState extends ValueProps {
	name?: string;
	summary: string;
	description: string;
	notification?: Notification | null;
	comment: string;
}

class Editor extends React.Component<EditorProps & WithStyles<typeof styles> & WithTranslation, EditorState> {

	constructor(props: EditorProps & WithStyles<typeof styles> & WithTranslation) {
		super(props);
		this.state = {
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
	}

	handleSummaryChange = (summary: string) => {
		this.setState({ summary });
	}

	handleDescriptionChange = (description: string) => {
		this.setState({ description });
	}

	handleTypeValueChange = (newState: Pick<EditorState, 'type' | 'value'>) => {
		this.setState(newState);
	}

	handleNotificationChange = (notification: Notification | null) => {
		this.setState({ notification });
	}

	handleCommentChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		this.setState({ comment: event.target.value });
	}

	handleClose = () => {
		this.props.onClose();
	}

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
	}

	render() {
		const { t } = this.props;
		return (
			<Dialog open onClose={this.handleClose} fullWidth fullScreen={this.props.fullScreen}>
				{this.props.create ? <DialogTitle>{t('param.menu.create')}</DialogTitle> : (
					<DialogTitle>
						{this.props.path}
						{this.props.summary !== '' && <Typography variant="body2" color="textSecondary">{this.props.summary}</Typography>}
					</DialogTitle>
				)}
				<DialogContent>
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
					<Button color="primary" onClick={this.handleSave}>{t('button.save')}</Button>
				</DialogActions>
			</Dialog>
		);
	}

}

export default withStyles(styles)(withMobileDialog<EditorProps>()(withTranslation()(Editor)));
