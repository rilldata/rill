<script lang="ts">
  import ExploreEmbed from "@rilldata/web-admin/features/embeds/ExploreEmbed.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;
  $: ({ instanceId, exploreName, dynamicHeight } = data);
</script>

<ExploreEmbed {instanceId} {exploreName} {dynamicHeight} />
