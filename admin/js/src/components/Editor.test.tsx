import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { act, Simulate } from 'react-dom/test-utils';
import { ThemeProvider } from '@mui/material/styles';

import '../i18n';
import createTheme from '../theme';
import * as API from '../api';
import { ParamType, EncryptionConfig } from '../api';
import { EncryptionContext } from './UIConfig';
import Editor from './Editor';

jest.mock('../api');
const postParam = API.postParam as jest.Mock;

const on: EncryptionConfig = { enabled: true };
const off: EncryptionConfig = { enabled: false };

interface Opened {
	create?: boolean;
	encrypted?: boolean;
	initialEncrypted?: boolean;
}

// opens the editor on /p, a list value "a" at version 1 (or a create under /p)
function openEditor(encryption: EncryptionConfig, opened: Opened) {
	postParam.mockReset();
	postParam.mockResolvedValue({});
	const div = document.createElement('div');
	document.body.appendChild(div);
	act(() => {
		ReactDOM.render(
			<ThemeProvider theme={createTheme()}>
				<EncryptionContext.Provider value={encryption}>
					<Editor
						path="/p"
						version={opened.create ? undefined : 1}
						create={opened.create}
						type="application/x-list"
						value="a"
						encrypted={opened.encrypted}
						initialEncrypted={opened.initialEncrypted}
						revealed={opened.encrypted || opened.initialEncrypted}
						summary=""
						description=""
						onClose={() => undefined}
						onChange={() => undefined}
						onError={() => undefined}
					/>
				</EncryptionContext.Provider>
			</ThemeProvider>,
			div,
		);
	});
	return () => {
		ReactDOM.unmountComponentAtNode(div);
		div.remove();
	};
}

function dialog() {
	const d = document.body.querySelector<HTMLElement>('.MuiDialog-root');
	if (d === null) {
		throw new Error('no editor dialog');
	}
	return d;
}

function inputByLabel(label: string) {
	const l = Array.from(dialog().querySelectorAll('label')).find(e => e.textContent?.startsWith(label));
	const input = l?.htmlFor ? document.getElementById(l.htmlFor) as HTMLInputElement | null : null;
	if (!input) {
		throw new Error(`no field labeled ${label}`);
	}
	return input;
}

function buttonByText(text: string) {
	const b = Array.from(document.body.querySelectorAll<HTMLButtonElement>('.MuiDialog-root button')).find(e => e.textContent === text);
	if (!b) {
		throw new Error(`no button ${text}`);
	}
	return b;
}

function setType(type: ParamType) {
	// the hidden native input of the type select takes a value as autofill would
	const input = dialog().querySelector<HTMLInputElement>('.MuiSelect-nativeInput');
	if (input === null) {
		throw new Error('no type select');
	}
	act(() => {
		Simulate.change(input, { target: { value: type } as unknown as EventTarget });
	});
}

function toggleEncrypted() {
	const input = dialog().querySelector<HTMLInputElement>('input[type="checkbox"]');
	if (input === null) {
		throw new Error('no Encrypted switch');
	}
	act(() => {
		input.checked = !input.checked;
		Simulate.change(input);
	});
}

function comment() {
	const input = inputByLabel('Comment');
	act(() => {
		input.value = 'c';
		Simulate.change(input);
	});
}

async function save() {
	await act(async () => {
		Simulate.click(buttonByText('Save'));
	});
}

async function confirmDeclassify() {
	await act(async () => {
		Simulate.click(buttonByText('Store in plaintext'));
	});
}

function posted() {
	expect(postParam).toHaveBeenCalledTimes(1);
	return postParam.mock.calls[0][1] as API.ParamModify;
}

afterEach(() => {
	document.body.innerHTML = '';
});

it('states the flag as switched for an eligible type while the feature is on', async () => {
	const cleanup = openEditor(on, {});
	comment();
	await save();
	expect(posted().encrypted).toBe(false);
	cleanup();
});

it('states the flag on when switched on, on a create too', async () => {
	const cleanup = openEditor(on, { create: true });
	act(() => {
		inputByLabel('Path').value = 'n';
		Simulate.change(inputByLabel('Path'));
	});
	toggleEncrypted();
	comment();
	await save();
	expect(posted().encrypted).toBe(true);
	expect(postParam.mock.calls[0][0]).toBe('/p/n');
	cleanup();
});

it('omits the flag for a plain parameter while the feature is off', async () => {
	const cleanup = openEditor(off, {});
	comment();
	await save();
	expect(posted().encrypted).toBeUndefined();
	cleanup();
});

it('keeps the flag on an encrypted parameter re-saved, whatever the switch', async () => {
	const cleanup = openEditor(off, { encrypted: true });
	comment();
	await save();
	expect(posted().encrypted).toBe(true);
	cleanup();
});

it('declassifies with a confirmation when the flag is switched off', async () => {
	const cleanup = openEditor(on, { encrypted: true });
	toggleEncrypted();
	comment();
	await save();
	expect(postParam).not.toHaveBeenCalled();
	await confirmDeclassify();
	expect(posted().encrypted).toBe(false);
	cleanup();
});

it('declassifies with a confirmation when an ineligible type replaces a secret', async () => {
	for (const type of ['application/x-case', 'application/x-symlink', 'application/x-template'] as ParamType[]) {
		const cleanup = openEditor(on, { encrypted: true });
		setType(type);
		comment();
		await save();
		expect(postParam).not.toHaveBeenCalled();
		await confirmDeclassify();
		const p = posted();
		expect(p.mime).toBe(type);
		expect(p.encrypted).toBe(false);
		cleanup();
	}
});

it('omits the flag when a plain parameter becomes a case', async () => {
	const cleanup = openEditor(on, {});
	setType('application/x-case');
	comment();
	await save();
	const p = posted();
	expect(p.mime).toBe('application/x-case');
	expect(p.encrypted).toBeUndefined();
	cleanup();
});

it('states the flag as the switch says when a secret version seeds it on a plain parameter, without a confirmation', async () => {
	// the feature off and the parameter plain: the switch, seeded on, is
	// shown and its state is posted — an omitted field would store the
	// revealed secret in plaintext
	let cleanup = openEditor(off, { encrypted: false, initialEncrypted: true });
	expect(dialog().querySelector<HTMLInputElement>('input[type="checkbox"]')?.checked).toBe(true);
	comment();
	await save();
	expect(posted().encrypted).toBe(true);
	cleanup();
	// switched off: plaintext, and nothing to confirm
	cleanup = openEditor(off, { encrypted: false, initialEncrypted: true });
	toggleEncrypted();
	comment();
	await save();
	expect(posted().encrypted).toBeUndefined();
	expect(document.body.textContent).not.toContain('Store in plaintext?');
	cleanup();
});
