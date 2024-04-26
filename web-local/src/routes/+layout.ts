import { dev } from "$app/environment";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

export const ssr = false;

// When testing, we need to use the relative path to the server
const HOST = dev ? "http://localhost:9009" : "";
const INSTANCE_ID = "default";

const runtimeInit = {
  host: HOST,
  instanceId: INSTANCE_ID,
};

export function load() {
  runtime.set(runtimeInit);
  return runtimeInit;
}
