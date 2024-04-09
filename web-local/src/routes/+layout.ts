import { dev } from "$app/environment";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

// When testing, we need to use the relative path to the server
const host = dev ? "http://localhost:9009" : "";

export const ssr = false;

export function load() {
  runtime.set({
    host,
    instanceId: "default",
  });
  return {
    host,
    instanceId: "default",
  };
}
