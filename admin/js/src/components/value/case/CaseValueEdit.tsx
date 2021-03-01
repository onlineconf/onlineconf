import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { createStyles, Theme, WithStyles, withStyles } from '@material-ui/core/styles';
import IconButton from '@material-ui/core/IconButton';
import CircularProgress from '@material-ui/core/CircularProgress';
import MenuItem from '@material-ui/core/MenuItem';
import TextField, { TextFieldProps } from '@material-ui/core/TextField';

import AddIcon from '@material-ui/icons/AddCircle';
import RemoveIcon from '@material-ui/icons/RemoveCircle';

import { getDictionary, dictionaryKeys } from '../../../cache';
import { Case, CaseConditions, caseConditions } from './common';
import { EditNonnullValueProps } from '../common';
import { ParamType } from '../../../api';
import TypeValueFields from '../../TypeValueFields';
import ValueOutline from '../ValueOutline';

const styles = (theme: Theme) => createStyles({
	label: {
		fontFamily: theme.typography.fontFamily,
		fontSize: theme.typography.pxToRem(12),
		color: theme.palette.text.secondary,
		lineHeight: 1 / 0.75,
	},
	case: {
		borderBottomWidth: 1,
		borderBottomStyle: 'solid',
		borderBottomColor: theme.palette.type === 'light' ? 'rgba(0, 0, 0, 0.23)' : 'rgba(255, 255, 255, 0.23)',
		paddingBottom: theme.spacing(0.5),
		marginBottom: theme.spacing(0.5),
	},
	caseKey: {
		display: 'flex',
		flexWrap: 'wrap',
		marginLeft: theme.spacing(0.5),
		marginRight: theme.spacing(0.5),
	},
	caseField: {
		flex: 'auto',
		marginLeft: theme.spacing(0.5),
		marginRight: theme.spacing(0.5),
	},
	remove: {
		margin: theme.spacing(1, 0.5),
	},
	caseValue: {
		paddingLeft: theme.spacing(1),
		paddingRight: theme.spacing(1),
		paddingBottom: theme.spacing(0.5),
	},
	add: {
		textAlign: 'center',
		marginBottom: theme.spacing(0.5),
	},
});

interface CaseValueEditState {
	loading: {
		groups?: boolean;
		datacenters?: boolean;
		services?: boolean;
	};
	groups?: { [K: string]: string };
	datacenters?: { [K: string]: string };
	services?: { [K: string]: string };
}

export default withStyles(styles)(withTranslation()(
	class CaseValueEdit extends React.Component<EditNonnullValueProps & WithStyles<typeof styles> & WithTranslation, CaseValueEditState> {

		state: CaseValueEditState = {
			loading: {},
		};

		timer?: ReturnType<typeof setTimeout>;

		handleChange(cb: (cases: Case[]) => void) {
			const cases: Case[] = JSON.parse(this.props.value);
			cb.call(this, cases);
			this.props.onChange(JSON.stringify(cases));
		}

		handleAddCase = () => {
			this.handleChange(cases => {
				cases.push({ server: '', mime: 'application/x-null', value: null });
			});
		}

		createRemoveCaseHandler(id: number) {
			return (event: React.MouseEvent<HTMLElement>) => {
				this.handleChange(cases => {
					cases.splice(id, 1);
				});
			};
		}

		createCaseTypeHandler(id: number) {
			return (event: React.ChangeEvent<HTMLInputElement>) => {
				this.handleChange(cases => {
					const newType = event.target.value as CaseConditions | 'default';
					for (const t of caseConditions) {
						if (t in cases[id] && t !== newType) {
							delete(cases[id][t]);
						}
					}
					if (newType !== 'default' && !(newType in cases[id])) {
						cases[id][newType] = '';
					}
				});
			};
		}

		createCaseKeyHandler(id: number) {
			return (event: React.ChangeEvent<HTMLInputElement>) => {
				this.handleChange(cases => {
					for (const t of caseConditions) {
						if (t in cases[id]) {
							cases[id][t] = event.target.value;
						}
					}
				});
			};
		}

		createDataHandler(id: number) {
			return ({ type, value }: { type?: ParamType, value?: string | null }) => {
				this.handleChange(cases => {
					if (type !== undefined) {
						cases[id].mime = type;
					}
					if (value !== undefined) {
						cases[id].value = value;
					}
				});
			};
		}

		loadDictionaries() {
			const cases: Case[] = JSON.parse(this.props.value);
			for (const t of dictionaryKeys) {
				let has = false;
				for (const c of cases) {
					if (t in c) {
						has = true;
						break;
					}
				}
				if (has) {
					this.timer = setTimeout(() => {
						this.setState(prevState => ({ loading: { ...prevState.loading, [`${t}s`]: false } }));
					}, 1000);
					getDictionary(t)
						.then(data => {
							this.setState(prevState => ({ [`${t}s`]: data, loading: { ...prevState.loading, [`${t}s`]: false } }));
							if (this.timer !== undefined) {
								clearTimeout(this.timer);
							}
						})
						.catch(error => this.props.onError(error));
				}
			}
		}

		componentDidMount() {
			this.loadDictionaries();
		}

		componentWillUnmount() {
			if (this.timer !== undefined) {
				clearTimeout(this.timer);
			}
		}

		componentDidUpdate(prevProps: EditNonnullValueProps) {
			if (prevProps.value !== this.props.value) {
				this.loadDictionaries();
			}
		}

		render() {
			const { classes, t } = this.props;
			const cases: Case[] = JSON.parse(this.props.value);
			return (
				<ValueOutline>
					<div>
						{cases.map((c, i) => {
							let caseType: CaseConditions | 'default' = 'default';
							for (const t of caseConditions) {
								if (t in c) {
									caseType = t;
									break;
								}
							}

							let caseKey;
							const commonProps: TextFieldProps = {
								variant: 'outlined',
								margin: 'dense',
								className: classes.caseField,
								onChange: this.createCaseKeyHandler(i),
							};
							switch (caseType) {
								case 'server': {
									caseKey = <TextField label={t('param.case.server')} value={c.server} {...commonProps}/>;
									break;
								}
								case 'group': {
									const { groups, loading } = this.state;
									if (groups) {
										caseKey = (
											<TextField select label={t('param.case.group')} value={c.group} {...commonProps}>
												{Object.keys(groups).map(key => <MenuItem key={key} value={key}>{groups[key]}</MenuItem>)}
											</TextField>
										);
									} else if (loading.groups) {
										caseKey = <CircularProgress/>;
									}
									break;
								}
								case 'datacenter': {
									const { datacenters, loading } = this.state;
									if (datacenters) {
										caseKey = (
											<TextField select label={t('param.case.datacenter')} value={c.datacenter} {...commonProps}>
												{Object.keys(datacenters).map(key => <MenuItem key={key} value={key}>{datacenters[key]}</MenuItem>)}
											</TextField>)
										;
									} else if (loading.datacenters) {
										caseKey = <CircularProgress/>;
									}
									break;
								}
								case 'service': {
									const { services, loading } = this.state;
									if (services) {
										caseKey = (
											<TextField select label={t('param.case.service')} value={c.service} {...commonProps}>
												{Object.keys(services).map(key => <MenuItem key={key} value={key}>{services[key]}</MenuItem>)}
											</TextField>
										);
									} else if (loading.services) {
										caseKey = <CircularProgress/>;
									}
									break;
								}
							}

							return (
								<div key={i} className={classes.case}>
									<div className={classes.caseKey}>
										<IconButton className={classes.remove} onClick={this.createRemoveCaseHandler(i)}>
											<RemoveIcon/>
										</IconButton>
										<TextField
											select
											label={t('param.case.by')}
											value={caseType}
											variant="outlined"
											margin="dense"
											className={classes.caseField}
											onChange={this.createCaseTypeHandler(i)}
										>
											<MenuItem value="default">{t('param.case.default')}</MenuItem>
											<MenuItem value="server">{t('param.case.server')}</MenuItem>
											<MenuItem value="group">{t('param.case.group')}</MenuItem>
											<MenuItem value="datacenter">{t('param.case.datacenter')}</MenuItem>
											<MenuItem value="service">{t('param.case.service')}</MenuItem>
										</TextField>
										{caseKey}
									</div>
									<div className={classes.caseValue}>
										<TypeValueFields type={c.mime} value={c.value} onChange={this.createDataHandler(i)} onError={this.props.onError}/>
									</div>
								</div>
							);
						})}
					</div>
					<div className={classes.add}>
						<IconButton onClick={this.handleAddCase}>
							<AddIcon/>
						</IconButton>
					</div>
				</ValueOutline>
			);
		}

	}
));
