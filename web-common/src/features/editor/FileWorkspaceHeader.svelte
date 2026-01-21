<script lang="ts">
  import { goto } from "$app/navigation";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { splitFolderAndFileName } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { WorkspaceHeader } from "../../layout/workspace";
  import type { ResourceKind } from "../entity-management/resource-selectors";
  import { PROTECTED_FILES } from "../file-explorer/protected-paths";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { Table } from "lucide-svelte";

  export let filePath: string;
  export let hasUnsavedChanges: boolean;
  export let resourceKind: ResourceKind | undefined;
  export let resource: V1Resource | undefined;

  let fileName: string;

  $: [, fileName] = splitFolderAndFileName(filePath);
  $: isProtectedFile = PROTECTED_FILES.includes(filePath);
  $: isEditableCSV = filePath.startsWith("/data/") && filePath.endsWith(".csv");

  $: ({ instanceId } = $runtime);

  const onChangeCallback = async (newTitle: string) => {
    const route = await handleEntityRename(
      instanceId,
      newTitle,
      filePath,
      fileName,
    );
    if (route) await goto(route);
  };

  async function handleEditInTable() {
    // Remove leading slash for the route
    const pathWithoutLeadingSlash = filePath.startsWith("/")
      ? filePath.slice(1)
      : filePath;
    await goto(`/mapping/edit/${pathWithoutLeadingSlash}`);
  }
</script>

<WorkspaceHeader
  {filePath}
  {resourceKind}
  editable={!isProtectedFile}
  onTitleChange={onChangeCallback}
  {hasUnsavedChanges}
  showInspectorToggle={false}
  titleInput={fileName}
  {resource}
>
  <svelte:fragment slot="cta">
    {#if isEditableCSV}
      <Button type="secondary" onClick={handleEditInTable}>
        <Table size="14px" />
        Edit in Table
      </Button>
    {/if}
  </svelte:fragment>
</WorkspaceHeader>
