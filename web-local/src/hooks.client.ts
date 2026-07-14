import { setLocalServiceHost } from "@rilldata/web-common/runtime-client/local-service";
import { LOCAL_HOST } from "./lib/runtime-client";

// Initialize LocalService client with the runtime host
setLocalServiceHost(LOCAL_HOST);
