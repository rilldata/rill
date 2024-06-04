import { dev } from "$app/environment";
import {
  projectInitialized,
  runtime,
} from "@rilldata/web-common/runtime-client/runtime-store";

/** INITIALIZE RUNTIME STORE **/

// When testing, we need to use the relative path to the server
const HOST = dev ? "http://localhost:9009" : "";
const INSTANCE_ID = "default";

const runtimeInit = {
  host: HOST,
  instanceId: INSTANCE_ID,
};

runtime.set(runtimeInit);

projectInitialized.init().catch(console.error);
