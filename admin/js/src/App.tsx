import * as React from 'react';
import clsx from 'clsx';
import { BrowserRouter, Route } from 'react-router-dom';
import { Theme } from '@mui/material/styles';
import makeStyles from '@mui/styles/makeStyles';
import CssBaseline from '@mui/material/CssBaseline';

import ErrorContext from './components/ErrorContext';
import { WhoAmIProvider } from './components/WhoAmIContext';
import UIConfigProvider, { PaletteType } from './components/UIConfig';
import TopBar from './components/TopBar';
import LeftMenu, { leftMenuWidth } from './components/LeftMenu';
import ConfigTree from './components/ConfigTree';
import Servers from './components/Servers';
import Access from './components/Access';
import ErrorSnackbar from './components/ErrorSnackbar';
import GlobalLog from './components/GlobalLog';


const useStyles = makeStyles((theme: Theme) => ({
	root: {
		['--left-menu-width' as never]: '0px',
	},
	menu: {
		width: leftMenuWidth,
	},
	main: {
		overflow: 'hidden',
		transition: theme.transitions.create('margin', {
			easing: theme.transitions.easing.sharp,
			duration: theme.transitions.duration.leavingScreen,
		}),
		'&$menuSliding .tree-value-column': {
			transition: theme.transitions.create('width', {
				easing: theme.transitions.easing.sharp,
				duration: theme.transitions.duration.leavingScreen,
			}),
		},
	},
	mainShift: {
		transition: theme.transitions.create('margin', {
			easing: theme.transitions.easing.easeOut,
			duration: theme.transitions.duration.enteringScreen,
		}),
		'&$menuSliding .tree-value-column': {
			transition: theme.transitions.create('width', {
				easing: theme.transitions.easing.easeOut,
				duration: theme.transitions.duration.enteringScreen,
			}),
		},
		[theme.breakpoints.up('md')]: {
			marginLeft: leftMenuWidth,
			['--left-menu-width' as never]: `${leftMenuWidth}px`,
		},
	},
	menuSliding: {},
}), { name: 'App' });

interface AppRootProps {
	error?: unknown;
	onError(error: unknown): void;
	paletteType: PaletteType;
	onChangePaletteType(value: PaletteType): void;
}

function AppRoot(props: AppRootProps) {
	const classes = useStyles();
	const { onError } = props;

	const [ menu, setMenu ] = React.useState(false);
	const [ menuSliding, setMenuSliding ] = React.useState(false);
	const [ search, setSearch ] = React.useState('');
	const [ searching, setSearching ] = React.useState(false);

	return (
		<React.Fragment>
			<CssBaseline/>
			<BrowserRouter>
				<div className={classes.root}>
					<TopBar onMenu={() => { setMenu(m => !m); setMenuSliding(true) }} onSearch={setSearch} searching={searching}/>
					<LeftMenu
						open={menu}
						onClose={() => { setMenu(false); setMenuSliding(true) }}
						onSlideEnd={() => setMenuSliding(false)}
						paletteType={props.paletteType}
						onChangePaletteType={props.onChangePaletteType}
					/>
					<main className={clsx(classes.main, menu && classes.mainShift, menuSliding && classes.menuSliding)}>
						<Route exact path="/" render={props => <ConfigTree {...props} search={search} onSearching={setSearching} onError={onError}/>}/>
						<Route exact path="/history/" render={props => <GlobalLog {...props} onError={onError}/>}/>
						<Route exact path="/server/" render={props => <Servers {...props} onError={onError}/>}/>
						<Route exact path="/access-group/" render={props => <Access {...props} onError={onError}/>}/>
					</main>
				</div>
			</BrowserRouter>
			<ErrorSnackbar error={props.error}/>
		</React.Fragment>
	);
}

export default function App() {
	const [ error, setError ] = React.useState<unknown>();

	const [ paletteType, setPaletteType ] = React.useState<PaletteType>(() => {
		const value = window.localStorage.getItem('paletteType');
		return value === 'light' || value === 'dark' ? value : 'system';
	});
	const onChangePaletteType = (paletteType: PaletteType) => {
		setPaletteType(paletteType);
		window.localStorage.setItem('paletteType', paletteType);
	};

	return (
		<ErrorContext.Provider value={setError}>
			<WhoAmIProvider>
				<UIConfigProvider onError={setError} paletteType={paletteType}>
					<AppRoot error={error} onError={setError} paletteType={paletteType} onChangePaletteType={onChangePaletteType}/>
				</UIConfigProvider>
			</WhoAmIProvider>
		</ErrorContext.Provider>
	);
}
