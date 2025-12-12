import { page } from "$app/stores";
import { derived, type Readable } from "svelte/store";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import type { V1ResourceName } from "@rilldata/web-common/runtime-client";
import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";

// TODO: merge with appScreen?
export type ActiveFileArtifact = {
  filePath: string;
  resource: V1ResourceName | undefined;
};
export function getActiveFileArtifactStore(): Readable<ActiveFileArtifact> {
  const activeFilePathStore = derived(
    page,
    (pageState) => pageState.params?.file ?? "",
  );

  return derived(activeFilePathStore, (filePath, set) => {
    if (!filePath) set({ filePath, resource: undefined });

    const fileArtifact = fileArtifacts.getFileArtifact(
      addLeadingSlash(filePath),
    );
    return fileArtifact.resourceName.subscribe((resource) => {
      set({
        filePath,
        resource,
      } satisfies ActiveFileArtifact);
    });
  });
}
