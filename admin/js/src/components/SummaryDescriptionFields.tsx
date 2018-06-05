import * as React from 'react';
import { TextField } from '@material-ui/core';

interface SummaryDescriptionFieldsProps {
	summary: string;
	description: string;
	onChange: (value: { summary?: string, description?: string }) => void;
}

const SummaryDescriptionFields = (props: SummaryDescriptionFieldsProps) => (
	<React.Fragment>
		<TextField
			label="Summary"
			value={props.summary}
			onChange={event => props.onChange({ summary: event.target.value })}
			fullWidth
			margin="dense"
		/>
		<TextField
			label="Description"
			value={props.description}
			onChange={event => props.onChange({ description: event.target.value })}
			multiline
			fullWidth
			margin="dense"
		/>
	</React.Fragment>
);

export default SummaryDescriptionFields;
