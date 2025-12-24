import { dev } from "$app/environment";
import httpClient from "@rilldata/web-common/runtime-client/http-client";

/** INITIALIZE RUNTIME STORE **/
// When testing, we need to use the relative path to the server
const HOST = dev ? "http://localhost:9009" : "";
const INSTANCE_ID = "default";

void httpClient.updateQuerySettings({
  instanceId: INSTANCE_ID,
  host: HOST,
  token: undefined,
  authContext: "user",
});
