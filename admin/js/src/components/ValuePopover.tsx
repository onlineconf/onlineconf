import * as React from 'react';
import Popover from '@mui/material/Popover';
import { Theme } from '@mui/material/styles';

import { WithStyles } from '@mui/styles';
import withStyles from '@mui/styles/withStyles';

import { ValueProps } from './common';
import { ValueView } from './value';

const styles = (theme: Theme) => ({
	popover: {
		'pointer-events': 'none',
	},
	paper: {
		'padding': theme.spacing(1),
	},
});

interface ValuePopoverProps extends ValueProps {
	anchorEl: HTMLElement;
	onClose: () => void;
}

const ValuePopover = (props: ValuePopoverProps & WithStyles<typeof styles>) => (
	<Popover
		open
		onClose={props.onClose}
		anchorEl={props.anchorEl}
		anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
		transformOrigin={{ vertical: 'top', horizontal: 'left' }}
		className={props.classes.popover}
		classes={{ paper: props.classes.paper }}
		disableRestoreFocus
	>
		<ValueView type={props.type} value={props.value} disableDecoration/>
	</Popover>
);

export default withStyles(styles)(ValuePopover);
