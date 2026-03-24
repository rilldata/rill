<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { GitBranch } from "lucide-svelte";
  import { goto } from "$app/navigation";
  import { resourceShorthandMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { type ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  export let filePath: string;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  const queryClient = useQueryClient();

  $: exploreQuery = fileArtifact.getResource(queryClient);
  $: exploreResource = $exploreQuery.data;

  function viewGraph() {
    const name = exploreResource?.meta?.name?.name;
    const kind = exploreResource?.meta?.name?.kind as ResourceKind | undefined;
    if (!name || !kind) return;
    const shortKind = resourceShorthandMapping[kind];
    goto(`/graph?resource=${encodeURIComponent(`${shortKind}:${name}`)}`);
  }
</script>

<NavigationMenuItem onclick={viewGraph}>
  <GitBranch slot="icon" size="14px" />
  View Resource Graph
</NavigationMenuItem>
