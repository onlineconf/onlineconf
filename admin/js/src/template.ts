import * as CodeMirror from 'codemirror';

CodeMirror.defineMode('template', () => ({
	startState: () => ({ variable: false }),
	token: (stream, state) => {
		if (state.variable && stream.skipTo('}')) {
			stream.next();
			state.variable = false;
			return 'variable-2';
		} else if (!state.variable && stream.skipTo('${')) {
			state.variable = true;
			return null;
		} else {
			stream.skipToEnd();
			return null;
		}
	}
}));
(CodeMirror as any).defineMIME('application/x-template', 'template');
