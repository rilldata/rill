// The TTL is actually set in the Admin server – we just use the value for some frontend logic
export const RUNTIME_ACCESS_TOKEN_DEFAULT_TTL = 30 * 60 * 1000; // 30 minutes

// Extra buffer to ensure the JWT hasn't expired by the time it reaches the server
export const JWT_EXPIRY_WARNING_WINDOW = 2 * 1000;

// Interval to recheck JWT freshness while waiting for a refresh
export const CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL = 50;
