import * as React from 'react';
import { ThemeProvider, Theme, StyledEngineProvider, adaptV4Theme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';

import { UIConfig, AvatarConfig, getUIConfig } from '../api';
import createTheme from '../theme';


declare module '@mui/styles/defaultTheme' {
  // eslint-disable-next-line @typescript-eslint/no-empty-interface
  interface DefaultTheme extends Theme {}
}


export const AvatarContext = React.createContext<AvatarConfig | undefined>(undefined);

export type PaletteType = 'system' | 'light' | 'dark';

interface UIConfigProviderProps {
	children: React.ReactNode;
	onError: (error: unknown) => void;
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
	const theme = React.useMemo(() => createTheme(adaptV4Theme({
		palette: {
			mode: themePaletteType,
		},
	})), [themePaletteType]);

	return (
		<StyledEngineProvider injectFirst>
			<ThemeProvider theme={theme}>
				<AvatarContext.Provider value={config?.avatar}>
					{props.children}
				</AvatarContext.Provider>
			</ThemeProvider>
		</StyledEngineProvider>
	);
}
