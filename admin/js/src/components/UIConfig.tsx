import * as React from 'react';
import { ThemeProvider, Theme, StyledEngineProvider, adaptV4Theme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';

import { UIConfig, AvatarConfig, EncryptionConfig, getUIConfig } from '../api';
import createTheme from '../theme';


declare module '@mui/styles/defaultTheme' {
  // eslint-disable-next-line @typescript-eslint/no-empty-interface
  interface DefaultTheme extends Theme {}
}


export const AvatarContext = React.createContext<AvatarConfig | undefined>(undefined);

const noEncryption: EncryptionConfig = { enabled: false };

// Off until /ui-config arrives: the switch must not show on an installation
// that has the feature off.
export const EncryptionContext = React.createContext<EncryptionConfig>(noEncryption);

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

	const encryption = config?.encryption ?? noEncryption;

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
					<EncryptionContext.Provider value={encryption}>
						{props.children}
					</EncryptionContext.Provider>
				</AvatarContext.Provider>
			</ThemeProvider>
		</StyledEngineProvider>
	);
}
