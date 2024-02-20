import * as React from 'react';
import clsx from 'clsx';
import { Theme } from '@mui/material/styles';

import { WithStyles } from '@mui/styles';
import withStyles from '@mui/styles/withStyles';
import createStyles from '@mui/styles/createStyles';

import { Case, caseConditions } from './common';
import { NonNullValueProps } from '../common';
import { ValueView } from '../../value';

const styles = (theme: Theme) => createStyles({
	root: {
		color: theme.palette.text.secondary,
		[theme.breakpoints.up('sm')]: {
			display: 'grid',
			gridTemplateColumns: 'auto 1fr',
		}
	},
	keyBlock: {
		whiteSpace: 'nowrap',
	},
	key: {
		color: theme.palette.onlineconf.caseKey,
	},
	keyDefault: {
		fontStyle: 'italic',
	},
	value: {
		marginLeft: theme.spacing(2),
	},
});

const CaseValueView = (props: NonNullValueProps & WithStyles<typeof styles>) => {
	const cases = JSON.parse(props.value);

	return (
		<div className={props.classes.root}>
			{cases.map((c: Case, i: number) => {
				let key = 'default';
				let isDefault = true;

				for (const k of caseConditions) {
					if (k in c) {
						key = c[k]!;
						isDefault = false;
						break;
					}
				}

				const keyClassName = clsx(props.classes.key, { [props.classes.keyDefault]: isDefault });
				return (
					<React.Fragment key={i}>
						<div className={props.classes.keyBlock}>
							<span className={keyClassName}>{key}</span>:
						</div>
						<ValueView type={c.mime} value={c.value} disableDecoration className={props.classes.value}/>
					</React.Fragment>
				);
			})}
		</div>
	);
};

export default withStyles(styles)(CaseValueView);
