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
// live (at version 5) or deleted
async function openDialog(entry: { deleted?: boolean, encrypted?: boolean, value: string | null }, exists: boolean) {
	getParam.mockReset();
	postParam.mockReset();
	deleteParam.mockReset();
	if (exists) {
		getParam.mockResolvedValue(current);
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
					type="application/x-list"
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
