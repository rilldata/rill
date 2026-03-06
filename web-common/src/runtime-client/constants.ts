// The TTL is actually set in the Admin server – we just use the value for some frontend logic
export const RUNTIME_ACCESS_TOKEN_DEFAULT_TTL = 30 * 60 * 1000; // 30 minutes

// Buffer to avoid sending a request whose JWT expires in transit.
// Only needs to cover network round-trip; the 15-min refetch cycle
// handles proactive refresh. Already-expired JWTs are always blocked.
export const JWT_EXPIRY_WARNING_WINDOW = 1000; // 1 second

// Interval to recheck JWT freshness while waiting for a refresh
export const CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL = 50;
