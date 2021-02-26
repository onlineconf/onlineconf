import * as React from 'react';
import { MuiThemeProvider } from '@material-ui/core/styles';

import { UIConfig, AvatarConfig, getUIConfig } from '../api';
import createTheme from '../theme';

const theme = createTheme();

export const AvatarContext = React.createContext<AvatarConfig | undefined>(undefined);

interface UIConfigProviderProps {
	children: React.ReactNode;
	onError: (error: Error) => void;
}

export default function UIConfigProvider(props: UIConfigProviderProps) {
	const [config, setConfig] = React.useState<UIConfig>();
	const { onError } = props;
	React.useEffect(() => {
		getUIConfig()
			.then(data => setConfig(data))
			.catch(error => onError(error));
	}, [onError]);
	return (
		<MuiThemeProvider theme={theme}>
			<AvatarContext.Provider value={config?.avatar}>
				{props.children}
			</AvatarContext.Provider>
		</MuiThemeProvider>
	);
}
