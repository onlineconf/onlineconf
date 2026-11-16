import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { act, Simulate } from 'react-dom/test-utils';
import { ThemeProvider } from '@mui/material/styles';

import '../i18n';
import createTheme from '../theme';
import * as API from '../api';
import RollbackDialog from './RollbackDialog';

jest.mock('../api');
const getParam = API.getParam as jest.Mock;
const postParam = API.postParam as jest.Mock;
const deleteParam = API.deleteParam as jest.Mock;

const current: API.IParam = {
	path: '/p', name: 'p', mime: 'application/x-list', data: 'cur', summary: '', description: '', version: 5, mtime: '',
	rw: true, notification: 'none', access_modified: false, notification_modified: false, num_children: 0,
};

// opens the dialog on the log entry 42 of /p; `exists` says whether /p is
// live (at version 5, with `cur` overriding its fields) or deleted
async function openDialog(entry: { deleted?: boolean, encrypted?: boolean, value: string | null }, exists: boolean, cur: Partial<API.IParam> = {}, type: API.ParamType = 'application/x-list') {
	getParam.mockReset();
	postParam.mockReset();
	deleteParam.mockReset();
	if (exists) {
		getParam.mockResolvedValue({ ...current, ...cur });
	} else {
		getParam.mockRejectedValue({ response: { status: 404 } });
	}
	postParam.mockResolvedValue(current);
	deleteParam.mockResolvedValue(undefined);
	const div = document.createElement('div');
	document.body.appendChild(div);
	await act(async () => {
		ReactDOM.render(
			<ThemeProvider theme={createTheme()}>
				<RollbackDialog
					path="/p"
					id={42}
					version={2}
					type={type}
					value={entry.value}
					encrypted={entry.encrypted}
					deleted={entry.deleted === true}
					onLoaded={() => undefined}
					onChange={() => undefined}
					onError={() => undefined}
					onClose={() => undefined}
				/>
			</ThemeProvider>,
			div,
		);
	});
	return () => {
		ReactDOM.unmountComponentAtNode(div);
		div.remove();
	};
}

// types a comment and presses Rollback
async function rollback() {
	const input = document.body.querySelector<HTMLInputElement>('.MuiDialog-root input');
	if (input === null) {
		throw new Error('no comment field');
	}
	await act(async () => {
		input.value = 'c';
		Simulate.change(input);
	});
	const buttons = Array.from(document.body.querySelectorAll<HTMLButtonElement>('.MuiDialog-root button'));
	const button = buttons.find(b => b.textContent === 'Rollback') || buttons[buttons.length - 1];
	await act(async () => {
		Simulate.click(button);
	});
}

it('deletes the current parameter for a deletion entry', async () => {
	const cleanup = await openDialog({ deleted: true, value: null }, true);
	await rollback();
	expect(deleteParam).toHaveBeenCalledWith('/p', { version: 5, comment: 'c' });
	expect(postParam).not.toHaveBeenCalled();
	cleanup();
});

it('restores a live parameter by the log entry, with its version', async () => {
	const cleanup = await openDialog({ value: 'old' }, true);
	await rollback();
	expect(postParam).toHaveBeenCalledWith('/p', { version: 5, restore: 42, comment: 'c' });
	expect(deleteParam).not.toHaveBeenCalled();
	cleanup();
});

it('brings a deleted parameter back by the log entry, a masked entry included', async () => {
	const cleanup = await openDialog({ value: null, encrypted: true }, false);
	await rollback();
	expect(postParam).toHaveBeenCalledWith('/p', { restore: 42, comment: 'c' });
	expect(deleteParam).not.toHaveBeenCalled();
	cleanup();
});

const note = 'The parameter is encrypted: the restored value will be stored encrypted too.';

it('notes the never-downgrade for an eligible entry of an encrypted parameter only', async () => {
	// the parameter's own flag says it is encrypted, not can_reveal
	let cleanup = await openDialog({ value: 'old' }, true, { encrypted: true, can_reveal: false, data: null });
	expect(document.body.textContent).toContain(note);
	cleanup();
	// a symlink version comes back plain: no note
	cleanup = await openDialog({ value: '/q' }, true, { encrypted: true, can_reveal: true, data: null }, 'application/x-symlink');
	expect(document.body.textContent).not.toContain(note);
	cleanup();
	// can_reveal without the flag is not an encrypted parameter
	cleanup = await openDialog({ value: 'old' }, true, { can_reveal: true });
	expect(document.body.textContent).not.toContain(note);
	cleanup();
});

it('edits a revealed secret version of a plain parameter with the switch on and no declassify confirmation', async () => {
	const cleanup = await openDialog({ value: 'old-secret', encrypted: true }, true);
	const pencils = Array.from(document.body.querySelectorAll<HTMLButtonElement>('.MuiDialog-root button')).filter(b => b.querySelector('svg[data-testid="EditIcon"]'));
	expect(pencils).toHaveLength(2);
	await act(async () => {
		Simulate.click(pencils[1]);
	});
	const editor = Array.from(document.body.querySelectorAll<HTMLElement>('.MuiDialog-root')).pop();
	if (!editor) {
		throw new Error('no editor');
	}
	// the version's flag seeds the switch...
	const toggle = editor.querySelector<HTMLInputElement>('input[type="checkbox"]');
	expect(toggle?.checked).toBe(true);
	// ...but the parameter is plain: turning it off declassifies nothing
	await act(async () => {
		toggle!.checked = false;
		Simulate.change(toggle!);
	});
	const label = Array.from(editor.querySelectorAll('label')).find(e => e.textContent?.startsWith('Comment'));
	const comment = label?.htmlFor ? document.getElementById(label.htmlFor) as HTMLInputElement | null : null;
	if (!comment) {
		throw new Error('no comment field in the editor');
	}
	await act(async () => {
		comment.value = 'c';
		Simulate.change(comment);
	});
	const saveButton = Array.from(editor.querySelectorAll<HTMLButtonElement>('button')).find(b => b.textContent === 'Save');
	if (!saveButton) {
		throw new Error('no Save button');
	}
	await act(async () => {
		Simulate.click(saveButton);
	});
	expect(document.body.textContent).not.toContain('Store in plaintext?');
	expect(postParam).toHaveBeenCalledTimes(1);
	const posted = postParam.mock.calls[0][1] as API.ParamModify;
	expect(posted.data).toBe('old-secret');
	expect(posted.encrypted).toBeUndefined();
	cleanup();
});
