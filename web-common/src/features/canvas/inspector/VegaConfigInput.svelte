<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { CanvasComponentObj } from "@rilldata/web-common/features/canvas/components/util";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";

  export let component: CanvasComponentObj;
  export let paramValues: V1ComponentSpecRendererProperties;

  const KEY = "vl_config";

  let error: string | null = null;

  async function updateConfig() {
    let config = paramValues[KEY];
    if (!config) {
      return;
    }

    // config is a string, we need to validate if it is a valid JSON
    try {
      config = JSON.parse(config);
      config = JSON.stringify(config, null, 2);
    } catch (e) {
      error = "Invalid JSON";
      return;
    }

    error = null;
    component.updateProperty(KEY, config);
  }
</script>

<div class="config-input">
  <InputLabel small label="Vega Lite config" optional id="vl-config" />
  <textarea
    class="w-full mt-2 p-2 border border-gray-300 rounded-sm"
    rows="20"
    bind:value={paramValues[KEY]}
    on:blur={updateConfig}
    placeholder="Optionally enter a Vega Lite config"
  />

  {#if error}
    <div class="text-red-500 text-sm">{error}</div>
  {/if}
</div>

<style lang="postcss">
  .config-input {
    @apply py-3 px-5;
    @apply border-t border-gray-200;
  }
</style>
