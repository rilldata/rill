import type { FileIO } from "@rilldata/web-common/features/entity-management/file-io.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import type { QueryFunction } from "@tanstack/svelte-query";
import {
  adminServiceGetPersonalFile,
  adminServicePutPersonalFile,
  getAdminServiceGetPersonalFileQueryKey,
} from "@rilldata/web-admin/client";
import { splitFolderFileNameAndExtension } from "@rilldata/web-common/features/entity-management/file-path-utils.ts";

export class VirtualFileIo implements FileIO {
  public constructor(
    private readonly org: string,
    private readonly project: string,
    private readonly userId: string,
    private readonly onWrite?: (
      name: string,
      kind?: string,
    ) => void | Promise<void>,
  ) {}

  async read(path: string, invalidate = false): Promise<string | undefined> {
    const name = this.getNameFromPath(path);
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
    const name = this.getNameFromPath(path);
    // Optimistically update the query
    queryClient.setQueryData(
      getAdminServiceGetPersonalFileQueryKey(this.org, this.project, name),
      { yaml },
    );

    await adminServicePutPersonalFile(this.org, this.project, name, {
      yaml,
      kind,
    });
    this.onWrite?.(name, kind);
  }

  private getNameFromPath(path: string) {
    const [, name] = splitFolderFileNameAndExtension(path);
    return name.replace("_" + this.userId, "");
  }
}
