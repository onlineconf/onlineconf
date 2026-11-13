import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { Theme } from '@mui/material/styles';
import makeStyles from '@mui/styles/makeStyles';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import CircularProgress from '@mui/material/CircularProgress';

import LockIcon from '@mui/icons-material/Lock';
import VisibilityIcon from '@mui/icons-material/Visibility';
import VisibilityOffIcon from '@mui/icons-material/VisibilityOff';

import * as API from '../../api';

// Encrypted values are served masked: `data` is null until the user asks the
// server to reveal it (an audited request). These are the masked renderings
// and the reveal control shared by every place a value shows.

const useStyles = makeStyles((theme: Theme) => ({
	root: {
		display: 'inline-flex',
		alignItems: 'center',
		columnGap: theme.spacing(0.5),
		color: theme.palette.text.secondary,
	},
	dots: {
		letterSpacing: 2,
	},
	icon: {
		fontSize: 'inherit',
	},
}), { name: 'EncryptedValue' });

export const maskedText = '••••••';

export function EncryptedValuePreview() {
	const classes = useStyles();
	const { t } = useTranslation();
	return (
		<span className={classes.root} title={t('param.encrypted.masked')}>
			<LockIcon className={classes.icon}/>
			<span className={classes.dots}>{maskedText}</span>
		</span>
	);
}

export const EncryptedValueView = EncryptedValuePreview;

interface RevealButtonProps {
	revealed: boolean;
	loading?: boolean;
	onReveal: () => void;
	onHide: () => void;
	size?: 'small' | 'medium' | 'large';
	className?: string;
}

export function RevealButton(props: RevealButtonProps) {
	const { t } = useTranslation();
	const { revealed, loading, onReveal, onHide, size = 'small', className } = props;
	const handleClick = (event: React.MouseEvent) => {
		event.stopPropagation();
		if (revealed) {
			onHide();
		} else {
			onReveal();
		}
	};
	if (loading) {
		return <CircularProgress size={size === 'small' ? 16 : 24} className={className}/>;
	}
	return (
		<Tooltip title={t(revealed ? 'param.encrypted.hide' : 'param.encrypted.reveal') as string}>
			<IconButton size={size} onClick={handleClick} className={className}>
				{revealed ? <VisibilityOffIcon fontSize="inherit"/> : <VisibilityIcon fontSize="inherit"/>}
			</IconButton>
		</Tooltip>
	);
}

// useReveal fetches a parameter's plaintext on demand and keeps it in the
// calling component's state only: nothing revealed ever enters the tree or
// log caches, and a new version of the parameter forgets the plaintext.
export function useReveal(path: string, version: number, onError: (error: unknown) => void) {
	const [ revealed, setRevealed ] = React.useState<API.IParam>();
	const [ loading, setLoading ] = React.useState(false);
	React.useEffect(() => setRevealed(undefined), [path, version]);
	const reveal = React.useCallback(async () => {
		setLoading(true);
		try {
			setRevealed(await API.getParam(path, undefined, 'none', {}, true));
		} catch (error) {
			onError(error);
		}
		setLoading(false);
	}, [path, onError]);
	const hide = React.useCallback(() => setRevealed(undefined), []);
	return { revealed, loading, reveal, hide };
}
