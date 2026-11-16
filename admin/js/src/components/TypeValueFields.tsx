import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { Theme } from '@mui/material/styles';
import { WithStyles } from '@mui/styles';
import createStyles from '@mui/styles/createStyles';
import withStyles from '@mui/styles/withStyles';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';

import LockIcon from '@mui/icons-material/Lock';
import VisibilityIcon from '@mui/icons-material/Visibility';
import VisibilityOffIcon from '@mui/icons-material/VisibilityOff';

import { types, encryptionEligible, ValueProps } from './common';
import { Case } from './value/case/common';
import { ParamType } from '../api';
import { ValueEdit } from './value';
import { EncryptionContext } from './UIConfig';

const styles = (theme: Theme) => createStyles({
	flags: {
		display: 'flex',
		alignItems: 'center',
		columnGap: theme.spacing(1),
	},
	flagLabel: {
		display: 'inline-flex',
		alignItems: 'center',
		columnGap: theme.spacing(0.5),
	},
	flagIcon: {
		fontSize: theme.typography.pxToRem(16),
		color: theme.palette.text.secondary,
	},
});

interface TypeValueFieldsProps extends ValueProps {
	// the fields edit a case branch, which is never encrypted (a branch that
	// needs a secret is a symlink to an encrypted parameter): no switch
	branch?: boolean;
	// start with the typed text hidden (a secret typed blind)
	maskByDefault?: boolean;
	onChange: (state: ValueProps) => void;
	onError: (error: unknown) => void;
}

interface TypeValueFieldsState {
	masked: boolean;
}

class TypeValueFields extends React.Component<TypeValueFieldsProps & WithStyles<typeof styles> & WithTranslation, TypeValueFieldsState> {

	static contextType = EncryptionContext;
	context!: React.ContextType<typeof EncryptionContext>;

	state: TypeValueFieldsState = {
		masked: this.props.maskByDefault === true,
	};

	handleTypeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		const type = event.target.value as ParamType;
		if (type !== this.props.type) {
			// the marker survives a type change while the new type can carry
			// it, and is cleared (not failed) otherwise
			const state: ValueProps = {
				type,
				value: this.props.value,
				encrypted: encryptionEligible(type) ? this.props.encrypted : undefined,
			};
			if (type === 'application/x-null') {
				state.value = null;
			} else if (type === 'application/x-case') {
				state.value = JSON.stringify([{ mime: this.props.type, value: this.props.value }]);
			} else if (this.props.value === null) {
				state.value = '';
			} else if (this.props.type === 'application/x-case') {
				const cases: Case[] = JSON.parse(this.props.value || '[]');
				const defaultCase = cases.filter(c => (c.server === undefined && c.group === undefined && c.datacenter === undefined && c.service === undefined))[0];
				state.value = defaultCase !== undefined && defaultCase.value !== null ? defaultCase.value : '';
			}
			this.props.onChange(state);
		}
	};

	handleValueChange = (value: string | null) => {
		this.props.onChange({ type: this.props.type, value, encrypted: this.props.encrypted });
	};

	handleEncryptedChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		this.props.onChange({ type: this.props.type, value: this.props.value, encrypted: event.target.checked });
	};

	handleMaskToggle = () => {
		this.setState(prevState => ({ masked: !prevState.masked }));
	};

	render() {
		const { classes, t } = this.props;
		const eligible = encryptionEligible(this.props.type);
		// the switch is offered where encrypted writes are accepted — disabled
		// for types that cannot carry it, so that it is discoverable before
		// the type is chosen — and always on a value that is encrypted already
		// so that it can be declassified; never for a case branch
		const showSwitch = !this.props.branch && (this.context.enabled || this.props.encrypted === true);
		const encrypted = eligible && this.props.encrypted === true;
		return (
			<React.Fragment>
				<TextField
					id="mime"
					select
					label={t('param.type')}
					value={this.props.type}
					variant="outlined"
					margin="dense"
					fullWidth
					onChange={this.handleTypeChange}
				>
					{types.map((type: ParamType) => (
						<MenuItem key={type} value={type}>{t(`param.types.${type}`)}</MenuItem>
					))}
				</TextField>
				{showSwitch && (
					<div className={classes.flags}>
						<Tooltip title={eligible ? '' : t('param.encrypted.ineligible') as string}>
							<FormControlLabel
								control={<Switch checked={encrypted} disabled={!eligible} onChange={this.handleEncryptedChange} color="primary"/>}
								label={<span className={classes.flagLabel}><LockIcon className={classes.flagIcon}/>{t('param.encrypted.flag')}</span>}
							/>
						</Tooltip>
						{encrypted && (
							<Tooltip title={t(this.state.masked ? 'param.encrypted.unmask' : 'param.encrypted.mask') as string}>
								<IconButton size="small" onClick={this.handleMaskToggle}>
									{this.state.masked ? <VisibilityIcon fontSize="small"/> : <VisibilityOffIcon fontSize="small"/>}
								</IconButton>
							</Tooltip>
						)}
					</div>
				)}
				<ValueEdit
					type={this.props.type}
					value={this.props.value}
					masked={encrypted && this.state.masked}
					onChange={this.handleValueChange}
					onError={this.props.onError}
				/>
			</React.Fragment>
		);
	}

}

export default withStyles(styles)(withTranslation()(TypeValueFields));
