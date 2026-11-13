// Centralized runtime configuration read from Vite env vars.
// VITE_DATA_SOURCE selects the data layer (mock JSON vs the Go API).

export type DataSource = 'mock' | 'api';

export const DATA_SOURCE: DataSource = import.meta.env.VITE_DATA_SOURCE === 'api' ? 'api' : 'mock';

export const API_URL: string = import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api';

export const APP_NAME = 'WeBond';

/** Non-secret marker (never a token): "this browser signed in recently", used to
 *  decide whether to try restoring a session after a page reload. */
export const SESSION_HINT_KEY = 'webond.hasSession';

/** Show the demo account hint on the login page. On by default (this is a demo
 *  app); production builds set VITE_SHOW_DEMO_HINT=false. */
export const SHOW_DEMO_HINT: boolean = import.meta.env.VITE_SHOW_DEMO_HINT !== 'false';
