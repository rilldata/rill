<script lang="ts">
  import { Switch } from "@rilldata/web-common/components/button";
  import { createEventDispatcher } from "svelte";
  import { getStateManagers } from "../state-managers/state-managers";

  const stateManagers = getStateManagers();

  $: nestedMode = false;

  function toggleNestedMode() {
    console.log("toggleNestedMode");
  }
</script>

<div class="flex items-center p-2 mt-1 gap-x-4">
  <Switch checked={nestedMode} on:click={() => toggleNestedMode()}>
    Nested
  </Switch>
  <div>Expand All</div>
</div>
