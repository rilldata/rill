<script lang="ts">
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { getPrettySelectedTimeRange } from "@rilldata/web-admin/features/bookmarks/selectors.js";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.js";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let metricsViewName: string;
  export let exploreName: string;
  export let checked: boolean;

  $: ({ instanceId } = $runtime);
  $: selectedTimeRange = getPrettySelectedTimeRange(
    queryClient,
    instanceId,
    metricsViewName,
    exploreName,
  );
</script>

<div class="flex items-center space-x-2">
  <Switch bind:checked id="absoluteTimeRange" />
  <Label class="flex flex-col" for="absoluteTimeRange">
    <div class="text-left text-sm font-medium">Absolute time range</div>
    <div class="text-gray-500 text-sm">{$selectedTimeRange}</div>
  </Label>
</div>
