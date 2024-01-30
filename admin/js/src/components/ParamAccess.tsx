import * as React from 'react';
import axios, { CancelTokenSource } from 'axios';
import { withTranslation, WithTranslation } from 'react-i18next';
import { Theme } from '@mui/material/styles';
import { WithStyles } from '@mui/styles';
import createStyles from '@mui/styles/createStyles';
import withStyles from '@mui/styles/withStyles';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import ListItemSecondaryAction from '@mui/material/ListItemSecondaryAction';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import { useTheme, useMediaQuery } from '@mui/material';

import ParamDialogTitle from './ParamDialogTitle';

import { ParamDialogProps } from './common';
import * as API from '../api';

const styles = (theme: Theme) => createStyles({
	list: {
		display: 'flex',
		flexWrap: 'wrap',
	},
	item: {
		flex: '400px',
		margin: theme.spacing(0, 1),
	},
	checkbox: {
		width: 32,
		height: 32,
		margin: theme.spacing(0, 0.5),
	},
	button: {
		padding: theme.spacing(0.5),
		minWidth: 32,
		minHeight: 32,
		margin: theme.spacing(0, 0.5),
		boxShadow: theme.shadows[0],
	},
});

interface ParamAccessGroup extends API.ParamAccessGroup {
	indeterminate?: true;
}

interface ParamAccessState {
	open: boolean;
	access: ParamAccessGroup[];
}

class ParamAccess extends React.Component<ParamDialogProps & WithStyles<typeof styles> & WithTranslation, ParamAccessState> {

	state: ParamAccessState = {
		open: false,
		access: [],
	};

	private cts?: CancelTokenSource;

	componentDidMount() {
		this.load();
	}

	componentDidUpdate(prevProps: ParamDialogProps) {
		if (this.props.path !== prevProps.path) {
			this.cancel();
			this.load();
		}
	}

	componentWillUnmount() {
		this.cancel();
	}

	private async load() {
		const { onLoaded, onError } = this.props;
		try {
			this.setState({ open: false });
			this.cts = axios.CancelToken.source();
			const access = await API.getParamAccess(this.props.path, { cancelToken: this.cts.token });
			this.setState({ access, open: true });
			onLoaded();
		} catch (error) {
			if (axios.isCancel(error)) {
				onLoaded();
			} else {
				onError(error);
			}
		}
	}

	private cancel() {
		if (this.cts) {
			this.cts.cancel('Operation canceled by the user.');
			this.cts = undefined;
		}
	}

	private overrideHandler(group: string) {
		return (event: React.ChangeEvent<HTMLInputElement>) => {
			if (event.target.checked) {
				this.setState(({ access }) => {
					for (const g of access) {
						if (g.group === group) {
							g.overridden = true;
							g.indeterminate = true;
						}
					}
					return { access };
				});
			} else {
				for (const g of this.state.access) {
					if (g.group === group && g.indeterminate) {
						g.overridden = false;
						delete(g.indeterminate);
						this.setState({ access: this.state.access });
						return;
					}
				}
				(async () => {
					const data = await API.deleteParamAccess(this.props.path, group);
					this.setState(({ access }) => {
						for (const group of access) {
							if (group.group === data.group) {
								group.overridden = data.overridden;
								group.rw = data.rw;
								delete(group.indeterminate);
							}
						}
						return { access };
					});
				})();
			}
		};
	}

	private rwHandler(group: string, rw: boolean | null) {
		return async () => {
			const data = await API.postParamAccess(this.props.path, group, rw);
			this.setState(({ access }) => {
				for (const group of access) {
					if (group.group === data.group) {
						group.overridden = data.overridden;
						group.rw = data.rw;
						delete(group.indeterminate);
					}
				}
				return { access };
			});
		};
	}

	render() {
		const size: { maxWidth?: false, fullWidth?: true } = this.state.access.length > 20 ? { maxWidth: false, fullWidth: true } : {};
		return (
			<Dialog open={this.state.open} onClose={this.props.onClose} fullScreen={this.props.fullScreen} {...size}>
				<ParamDialogTitle path={this.props.path}>{this.props.t('param.menu.access')}</ParamDialogTitle>
				<DialogContent>
					<List className={this.props.classes.list}>
						{this.state.access.map(group => (
							<ListItem key={group.group} divider ContainerProps={{ className: this.props.classes.item }}>
								<ListItemText>{group.group}</ListItemText>
								<ListItemSecondaryAction>
									<Checkbox
										checked={group.overridden}
										indeterminate={group.indeterminate}
										onChange={this.overrideHandler(group.group)}
										className={this.props.classes.checkbox}
									/>
									<Button
										disabled={!group.overridden}
										variant={group.rw === null ? 'contained' : 'text'}
										className={this.props.classes.button}
										onClick={this.rwHandler(group.group, null)}
									>
										NO
									</Button>
									<Button
										disabled={!group.overridden}
										variant={group.rw === false ? 'contained' : 'text'}
										className={this.props.classes.button}
										onClick={this.rwHandler(group.group, false)}
									>
										RO
									</Button>
									<Button
										disabled={!group.overridden}
										variant={group.rw === true ? 'contained' : 'text'}
										className={this.props.classes.button}
										onClick={this.rwHandler(group.group, true)}
									>
										RW
									</Button>
								</ListItemSecondaryAction>
							</ListItem>
						))}
					</List>
				</DialogContent>
				<DialogActions>
					<Button color="primary" onClick={this.props.onClose}>{this.props.t('button.close')}</Button>
				</DialogActions>
			</Dialog>
		);
	}

}

/* eslint-disable react/display-name */
const withMobileDialog = () => <P extends ParamDialogProps >(WrappedComponent: React.ComponentType<P>) => (props: any) => {
	const theme = useTheme();
	const fullScreen = useMediaQuery(theme.breakpoints.down('sm'));
	return <WrappedComponent {...props} width="lg" fullScreen={fullScreen} />;
};

export default withStyles(styles)(withMobileDialog()(withTranslation()(ParamAccess)));
