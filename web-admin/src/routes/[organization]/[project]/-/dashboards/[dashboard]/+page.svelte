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

  $: ({ items = [] } = canvas || { items: [] });
</script>

<CanvasDashboardEmbed {items} />
