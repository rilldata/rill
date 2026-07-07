import type { FileIO } from "@rilldata/web-common/features/entity-management/file-io.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import type { QueryFunction } from "@tanstack/svelte-query";
import {
  adminServiceGetPersonalFile,
  adminServiceEditPersonalFile,
  getAdminServiceGetPersonalFileQueryKey,
} from "@rilldata/web-admin/client";
import { splitFolderFileNameAndExtension } from "@rilldata/web-common/features/entity-management/file-path-utils.ts";
import { inferResourceKind } from "@rilldata/web-common/features/entity-management/infer-resource-kind.ts";
import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";

type VirtualFileEvents = {
  write: { name: string; kind: string };
};

export class VirtualFileIo implements FileIO {
  private events = new EventEmitter<VirtualFileEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;

  public constructor(
    private readonly org: string,
    private readonly project: string,
  ) {}

  updateClient() {}

  async read(path: string, invalidate = false): Promise<string | undefined> {
    const [, name] = splitFolderFileNameAndExtension(path);
    const queryKey = getAdminServiceGetPersonalFileQueryKey(
      this.org,
      this.project,
      name,
    );

    if (invalidate) await queryClient.invalidateQueries({ queryKey });

    const queryFn: QueryFunction<
      Awaited<ReturnType<typeof adminServiceGetPersonalFile>>
    > = () => adminServiceGetPersonalFile(this.org, this.project, name);

    try {
      const response = await queryClient.fetchQuery({
        queryKey,
        queryFn,
        staleTime: Infinity,
      });
      return response.yaml;
    } catch (e) {
      console.log("FETCH ERROR", e);
      return undefined;
    }
  }

  async write(path: string, yaml: string, kind?: string): Promise<void> {
    if (!kind) {
      kind = inferResourceKind(path, yaml) as string | undefined;
      if (!kind) throw new Error("Could not infer resource kind");
    }

    const [, name] = splitFolderFileNameAndExtension(path);
    // Optimistically update the query
    queryClient.setQueryData(
      getAdminServiceGetPersonalFileQueryKey(this.org, this.project, name),
      { yaml, path },
    );

    await adminServiceEditPersonalFile(this.org, this.project, name, {
      yaml,
      kind,
    });
    this.events.emit("write", { name, kind });
  }
}
