<script lang="ts">
  import { goto } from "$app/navigation";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { resourceShorthandMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { type ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { GitBranch } from "lucide-svelte";

  export let filePath: string;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  const queryClient = useQueryClient();

  $: ({ instanceId } = $runtime);

  $: connectorQuery = fileArtifact.getResource(queryClient, instanceId);
  $: connectorResource = $connectorQuery.data;

  function viewGraph() {
    const name = connectorResource?.meta?.name?.name;
    const kind = connectorResource?.meta?.name?.kind as
      | ResourceKind
      | undefined;
    if (!name || !kind) return;
    const shortKind = resourceShorthandMapping[kind];
    goto(`/graph?resource=${encodeURIComponent(`${shortKind}:${name}`)}`);
  }
</script>

<NavigationMenuItem on:click={viewGraph}>
  <GitBranch slot="icon" size="14px" />
  View DAG graph
</NavigationMenuItem>
