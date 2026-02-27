import {
  RuntimeClient,
  type AuthContext,
} from "@rilldata/web-common/runtime-client/v2";

/**
 * Construct a RuntimeClient from the layout `runtime` data
 * that web-admin passes through `await parent()`.
 */
export function createRuntimeClientFromLayout(runtime: {
  host: string;
  instanceId: string;
  jwt?: { token: string; authContext?: string };
}): RuntimeClient {
  return new RuntimeClient({
    host: runtime.host,
    instanceId: runtime.instanceId,
    jwt: runtime.jwt?.token,
    authContext: runtime.jwt?.authContext as AuthContext,
  });
}
