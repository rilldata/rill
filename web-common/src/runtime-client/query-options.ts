import type { FetchQueryOptions } from "@tanstack/svelte-query";
import type { V1GetResourceResponse } from "./gen/index.schemas";
import { getRuntimeServiceGetResourceQueryKey } from "./v2/gen/runtime-service";

interface RuntimeInfo {
  host: string;
  instanceId: string;
  jwt?: { token: string } | undefined;
}

interface GetResourceParams {
  name: { name: string; kind: string };
}

export function getRuntimeServiceGetResourceQueryOptions(
  runtime: RuntimeInfo,
  params: GetResourceParams,
) {
  return <FetchQueryOptions<V1GetResourceResponse>>{
    queryKey: getRuntimeServiceGetResourceQueryKey(runtime.instanceId, params),
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      // Flatten nested params for the REST API query string
      searchParams.set("name.name", params.name.name);
      searchParams.set("name.kind", params.name.kind);
      const url = `${runtime.host}/v1/instances/${runtime.instanceId}/resource?${searchParams}`;
      const headers: Record<string, string> = {};
      if (runtime.jwt) {
        headers["Authorization"] = `Bearer ${runtime.jwt.token}`;
      }
      const resp = await fetch(url, { headers });
      if (!resp.ok) {
        const data = await resp.json().catch(() => ({}));
        throw { response: { status: resp.status, data } };
      }
      return (await resp.json()) as V1GetResourceResponse;
    },
    staleTime: Infinity,
  };
}
