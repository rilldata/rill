import {
  getRuntimeClient,
  type RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";

/** For load functions and tests. In components, use {@link useRuntimeClient} instead. */
export function getCloudRuntimeClient(runtime: {
  host: string;
  instanceId: string;
  jwt?: { token: string; authContext?: string };
}): RuntimeClient {
  return getRuntimeClient({
    host: runtime.host,
    instanceId: runtime.instanceId,
    jwt: runtime.jwt?.token,
    authContext: runtime.jwt?.authContext,
  });
}
