<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";

  export let instanceId: string;
  export let canvasName: string;

  $: canvasQuery = useResource(instanceId, canvasName, ResourceKind.Canvas);

  $: canvas = $canvasQuery.data?.canvas.spec;

  $: ({
    items = [],
    columns,
    gap,
  } = canvas || { items: [], columns: 24, gap: 2 });
</script>

<CanvasDashboardEmbed {canvasName} {columns} {items} {gap} />
