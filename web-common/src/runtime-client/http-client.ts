import type {
  FetchWrapperOptions,
  HTTPError,
} from "@rilldata/web-common/runtime-client/fetchWrapper";
import { get } from "svelte/store";
import {
  RUNTIME_ACCESS_TOKEN_DEFAULT_TTL,
  JWT_EXPIRY_WARNING_WINDOW,
  CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL,
} from "./constants";
import { HttpRequestQueue } from "./http-request-queue/HttpRequestQueue";
import { type JWT, runtime } from "./runtime-store";

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

  const interceptedConfig = { ...config };
  // Set host
  const host = get(runtime).host;
  if (host) interceptedConfig.baseUrl = host;

  // Set JWT
  let jwt = get(runtime).jwt;
  if (jwt && jwt.token && !interceptedConfig.headers?.["Authorization"]) {
    jwt = await maybeWaitForFreshJWT(jwt);
    interceptedConfig.headers = {
      ...interceptedConfig.headers,
      Authorization: `Bearer ${jwt.token}`,
    };
  }

  return (await httpRequestQueue.add(interceptedConfig)) as Promise<T>;
};

/**
 * If the JWT has expired, or is close to expiring, wait for a fresh one.
 */
async function maybeWaitForFreshJWT(jwt: JWT): Promise<JWT> {
  // Embeds communicate directly with the runtime and have no admin server connection to refresh tokens.
  // The backend issues embed JWTs with a 24h TTL, so skip the client-side expiry check.
  if (jwt.authContext === "embed") return jwt;

  // This is the approximate time at which the JWT will expire. We could parse the JWT to get the exact
  // expiration time, but it's better to treat tokens as opaque.
  let jwtExpiresAt = jwt.receivedAt + RUNTIME_ACCESS_TOKEN_DEFAULT_TTL;

  while (Date.now() + JWT_EXPIRY_WARNING_WINDOW > jwtExpiresAt) {
    // Note: Rather than waiting, it could be even better to immediately fetch a new token here. Anyways, in
    // practice, a request for new token is already in flight. So to start simpler, we just wait.
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
