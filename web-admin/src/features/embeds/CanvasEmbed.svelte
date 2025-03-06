<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";

  export let instanceId: string;
  export let canvasName: string;

  $: canvasQuery = useResource(instanceId, canvasName, ResourceKind.Canvas);

  $: resource = $canvasQuery.data;
</script>

<CanvasDashboardEmbed {resource} />
