import * as React from 'react';
import { Theme, makeStyles } from '@material-ui/core/styles';

import { Case, caseConditions } from './common';
import { NonNullValueProps } from '../common';
import { ValuePreview } from '../../value';

const useStyles = makeStyles((theme: Theme) => ({
	root: {
		color: theme.palette.text.secondary,
	},
	key: {
		color: theme.onlineconf.palette.case.key,
	},
}), { name: 'CaseValuePreview' });

export default function CaseValuePreview(props: NonNullValueProps) {
	const classes = useStyles();
	const cases = JSON.parse(props.value);

	return (
		<span className={classes.root}>
			{cases.map((c: Case, i: number) => {
				const isCase = c.mime === 'application/x-case';
				let key = 'default';
				let isDefault = true;

				for (const k of caseConditions) {
					if (k in c) {
						key = c[k]!;
						isDefault = false;
						break;
					}
				}

				return (
					<React.Fragment key={i}>
						{ i !== 0 && <React.Fragment>; </React.Fragment> }
						{ !isDefault && <React.Fragment><span className={classes.key}>{key}</span>: </React.Fragment> }
						{ isCase && '{ ' }
						<ValuePreview type={c.mime} value={c.value} />
						{ isCase && ' }' }
					</React.Fragment>
				);
			})}
		</span>
	);
}
