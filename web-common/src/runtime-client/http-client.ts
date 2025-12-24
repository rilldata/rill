import type {
  FetchWrapperOptions,
  HTTPError,
} from "@rilldata/web-common/runtime-client/fetchWrapper";
import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "./constants";
import { HttpRequestQueue } from "./http-request-queue/HttpRequestQueue";
import { type AuthContext, type JWT } from "./runtime-store";
import { invalidateRuntimeQueries } from "./invalidation";
import { queryClient } from "../lib/svelte-query/globalQueryClient";
import { featureFlags } from "../features/feature-flags";

/**
 * Runtime base URL
 *  Local
 *    In dev & prod: http://localhost:9009
 *  Cloud
 *    In dev: http://localhost:9009
 *    In prod: https://{region}.runtime.rilldata.com
 */

export const httpRequestQueue = new HttpRequestQueue();

export class HTTPClient {}

export const createHttpClient = (
  initialHost: string = "",
  initialJwt?: JWT,
) => {
  let _host = initialHost;
  let _jwt = initialJwt;
  let _instanceId = "";

  const client = async <T>(config: FetchWrapperOptions): Promise<T> => {
    // Naive request interceptors

    // Set host
    const interceptedConfig = { ...config, baseUrl: _host };

    // Set JWT
    if (_jwt && _jwt.token) {
      _jwt = await maybeWaitForFreshJWT(_jwt);
      interceptedConfig.headers = {
        ...interceptedConfig.headers,
        Authorization: `Bearer ${_jwt.token}`,
      };
    }

    return (await httpRequestQueue.add(interceptedConfig)) as Promise<T>;
  };

  client.setDefaultsForMocks = () => {
    _host = "http://localhost";
    _instanceId = "default";
  };

  client.updateQuerySettings = async ({
    host,
    token,
    authContext,
    instanceId,
  }: {
    host: string;
    token: string | undefined;
    authContext: AuthContext;
    instanceId: string;
  }) => {
    let invalidate = false;
    //  Don't update the store if the values have not changed
    // (especially, don't update the JWT `receivedAt`)
    if (
      host === _host &&
      token === _jwt?.token &&
      authContext === _jwt?.authContext &&
      instanceId === _instanceId
    ) {
      return;
    }

    // Mark the runtime queries for invalidation if the auth context has changed
    // E.g. when switching from a normal user to a mocked user
    const authContextChanged =
      !!_jwt?.authContext && authContext !== _jwt.authContext;
    if (authContextChanged) invalidate = true;

    _instanceId = instanceId;

    _host = host;
    _jwt =
      token && authContext
        ? { token: token, receivedAt: Date.now(), authContext }
        : undefined;

    void featureFlags.setInstanceId(instanceId);

    if (invalidate) await invalidateRuntimeQueries(queryClient, _instanceId);
  };

  client.getHost = () => _host;
  client.getJwt = () => _jwt;
  client.getInstanceId = () => _instanceId;

  client.updateJWT = async (
    token: string | undefined,
    authContext: AuthContext | undefined,
  ) => {
    await client.updateQuerySettings({
      host: _host,
      token: token,
      authContext: authContext || "user",
      instanceId: _instanceId,
    });
  };

  return client;
};

const httpClient = createHttpClient();

const JWT_EXPIRY_WARNING_WINDOW = 2 * 1000; // Extra time to ensure that the JWT is not expired when it ultimately reaches the server
const CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL = 50; // Interval to recheck JWT freshness in milliseconds

/**
 * If the JWT has expired, or is close to expiring, wait for a fresh one.
 */
async function maybeWaitForFreshJWT(jwt: JWT): Promise<JWT> {
  // This is the approximate time at which the JWT will expire. We could parse the JWT to get the exact
  // expiration time, but it's better to treat tokens as opaque.
  let jwtExpiresAt = jwt.receivedAt + RUNTIME_ACCESS_TOKEN_DEFAULT_TTL;

  while (Date.now() + JWT_EXPIRY_WARNING_WINDOW > jwtExpiresAt) {
    // Note: Rather than waiting, it could be even better to immediately fetch a new token here. Anyways, in
    // practice, a request for new token is already in flight. So to start simpler, we just wait.
    await new Promise((resolve) =>
      setTimeout(resolve, CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL),
    );
    // jwt = get(runtime).jwt as JWT;
    jwtExpiresAt = jwt.receivedAt + RUNTIME_ACCESS_TOKEN_DEFAULT_TTL;
  }

  return jwt;
}

export default httpClient;

// This overrides Orval's generated error type. (Orval expects this to be a generic.)
// eslint-disable-next-line @typescript-eslint/no-unused-vars
export type ErrorType<Error> = HTTPError;
