<script lang="ts">
  import { Switch } from "@rilldata/web-common/components/button/index";
  import { getPrettySelectedTimeRange } from "@rilldata/web-common/features/bookmarks/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  /**
   * This extracts filtersOnly and absoluteTimeRange switches.
   * Both create and edit have the exact same functionality.
   */

  export let metricsViewName: string;
  export let formState: any; // svelte-forms-lib's FormState
  const { form } = formState;

  const queryClient = useQueryClient();
  $: selectedTimeRange = getPrettySelectedTimeRange(
    queryClient,
    $runtime?.instanceId,
    metricsViewName,
  );
</script>

<Switch
  checked={$form["filtersOnly"]}
  id="filtersOnly"
  on:click={() => ($form["filtersOnly"] = !$form["filtersOnly"])}
>
  <div class="font-medium text-sm">Save filters only</div>
</Switch>
<Switch
  bind:checked={$form["absoluteTimeRange"]}
  id="absoluteTimeRange"
  on:click={() => ($form["absoluteTimeRange"] = !$form["absoluteTimeRange"])}
>
  <div class="flex flex-col">
    <div class="text-left text-sm font-medium">Absolute time range</div>
    <div class="text-gray-500 text-sm">{$selectedTimeRange}</div>
  </div>
</Switch>
