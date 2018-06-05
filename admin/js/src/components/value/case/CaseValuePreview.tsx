import * as React from 'react';
import { withStyles, Theme, WithStyles, createStyles } from '@material-ui/core';

import { NonNullValueProps, Case } from '../../common';
import ValuePreview from '../../ValuePreview';

const styles = (theme: Theme) => createStyles({
	root: {
		color: theme.onlineconf.palette.case.root,
	},
	key: {
		color: theme.onlineconf.palette.case.key,
	},
	value: {
		color: theme.onlineconf.palette.case.value,
	},
});

const CaseValuePreview = (props: NonNullValueProps & WithStyles<typeof styles>) => {
	const cases = JSON.parse(props.value);

	return (
		<span className={props.classes.root}>
			{cases.map((c: Case, i: number) => {
				const isCase = c.mime === 'application/x-case';
				let key = 'default';
				let isDefault = true;

				for (let k of ['datacenter', 'group', 'server']) {
					if (k in c) {
						key = c[k];
						isDefault = false;
						break;
					}
				}

				return (
					<span key={i}>
						{ i !== 0 && <React.Fragment>; </React.Fragment> }
						{ !isDefault && <React.Fragment><span className={props.classes.key}>{key}</span>: </React.Fragment> }
						{ isCase && '{ ' }
						<span className={props.classes.value}>
							<ValuePreview type={c.mime} value={c.value} />
						</span>
						{ isCase && ' }' }
					</span>
				);
			})}
		</span>
	);
};

export default withStyles(styles)(CaseValuePreview);
