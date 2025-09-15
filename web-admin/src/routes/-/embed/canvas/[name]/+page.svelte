<script lang="ts">
  import CanvasEmbed from "@rilldata/web-admin/features/embeds/CanvasEmbed.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;
  $: ({ instanceId, canvasName, navigationEnabled } = data);
</script>

<CanvasEmbed {instanceId} {canvasName} {navigationEnabled} />
