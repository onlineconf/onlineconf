import * as React from 'react';
import * as ReactDOM from 'react-dom';
import App from './App';
import './i18n';
import * as ServiceWorker from './registerServiceWorker';

import 'roboto-fontface/css/roboto/roboto-fontface.css';


ReactDOM.render(
	<App />,
	document.getElementById('root') as HTMLElement
);
ServiceWorker.unregister();
