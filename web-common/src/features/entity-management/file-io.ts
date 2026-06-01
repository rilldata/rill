import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceGetFile,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { QueryFunction } from "@tanstack/svelte-query";

export interface FileIO {
  read(path: string, invalidate?: boolean): Promise<string | undefined>;
  write(path: string, blob: string, kind?: string): Promise<void>;
}

export class RuntimeFileIO implements FileIO {
  private client!: RuntimeClient;

  updateClient(client: RuntimeClient) {
    this.client = client;
  }

  async read(path: string, invalidate = false): Promise<string | undefined> {
    if (!this.client) return;
    const queryParams = { path };
    const queryKey = getRuntimeServiceGetFileQueryKey(
      this.client.instanceId,
      queryParams,
    );

    if (invalidate) await queryClient.invalidateQueries({ queryKey });

    const queryFn: QueryFunction<
      Awaited<ReturnType<typeof runtimeServiceGetFile>>
    > = ({ signal }) =>
      runtimeServiceGetFile(this.client, queryParams, { signal });

    try {
      const response = await queryClient.fetchQuery({
        queryKey,
        queryFn,
        staleTime: Infinity,
      });
      return response.blob;
    } catch (e) {
      console.log("FETCH ERROR", e);
      return undefined;
    }
  }

  async write(path: string, blob: string): Promise<void> {
    if (!this.client) return;

    // Optimistically update the query
    queryClient.setQueryData(
      getRuntimeServiceGetFileQueryKey(this.client.instanceId, { path }),
      { blob },
    );

    await runtimeServicePutFile(this.client, { path, blob });
  }
}
