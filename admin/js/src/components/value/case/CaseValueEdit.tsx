import * as React from 'react';
import { createStyles, Theme, WithStyles, withStyles, IconButton, CircularProgress, TextField, MenuItem } from '@material-ui/core';

import AddIcon from '@material-ui/icons/AddCircle';
import RemoveIcon from '@material-ui/icons/RemoveCircle';

import { getDictionary } from '../../../cache';
import { Case, EditNonnullValueProps } from '../../common';
import { ParamType } from '../../../api';
import TypeValueFields from '../../TypeValueFields';

const styles = (theme: Theme) => createStyles({
	root: {
		marginTop: theme.spacing.unit,
		marginBottom: theme.spacing.unit / 2,
	},
	label: {
		fontFamily: theme.typography.fontFamily,
		fontSize: theme.typography.pxToRem(12),
		color: theme.palette.text.secondary,
		lineHeight: 1 / 0.75,
	},
	case: {
		margin: `${theme.spacing.unit / 2}px 0`,
		boxShadow: theme.shadows[1],
	},
	caseKey: {
		display: 'flex',
		flexWrap: 'wrap',
		marginLeft: theme.spacing.unit / 2,
		marginRight: theme.spacing.unit / 2,
	},
	caseField: {
		flex: 'auto',
		marginLeft: theme.spacing.unit / 2,
		marginRight: theme.spacing.unit / 2,
	},
	remove: {
		margin: `${theme.spacing.unit}px ${theme.spacing.unit / 2}px`,
	},
	caseValue: {
		paddingLeft: theme.spacing.unit,
		paddingRight: theme.spacing.unit,
		paddingBottom: theme.spacing.unit / 2,
	},
	add: {
		textAlign: 'center',
	},
});

interface CaseValueEditState {
	loading: {
		groups?: boolean;
		datacenters?: boolean;
	};
	groups?: { [K: string]: string };
	datacenters?: { [K: string]: string };
}

export default withStyles(styles)(
	class CaseValueEdit extends React.Component<EditNonnullValueProps & WithStyles<typeof styles>, CaseValueEditState> {

		state: CaseValueEditState = {
			loading: {},
		};

		timer: NodeJS.Timer;

		handleChange(cb: (cases: Case[]) => void) {
			let cases: Case[] = JSON.parse(this.props.value);
			cb.call(this, cases);
			this.props.onChange({ target: { value: JSON.stringify(cases) } });
		}

		handleAddCase = () => {
			this.handleChange(cases => {
				cases.push({ server: '', mime: 'application/x-null', value: null });
			});
		}

		createRemoveCaseHandler(id: number) {
			return (event: React.MouseEvent<{}>) => {
				this.handleChange(cases => {
					cases.splice(id, 1);
				});
			};
		}

		createCaseTypeHandler(id: number) {
			return (event: React.ChangeEvent<HTMLInputElement>) => {
				this.handleChange(cases => {
					const newType = event.target.value;
					for (const t of ['server', 'group', 'datacenter']) {
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
					for (const t of ['server', 'group', 'datacenter']) {
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
			for (const t of (['group', 'datacenter'] as ('group' | 'datacenter')[])) {
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
							clearTimeout(this.timer);
						})
						.catch(error => this.props.onError(error));
				}
			}
		}

		componentDidMount() {
			this.loadDictionaries();
		}

		componentWillUnmount() {
			clearTimeout(this.timer);
		}

		componentDidUpdate(prevProps: EditNonnullValueProps) {
			if (prevProps.value !== this.props.value) {
				this.loadDictionaries();
			}
		}

		render() {
			const { classes } = this.props;
			const cases: Case[] = JSON.parse(this.props.value);
			return (
				<div className={classes.root}>
					<div className={classes.label}>Value</div>
					<div>
						{cases.map((c, i) => {
							let caseType = 'default';
							for (const t of ['server', 'group', 'datacenter']) {
								if (t in c) {
									caseType = t;
									break;
								}
							}

							let caseKey;
							const commonProps = {
								margin: 'dense' as 'dense',
								className: classes.caseField,
								onChange: this.createCaseKeyHandler(i),
							};
							switch (caseType) {
								case 'server': {
									caseKey = <TextField label="Server" value={c.server} {...commonProps}/>;
									break;
								}
								case 'group': {
									const { groups, loading } = this.state;
									if (groups) {
										caseKey = (
											<TextField select label="Group" value={c.group} {...commonProps}>
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
											<TextField select label="Datacenter" value={c.datacenter} {...commonProps}>
												{Object.keys(datacenters).map(key => <MenuItem key={key} value={key}>{datacenters[key]}</MenuItem>)}
											</TextField>)
										;
									} else if (loading.datacenters) {
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
											label="Case by"
											value={caseType}
											margin="dense"
											className={classes.caseField}
											onChange={this.createCaseTypeHandler(i)}
										>
											<MenuItem value="default">Default</MenuItem>
											<MenuItem value="server">Server</MenuItem>
											<MenuItem value="group">Group</MenuItem>
											<MenuItem value="datacenter">Datacenter</MenuItem>
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
				</div>
			);
		}

	}
);
