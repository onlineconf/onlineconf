import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { act } from 'react-dom/test-utils';

import * as API from '../../api';
import { useReveal } from './encrypted';

jest.mock('../../api');
const getParam = API.getParam as jest.Mock;

const param = (version: number): API.IParam => ({
	path: '/p', name: 'p', mime: 'text/plain', data: 's' + version, summary: '', description: '', version, mtime: '',
	rw: true, notification: 'none', access_modified: false, notification_modified: false, num_children: 0,
});

let latest: ReturnType<typeof useReveal> | undefined;
function Harness(props: { version: number }) {
	latest = useReveal('/p', props.version, () => undefined);
	return null;
}

it('ignores a reveal that resolves after the version changed, and keeps one for the current version', async () => {
	let resolve: (p: API.IParam) => void = () => undefined;
	getParam.mockReset();
	getParam.mockReturnValue(new Promise<API.IParam>(r => { resolve = r; }));
	const div = document.createElement('div');
	act(() => {
		ReactDOM.render(<Harness version={1}/>, div);
	});
	await act(async () => {
		latest?.reveal();
	});
	act(() => {
		ReactDOM.render(<Harness version={2}/>, div);
	});
	await act(async () => {
		resolve(param(1));
	});
	expect(latest?.revealed).toBeUndefined();
	getParam.mockResolvedValue(param(2));
	await act(async () => {
		await latest?.reveal();
	});
	expect(latest?.revealed?.data).toBe('s2');
	ReactDOM.unmountComponentAtNode(div);
});
