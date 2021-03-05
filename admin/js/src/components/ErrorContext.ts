import * as React from 'react';

const ErrorContext = React.createContext((error: Error) => { return; });
export default ErrorContext;
