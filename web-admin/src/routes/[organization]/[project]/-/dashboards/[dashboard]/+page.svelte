<script lang="ts">
  import { page } from "$app/stores";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.js";

  $: ({ instanceId } = $runtime);
  $: canvasName = $page.params.dashboard;

  $: canvasQuery = useResource(instanceId, canvasName, ResourceKind.Canvas);

  $: canvas = $canvasQuery.data?.canvas.spec;

  $: ({ items = [], filtersEnabled } = canvas || {
    items: [],
    filtersEnabled: true,
  });
</script>

<CanvasDashboardEmbed {items} showFilterBar={filtersEnabled} spec={canvas} />
