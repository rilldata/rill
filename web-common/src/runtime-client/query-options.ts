import type { FetchQueryOptions } from "@tanstack/svelte-query";
import type {
  RuntimeServiceGetResourceParams,
  V1GetResourceResponse,
} from "./gen/index.schemas";
import type { RuntimeServiceGetResourceQueryResult } from "./gen/runtime-service/runtime-service";
import { getRuntimeServiceGetResourceQueryKey } from "./gen/runtime-service/runtime-service";

interface RuntimeInfo {
  host: string;
  instanceId: string;
  jwt?: { token: string } | undefined;
}

export function getRuntimeServiceGetResourceQueryOptions(
  runtime: RuntimeInfo,
  params: RuntimeServiceGetResourceParams,
) {
  return <FetchQueryOptions<RuntimeServiceGetResourceQueryResult>>{
    queryKey: getRuntimeServiceGetResourceQueryKey(runtime.instanceId, params),
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      for (const [key, value] of Object.entries(params)) {
        if (value !== undefined) searchParams.set(key, String(value));
      }
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
