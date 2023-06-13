<script lang="ts">
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createEventDispatcher } from "svelte";
  import { useModelNames } from "../../models/selectors";

  $: models = useModelNames($runtime.instanceId);

  const dispatch = createEventDispatcher();

  const buttonClasses =
    "inline hover:font-semibold underline underline-offset-2";
</script>

<!-- completely empty case -->
<div class="whitespace-normal">
  Auto-generate a <WithTogglableFloatingElement
    inline
    let:toggleFloatingElement
    let:active
  >
    <Tooltip distance={8} suppress={active}>
      <button
        disabled={!$models?.data?.length}
        class={buttonClasses}
        on:click={toggleFloatingElement}
        >metrics configuration off of a model</button
      >
      <TooltipContent slot="tooltip-content"
        >Select a data model and auto-generate the config</TooltipContent
      ></Tooltip
    >
    <Menu
      slot="floating-element"
      on:click-outside={toggleFloatingElement}
      on:escape={toggleFloatingElement}
    >
      {#each $models?.data as model}
        <MenuItem on:click={() => dispatch("select-model", { model })}>
          {model}
        </MenuItem>
      {/each}
    </Menu>
  </WithTogglableFloatingElement>or
  <button class={buttonClasses}>start with a skeleton</button>.
</div>
