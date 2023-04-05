import * as React from 'react';

const ErrorContext = React.createContext((error: unknown) => { return });
export default ErrorContext;
