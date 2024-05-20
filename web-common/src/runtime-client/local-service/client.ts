import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { LocalService } from "../../proto/gen/rill/local/v1/api_connect";

const transport = createConnectTransport({
  baseUrl: "http://localhost:9009",
});

export const localServiceClient = createPromiseClient(LocalService, transport);
