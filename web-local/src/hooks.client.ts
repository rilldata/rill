import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { setLocalServiceHost } from "@rilldata/web-common/runtime-client/local-service";
import { LOCAL_HOST, LOCAL_INSTANCE_ID } from "./lib/local-runtime-config";

// BRIDGE (temporary): keep global store initialized for unmigrated Orval consumers
runtime.set({ host: LOCAL_HOST, instanceId: LOCAL_INSTANCE_ID });

// Initialize LocalService client with the runtime host
setLocalServiceHost(LOCAL_HOST);
