<script lang="ts">
  import { page } from "$app/stores";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.js";

  $: instanceId = $runtime?.instanceId;
  $: canvasName = $page.params.dashboard;

  $: canvasQuery = useResource(instanceId, canvasName, ResourceKind.Canvas);

  $: canvas = $canvasQuery.data?.canvas.spec;

  // TODO: Remove later
  let columns = 20;
  let gap = 1;
  $: ({ items = [] } = canvas || { items: [], columns: 24, gap: 2 });
</script>

<CanvasDashboardEmbed {columns} {items} {gap} />
