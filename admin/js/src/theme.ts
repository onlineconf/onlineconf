// @ts-ignore
import deepmerge from 'deepmerge';
import createMuiTheme, { ThemeOptions } from '@material-ui/core/styles/createMuiTheme';

declare module '@material-ui/core/styles/createMuiTheme' {
	interface Theme {
		onlineconf: {
			palette: {
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
			palette?: {
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
	return createMuiTheme(
		deepmerge(
			{
				onlineconf: {
					palette: {
						null: 'gray',
						symlink: '#000088',
						case: {
							root: 'gray',
							key: '#005500',
							value: 'black',
						},
					},
				},
			},
			options,
		)
	);
}
