import deepmerge from 'deepmerge';
import { createMuiTheme, ThemeOptions } from '@material-ui/core/styles';
import { indigo, orange, green } from '@material-ui/core/colors';

declare module '@material-ui/core/styles/createMuiTheme' {
	interface Theme {
		onlineconf: {
			palette: {
				noAccess: React.CSSProperties['color']
				null: React.CSSProperties['color']
				symlink: React.CSSProperties['color']
				case: {
					key: React.CSSProperties['color']
				}
			}
		};
	}
	// allow configuration using `createMuiTheme`
	interface ThemeOptions {
		onlineconf?: {
			palette?: {
				noAccess?: React.CSSProperties['color']
				null?: React.CSSProperties['color']
				symlink?: React.CSSProperties['color']
				case?: {
					key?: React.CSSProperties['color']
				}
			}
		};
	}
}

export default function createTheme(options: ThemeOptions = {}) {
	const palette = process.env.REACT_APP_GREEN
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
							key: '#005500',
						},
					},
				},
				palette,
				typography: {
					useNextVariants: true,
				},
			},
			options,
		)
	);
}
