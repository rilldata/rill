<script lang="ts">
  import {
    Divider,
    Menu,
    MenuItem,
  } from "@rilldata/web-common/components/menu";
  import { useDashboardableModels } from "@rilldata/web-common/features/models/selectors";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();
  $: dashboardableModels = useDashboardableModels($runtimeStore?.instanceId);
  let models = [];
  $: if ($dashboardableModels?.isSuccess) models = $dashboardableModels?.data;
</script>

<Menu dark on:item-select on:escape on:click-outside>
  {#if models?.length}
    <h2 class="px-2 text-gray-400 py-1" style:font-size="11px">
      bootstrap a dashboard from a model
    </h2>
    <Divider />
  {/if}

  {#each models as entry (entry.name)}
    <MenuItem
      on:select={() => {
        dispatch("bootstrap-dashboard", entry);
      }}
    >
      {entry.name}
    </MenuItem>
  {:else}{/each}
  {#if models?.length}
    <Divider />
  {/if}
  <MenuItem
    on:select={() => {
      dispatch("create-empty-dashboard");
    }}>Create a blank dashboard</MenuItem
  >
</Menu>
