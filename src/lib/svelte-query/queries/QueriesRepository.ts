import type {
  QueryFunction,
  QueryKey,
  UseQueryOptions,
  UseQueryStoreResult,
} from "@sveltestack/svelte-query";
import { useQuery } from "@sveltestack/svelte-query";

export class QueriesRepository {
  private queriesMap = new Map<string, UseQueryStoreResult>();

  public useQuery<T, E>(
    queryKey: QueryKey,
    queryFn: QueryFunction<T>,
    queryOptions: UseQueryOptions<T, E>
  ): UseQueryStoreResult {
    const key = convertQueryKeyToString(queryKey);
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
}

// Copied from: https://github.com/SvelteStack/svelte-query/blob/main/src/queryCore/core/utils.ts
function convertQueryKeyToString(value: any): string {
  return JSON.stringify(value, (_, val) =>
    isPlainObject(val)
      ? Object.keys(val)
          .sort()
          .reduce((result, key) => {
            result[key] = val[key];
            return result;
          }, {} as any)
      : val
  );
}

// Copied from: https://github.com/jonschlinkert/is-plain-object
function isPlainObject(o: any): o is Object {
  if (!hasObjectPrototype(o)) {
    return false;
  }

  // If has modified constructor
  const ctor = o.constructor;
  if (typeof ctor === "undefined") {
    return true;
  }

  // If has modified prototype
  const prot = ctor.prototype;
  if (!hasObjectPrototype(prot)) {
    return false;
  }

  // If constructor does not have an Object-specific method
  // if (!prot.hasOwnProperty('isPrototypeOf')) {
  //   return false
  // }

  // Most likely a plain Object
  return true;
}

function hasObjectPrototype(o: any): boolean {
  return Object.prototype.toString.call(o) === "[object Object]";
}

export const queriesRepository = new QueriesRepository();
