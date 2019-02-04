import * as React from 'react';
import { TextField, createStyles, WithStyles, withStyles } from '@material-ui/core';

// Workaround for bug in outlined dense multiline TextField
const styles = createStyles({
	denseMultilineFix: {
		paddingTop: 15,
		paddingBottom: 15,
	},
});

interface SummaryDescriptionFieldsProps {
	summary: string;
	description: string;
	onChange: (value: { summary?: string, description?: string }) => void;
}

const SummaryDescriptionFields = (props: SummaryDescriptionFieldsProps & WithStyles<typeof styles>) => (
	<React.Fragment>
		<TextField
			label="Summary"
			value={props.summary}
			onChange={event => props.onChange({ summary: event.target.value })}
			variant="outlined"
			fullWidth
			margin="dense"
		/>
		<TextField
			label="Description"
			value={props.description}
			onChange={event => props.onChange({ description: event.target.value })}
			multiline
			variant="outlined"
			fullWidth
			margin="dense"
			InputProps={{ className: props.classes.denseMultilineFix }}
		/>
	</React.Fragment>
);

export default withStyles(styles)(SummaryDescriptionFields);
