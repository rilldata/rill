<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ResourcesOverviewSection from "@rilldata/web-common/features/resources/overview/ResourcesOverviewSection.svelte";
  import { countByKind } from "@rilldata/web-common/features/resources/overview-utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useResources } from "../selectors";

  $: ({ instanceId } = $runtime);
  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/status`;

  $: resources = useResources(instanceId);
  $: allResources = $resources.data?.resources ?? [];
  $: resourceCounts = countByKind(allResources);
</script>

<ResourcesOverviewSection
  {resourceCounts}
  onViewAll={() => goto(`${basePage}/resources`)}
  onChipClick={(kind) => goto(`${basePage}/resources?kind=${kind}`)}
/>
