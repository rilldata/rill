import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "@rilldata/web-admin/features/projects/selectors";
import type {
  FetchWrapperOptions,
  HTTPError,
} from "@rilldata/web-common/runtime-client/fetchWrapper";
import { get } from "svelte/store";
import { HttpRequestQueue } from "./http-request-queue/HttpRequestQueue";
import { JWT, runtime } from "./runtime-store";

/**
 * Runtime base URL
 *  Local
 *    In dev & prod: http://localhost:9009
 *  Cloud
 *    In dev: http://localhost:9009
 *    In prod: https://{region}.runtime.rilldata.com
 */

export const httpRequestQueue = new HttpRequestQueue();

export const httpClient = async <T>(
  config: FetchWrapperOptions,
): Promise<T> => {
  // Naive request interceptors

  // Set host
  const host = get(runtime).host;
  const interceptedConfig = { ...config, baseUrl: host };

  // Set JWT
  let jwt = get(runtime).jwt;
  if (jwt) {
    jwt = await maybeWaitForFreshJWT(jwt);
    interceptedConfig.headers = {
      ...interceptedConfig.headers,
      Authorization: `Bearer ${jwt.token}`,
    };
  }

  return (await httpRequestQueue.add(interceptedConfig)) as Promise<T>;
};

const JWT_EXPIRY_WARNING_WINDOW = 2 * 1000; // Extra time to ensure that the JWT is not expired when it ultimately reaches the server
const CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL = 50; // Interval to recheck JWT freshness in milliseconds

/**
 * When a user returns to the app after an extended period, their JWT may have expired. At the moment the user returns,
 * TanStack Query will automatically refetch stale queries.
 *
 * Without intervention, simultaneously:
 * - the `GetProject` query would be sent to the Admin server to get a fresh JWT
 * - dashboard queries would be sent to the Runtime server, but with the stale JWT! This would result in 401 errors.
 *
 * So, this function waits for a fresh JWT to be set in the runtime store before sending requests to the runtime.
 */
async function maybeWaitForFreshJWT(jwt: JWT): Promise<JWT> {
  // This is the approximate time at which the JWT will expire.
  // We could parse the JWT to get the exact expiration time, but it's better to treat the JWT as opaque.
  let jwtExpiresAt = jwt.receivedAt + RUNTIME_ACCESS_TOKEN_DEFAULT_TTL;

  // If the JWT has expired, or is close to expiring, wait for a fresh one.
  // Note: It could be even better to immediately fetch a new token here. However, in practice, the request
  // for new token is already in flight (see comment for `maybeWaitForFreshJWT`). So, to keep the code simpler, we just wait.
  while (Date.now() + JWT_EXPIRY_WARNING_WINDOW > jwtExpiresAt) {
    await new Promise((resolve) =>
      setTimeout(resolve, CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL),
    );
    jwt = get(runtime).jwt as JWT;
    jwtExpiresAt = jwt.receivedAt + RUNTIME_ACCESS_TOKEN_DEFAULT_TTL;
  }

  return jwt;
}

export default httpClient;

// This overrides Orval's generated error type. (Orval expects this to be a generic.)
// eslint-disable-next-line @typescript-eslint/no-unused-vars
export type ErrorType<Error> = HTTPError;
