import type { Query } from "@tanstack/svelte-query";

export function isRuntimeQuery(query: Query): boolean {
  const [apiPath] = query.queryKey as string[];
  return apiPath.startsWith("/v1/instances/");
}
