// import type { FetchQueryOptions } from "@tanstack/svelte-query";
// import type {
//   RuntimeServiceGetResourceParams,
//   V1GetResourceResponse,
// } from "./gen/index.schemas";
// import type { RuntimeServiceGetResourceQueryResult } from "./gen/runtime-service/runtime-service";
// import { getRuntimeServiceGetResourceQueryKey } from "./gen/runtime-service/runtime-service";
// import httpClient from "@rilldata/web-common/runtime-client/http-client.ts";
// import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";

// export function getRuntimeServiceGetResourceQueryOptions(
//   runtime: Runtime,
//   params: RuntimeServiceGetResourceParams,
// ) {
//   return <FetchQueryOptions<RuntimeServiceGetResourceQueryResult>>{
//     queryKey: getRuntimeServiceGetResourceQueryKey(runtime.instanceId, params),
//     queryFn: () =>
//       httpClient<V1GetResourceResponse>({
//         url: `/v1/instances/${runtime.instanceId}/resource`,
//         method: "GET",
//         params,
//         baseUrl: runtime.host,
//         headers: runtime.jwt
//           ? {
//               Authorization: `Bearer ${runtime.jwt?.token}`,
//             }
//           : undefined,
//       }),
//     staleTime: Infinity,
//   };
// }
