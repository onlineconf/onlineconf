// @ts-ignore
import deepmerge from 'deepmerge';
import createMuiTheme, { ThemeOptions } from '@material-ui/core/styles/createMuiTheme';
import indigo from '@material-ui/core/colors/indigo';
import orange from '@material-ui/core/colors/orange';
import green from '@material-ui/core/colors/green';

type OnlineConfInstance = 'production' | 'development';

declare module '@material-ui/core/styles/createMuiTheme' {
	interface Theme {
		onlineconf: {
			instance: OnlineConfInstance
			palette: {
				noAccess: React.CSSProperties['color']
				null: React.CSSProperties['color']
				symlink: React.CSSProperties['color']
				case: {
					root: React.CSSProperties['color']
					key: React.CSSProperties['color']
					value: React.CSSProperties['color']
				}
			}
		};
	}
	// allow configuration using `createMuiTheme`
	interface ThemeOptions {
		onlineconf?: {
			instance?: OnlineConfInstance
			palette?: {
				noAccess?: React.CSSProperties['color']
				null?: React.CSSProperties['color']
				symlink?: React.CSSProperties['color']
				case?: {
					root?: React.CSSProperties['color']
					key?: React.CSSProperties['color']
					value?: React.CSSProperties['color']
				}
			}
		};
	}
}

export default function createTheme(options: ThemeOptions = {}) {
	const palette = options.onlineconf && options.onlineconf.instance === 'development'
		? {
			primary: green,
			secondary: orange,
			background: {
				default: green[50],
			},
		}
		: {
			primary: indigo,
			secondary: orange,
		};

	return createMuiTheme(
		deepmerge(
			{
				onlineconf: {
					palette: {
						noAccess: 'gray',
						null: 'gray',
						symlink: '#000088',
						case: {
							root: 'gray',
							key: '#005500',
							value: 'black',
						},
					},
				},
				palette,
			},
			options,
		)
	);
}
