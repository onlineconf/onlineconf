import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { act } from 'react-dom/test-utils';
import { ThemeProvider } from '@mui/material/styles';

import '../i18n';
import createTheme from '../theme';
import { FeaturesContext, Features } from './UIConfig';
import TypeValueFields from './TypeValueFields';
import { ParamType } from '../api';

function renderFields(features: Features, type: ParamType, encrypted?: boolean) {
	const value = type === 'application/x-case' ? '[]' : type === 'application/x-null' ? null : 'a,b';
	const div = document.createElement('div');
	document.body.appendChild(div);
	act(() => {
		ReactDOM.render(
			<ThemeProvider theme={createTheme()}>
				<FeaturesContext.Provider value={features}>
					<TypeValueFields type={type} value={value} encrypted={encrypted} onChange={() => undefined} onError={() => undefined}/>
				</FeaturesContext.Provider>
			</ThemeProvider>,
			div,
		);
	});
	const cleanup = () => {
		ReactDOM.unmountComponentAtNode(div);
		div.remove();
	};
	const input = div.querySelector<HTMLInputElement>('input[type="checkbox"]');
	const hasSwitch = input !== null;
	const disabled = input !== null && input.disabled;
	return { hasSwitch, disabled, cleanup };
}

const on: Features = { encryption: { enabled: true, reveal: 'rw' } };
const off: Features = { encryption: { enabled: false, reveal: 'rw' } };

it('offers the Encrypted switch where encrypted writes are accepted', () => {
	const r = renderFields(on, 'application/x-list');
	expect(r.hasSwitch).toBe(true);
	expect(r.disabled).toBe(false);
	r.cleanup();
});

it('hides the switch while the feature is off, except on an encrypted value', () => {
	let r = renderFields(off, 'application/x-list');
	expect(r.hasSwitch).toBe(false);
	r.cleanup();
	r = renderFields(off, 'application/x-list', true);
	expect(r.hasSwitch).toBe(true);
	r.cleanup();
});

it('renders an encrypted value without leaking its flags into the DOM', () => {
	const errors = jest.spyOn(console, 'error').mockImplementation(() => undefined);
	try {
		// (text values render CodeMirror, which jsdom cannot host)
		for (const type of ['application/x-list', 'application/x-symlink'] as ParamType[]) {
			const r = renderFields(on, type, true);
			r.cleanup();
		}
		expect(errors).not.toHaveBeenCalled();
	} finally {
		errors.mockRestore();
	}
});

it('shows the switch disabled for an ineligible type', () => {
	for (const type of ['application/x-null', 'application/x-symlink', 'application/x-case'] as ParamType[]) {
		const r = renderFields(on, type);
		expect(r.hasSwitch).toBe(true);
		expect(r.disabled).toBe(true);
		r.cleanup();
	}
});
