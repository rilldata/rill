<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { GitBranch } from "lucide-svelte";
  import { openResourceGraphQuickView } from "@rilldata/web-common/features/resource-graph/quick-view/quick-view-store";

  export let filePath: string;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  const queryClient = useQueryClient();

  const instanceId = httpClient.getInstanceId();

  $: exploreQuery = fileArtifact.getResource(queryClient, instanceId);
  $: exploreResource = $exploreQuery.data;

  function viewGraph() {
    if (!exploreResource) {
      console.warn(
        "[ExploreMenuItems] Cannot open resource graph: resource unavailable.",
      );
      return;
    }
    openResourceGraphQuickView(exploreResource);
  }
</script>

<NavigationMenuItem on:click={viewGraph}>
  <GitBranch slot="icon" size="14px" />
  View dependency graph
</NavigationMenuItem>
