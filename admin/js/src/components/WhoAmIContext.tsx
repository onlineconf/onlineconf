import * as React from 'react';
import { getWhoAmI } from '../api';
import ErrorContext from './ErrorContext';

export interface WhoAmI {
	username?: string;
	userIsRoot: boolean;
}

const WhoAmIContext = React.createContext<WhoAmI>({ userIsRoot: false });
export default WhoAmIContext;

interface WhoAmIProviderProps {
	children: React.ReactNode;
}

export function WhoAmIProvider(props: WhoAmIProviderProps) {
	const onError = React.useContext(ErrorContext);
	const [ value, setValue ] = React.useState<WhoAmI>({ userIsRoot: false });
	React.useEffect(() => {
		getWhoAmI()
			.then(data => setValue({
				username: data.username,
				userIsRoot: data.can_edit_groups,
			}))
			.catch(onError);
	}, [onError]);
	return <WhoAmIContext.Provider value={value}>{props.children}</WhoAmIContext.Provider>;
}
