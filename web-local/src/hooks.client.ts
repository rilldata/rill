import { setLocalServiceHost } from "@rilldata/web-common/runtime-client/local-service";
import { LOCAL_HOST } from "./lib/local-runtime-config";

// Initialize LocalService client with the runtime host
setLocalServiceHost(LOCAL_HOST);
