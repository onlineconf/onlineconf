import * as CSS from 'csstype';
import deepmerge from 'deepmerge';
import { DeprecatedThemeOptions, adaptV4Theme, createTheme as createMuiTheme } from '@mui/material/styles';
import { indigo, orange, green } from '@mui/material/colors';

interface OnlineConfPalette {
	noAccess: CSS.Property.Color;
	null: CSS.Property.Color;
	symlink: CSS.Property.Color;
	caseKey: CSS.Property.Color;
	templateVariable: CSS.Property.Color;
}

interface CodeMirrorPalette {
	atom: CSS.Property.Color;      // yaml mapping key
	comment: CSS.Property.Color;
	def: CSS.Property.Color;       // yaml document start and end
	keyword: CSS.Property.Color;   // true, false, etc...
	meta: CSS.Property.Color;      // yaml -,{}[] and block literals
	number: CSS.Property.Color;
	string: CSS.Property.Color;    // quoted strings
	variable2: CSS.Property.Color; // yaml reference, template variable
}

declare module '@mui/material/styles/createPalette' {
	interface Palette {
		onlineconf: OnlineConfPalette;
		codemirror: CodeMirrorPalette;
	}

	interface PaletteOptions {
		onlineconf?: Partial<OnlineConfPalette>;
		codemirror?: Partial<CodeMirrorPalette>;
	}
}

const themeOptions: { [K in 'light' | 'dark']: DeprecatedThemeOptions } = {
	light: {
		palette: {
			background: {
				default: '#ffffff',
				paper: '#ffffff',
			},
			primary: { main: indigo[500] },
			secondary: { main: orange[500] },
			onlineconf: {
				noAccess: 'gray',
				null: 'gray',
				symlink: '#000088',
				caseKey: '#005500',
				templateVariable: '#0055aa',
			},
			// vscode Light+ theme
			codemirror: {
				atom: '#0451a5',
				comment: '#008000',
				def: '#000000',
				keyword: '#0000ff',
				meta: '#000000',
				number: '#098658',
				string: '#a31515',
				variable2: '#0055aa',
			},
		},
	},
	dark: {
		palette: {
			background: {
				default: '#121212',
				paper: '#121212',
			},
			primary: { main: indigo[200] },
			secondary: { main: orange[200] },
			onlineconf: {
				noAccess: 'gray',
				null: 'gray',
				symlink: '#bbbbff',
				caseKey: '#88dd88',
				templateVariable: '#9cdcfe',
			},
			// vscode Dark+ theme
			codemirror: {
				atom: '#9cdcfe',
				comment: '#6a9955',
				def: '#d4d4d4',
				keyword: '#569cd6',
				meta: '#d4d4d4',
				number: '#b5cea8',
				string: '#ce9178',
				variable2: '#9cdcfe',
			},
		},
		overrides: {
			// elevation overlay emulation
			MuiPaper: {
				elevation1: { backgroundColor: '#1d1d1d' },
				elevation2: { backgroundColor: '#222222' },
				elevation3: { backgroundColor: '#242424' },
				elevation4: { backgroundColor: '#272727' },
				elevation6: { backgroundColor: '#2c2c2c' },
				elevation8: { backgroundColor: '#2e2e2e' },
				elevation12: { backgroundColor: '#323232' },
				elevation16: { backgroundColor: '#353535' },
				elevation24: { backgroundColor: '#383838' },
			}
		},
	},
};

// used to distinguish between production (default) and development (green) instances
const greenThemeOptions: { [K in 'light' | 'dark']: DeprecatedThemeOptions } = {
	light: {
		palette: {
			background: {
				default: green[50],
				paper: green[50],
			},
			primary: { main: green[500] },
		},
		overrides: {
			MuiAppBar: {
				colorDefault: {
					backgroundColor: green[100],
				},
			},
		},
	},
	dark: {
		palette: {
			background: {
				default: '#0b190b',
				paper: '#0b190b',
			},
			primary: { main: green[200] },
		},
		overrides: {
			MuiAppBar: {
				colorDefault: {
					backgroundColor: '#142e15',
				},
			},
			// elevation overlay emulation
			MuiPaper: {
				elevation1: { backgroundColor: '#172517' },
				elevation2: { backgroundColor: '#1c291c' },
				elevation3: { backgroundColor: '#1e2b1e' },
				elevation4: { backgroundColor: '#212e21' },
				elevation6: { backgroundColor: '#263226' },
				elevation8: { backgroundColor: '#293529' },
				elevation12: { backgroundColor: '#2d392d' },
				elevation16: { backgroundColor: '#2f3b2f' },
				elevation24: { backgroundColor: '#323e32' },
			}
		},
	},
};

export default function createTheme(options: DeprecatedThemeOptions = {}) {
	return createMuiTheme(
		adaptV4Theme(deepmerge.all([
			{
				typography: {
					useNextVariants: true,
				},
				overrides: {
					MuiSelect: {
						select: {
							'&:focus': {
								// material-ui bug workaround
								backgroundColor: undefined,
							}
						},
					},
				},
			},
			themeOptions[options.palette?.mode || 'light'],
			process.env.REACT_APP_GREEN ? greenThemeOptions[options.palette?.mode || 'light'] : {},
			options,
		]))
	);
}
