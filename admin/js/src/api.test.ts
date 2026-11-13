import axios from 'axios';

import { getParam, getParamLog } from './api';

jest.mock('axios');
const get = axios.get as jest.Mock;

beforeEach(() => {
	get.mockReset();
	get.mockResolvedValue({ data: {} });
});

// the server accepts reveal=1 only, for the parameter and for its history
it('asks for a reveal as reveal=1', async () => {
	await getParam('/a/b', undefined, 'none', {}, true);
	expect(get).toHaveBeenLastCalledWith('/config/a/b', expect.objectContaining({ params: { symlink: undefined, depth: 'none', reveal: 1 } }));
	await getParamLog('/a/b', undefined, {}, true);
	expect(get).toHaveBeenLastCalledWith('/log/a/b', expect.objectContaining({ params: expect.objectContaining({ reveal: 1 }) }));
});

it('sends no reveal parameter for a plain read', async () => {
	await getParam('/a/b');
	expect(get).toHaveBeenLastCalledWith('/config/a/b', expect.objectContaining({ params: { symlink: undefined, depth: undefined } }));
	await getParamLog('/a/b');
	const params = get.mock.calls[get.mock.calls.length - 1][1].params;
	expect(params.reveal).toBeUndefined();
});
