import type {
  QueryKey,
  UseQueryStoreResult,
  UseQueryOptions,
  QueryFunction,
} from "@sveltestack/svelte-query";
import { useQuery } from "@sveltestack/svelte-query";

export class QueriesRepository {
  private queriesMap = new Map<string, UseQueryStoreResult>();

  public useQuery<T, E>(
    queryKey: QueryKey,
    queryFn: QueryFunction<T>,
    queryOptions: UseQueryOptions<T, E>
  ): UseQueryStoreResult {
    const key = this.convertQueryKeyToString(queryKey);
    let query: UseQueryStoreResult;
    if (!this.queriesMap.has(key)) {
      query = useQuery(queryKey, queryFn, queryOptions);
      this.queriesMap.set(key, query);
    } else {
      query = this.queriesMap.get(key);
      query.setOptions(queryKey, queryFn, queryOptions);
    }
    return query;
  }

  private convertQueryKeyToString(queryKey: QueryKey): string {
    return (queryKey as Array<string>).join("__");
  }
}

export const queriesRepository = new QueriesRepository();
