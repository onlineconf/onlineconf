import * as React from 'react';
import * as classNames from 'classnames';
import { withStyles, Theme, WithStyles, createStyles } from '@material-ui/core';

import { NonNullValueProps, Case } from '../../common';
import ValueView from '../../ValueView';

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
		marginLeft: 2 * theme.spacing.unit,
	},
});

const CaseValueView = (props: NonNullValueProps & WithStyles<typeof styles>) => {
	const cases = JSON.parse(props.value);

	return (
		<div className={props.classes.root}>
			{cases.map((c: Case, i: number) => {
				let key = 'default';
				let isDefault = true;

				for (const k of ['datacenter', 'group', 'server']) {
					if (k in c) {
						key = c[k];
						isDefault = false;
						break;
					}
				}

				const keyClassName = classNames(props.classes.key, { [props.classes.keyDefault]: isDefault });
				return (
					<React.Fragment key={i}>
						<div className={props.classes.keyBlock}>
							<span className={keyClassName}>{key}</span>:
						</div>
						<div className={props.classes.value}>
							<ValueView type={c.mime} value={c.value} />
						</div>
					</React.Fragment>
				);
			})}
		</div>
	);
};

export default withStyles(styles)(CaseValueView);
