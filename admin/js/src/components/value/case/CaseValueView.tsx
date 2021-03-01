import * as React from 'react';
import clsx from 'clsx';
import { withStyles, Theme, WithStyles, createStyles } from '@material-ui/core/styles';

import { Case, caseConditions } from './common';
import { NonNullValueProps } from '../common';
import { ValueView } from '../../value';

const styles = (theme: Theme) => createStyles({
	root: {
		color: theme.onlineconf.palette.case.root,
		[theme.breakpoints.up('sm')]: {
			display: 'grid',
			gridTemplateColumns: 'auto 1fr',
		}
	},
	keyBlock: {
		whiteSpace: 'nowrap',
	},
	key: {
		color: theme.onlineconf.palette.case.key,
	},
	keyDefault: {
		fontStyle: 'italic',
	},
	value: {
		color: theme.onlineconf.palette.case.value,
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
						<div className={props.classes.value}>
							<ValueView type={c.mime} value={c.value} accessible/>
						</div>
					</React.Fragment>
				);
			})}
		</div>
	);
};

export default withStyles(styles)(CaseValueView);
