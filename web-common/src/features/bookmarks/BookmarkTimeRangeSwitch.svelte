<script lang="ts">
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { useQueryClient } from "@rilldata/svelte-query";
  import { getPrettySelectedTimeRange } from "@rilldata/web-common/features/bookmarks/selectors.js";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.js";

  export let metricsViewName: string;
  export let checked: boolean;

  const queryClient = useQueryClient();
  $: selectedTimeRange = getPrettySelectedTimeRange(
    queryClient,
    $runtime?.instanceId,
    metricsViewName,
  );
</script>

<div class="flex items-center space-x-2">
  <Switch bind:checked id="absoluteTimeRange" />
  <Label class="flex flex-col" for="absoluteTimeRange">
    <div class="text-left text-sm font-medium">Absolute time range</div>
    <div class="text-gray-500 text-sm">{$selectedTimeRange}</div>
  </Label>
</div>
