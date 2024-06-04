import { dev } from "$app/environment";
import { isProjectInitialized } from "@rilldata/web-common/features/welcome/is-project-initialized";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
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

/** CHECK IF RILL.YAML EXISTS **/

isProjectInitialized(queryClient, runtimeInit.instanceId)
  .then((initialized) => projectInitialized.set(!!initialized))
  .catch(console.error);
