import { page } from "$app/stores";
import { derived, type Readable } from "svelte/store";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import type { V1ResourceName } from "@rilldata/web-common/runtime-client";
import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";

// TODO: merge with appScreen?
export function getActiveResourceStore(): Readable<V1ResourceName | undefined> {
  const activeFilePathStore = derived(
    page,
    (pageState) => pageState.params?.file ?? "",
  );

  return derived(activeFilePathStore, (filePath, set) => {
    if (!filePath) set(undefined);

    const fileArtifact = fileArtifacts.getFileArtifact(
      addLeadingSlash(filePath),
    );
    return fileArtifact.resourceName.subscribe(set);
  });
}
