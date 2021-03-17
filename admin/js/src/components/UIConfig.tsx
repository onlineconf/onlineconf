import * as React from 'react';
import { MuiThemeProvider } from '@material-ui/core/styles';
import useMediaQuery from '@material-ui/core/useMediaQuery';

import { UIConfig, AvatarConfig, getUIConfig } from '../api';
import createTheme from '../theme';

export const AvatarContext = React.createContext<AvatarConfig | undefined>(undefined);

export type PaletteType = 'system' | 'light' | 'dark';

interface UIConfigProviderProps {
	children: React.ReactNode;
	onError: (error: Error) => void;
	paletteType: PaletteType;
}

export default function UIConfigProvider(props: UIConfigProviderProps) {
	const [config, setConfig] = React.useState<UIConfig>();
	const { onError } = props;

	React.useEffect(() => {
		getUIConfig()
			.then(data => setConfig(data))
			.catch(error => onError(error));
	}, [onError]);

	const systemPaletteType = useMediaQuery('(prefers-color-scheme: dark)') ? 'dark' : 'light';
	const themePaletteType = props.paletteType === 'system' ? systemPaletteType : props.paletteType;
	const theme = React.useMemo(() => createTheme({
		palette: {
			type: themePaletteType,
		},
	}), [themePaletteType]);

	return (
		<MuiThemeProvider theme={theme}>
			<AvatarContext.Provider value={config?.avatar}>
				{props.children}
			</AvatarContext.Provider>
		</MuiThemeProvider>
	);
}
