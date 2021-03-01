import * as React from 'react';
import { withTranslation, WithTranslation } from 'react-i18next';
import { createStyles, WithStyles, withStyles } from '@material-ui/core/styles';
import TextField from '@material-ui/core/TextField';

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
	onSummaryChange: (summary: string) => void;
	onDescriptionChange: (description: string) => void;
}

const SummaryDescriptionFields = ({ t, ...props }: SummaryDescriptionFieldsProps & WithStyles<typeof styles> & WithTranslation) => (
	<React.Fragment>
		<TextField
			label={t('param.summary')}
			value={props.summary}
			onChange={event => props.onSummaryChange(event.target.value)}
			variant="outlined"
			fullWidth
			margin="dense"
		/>
		<TextField
			label={t('param.description')}
			value={props.description}
			onChange={event => props.onDescriptionChange(event.target.value)}
			multiline
			variant="outlined"
			fullWidth
			margin="dense"
			InputProps={{ className: props.classes.denseMultilineFix }}
		/>
	</React.Fragment>
);

export default withStyles(styles)(withTranslation()(SummaryDescriptionFields));
