<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "../entity-management/resource-icon-mapping";
  import type { ResourceKind } from "../entity-management/resource-selectors";
  import { Code2Icon } from "lucide-svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let selectedView: "code" | "viz" | "split" = "viz";
  export let resourceKind: ResourceKind;

  $: viewOptions = [
    { view: "code", icon: Code2Icon },
    { view: "viz", icon: resourceIconMapping[resourceKind] },
  ];
</script>

<div class="radio relative">
  {#each viewOptions as { view, icon: Icon } (view)}
    <Tooltip activeDelay={700} distance={8}>
      <button
        aria-label="Switch to {view === 'viz' ? 'visual' : view} editor"
        id="{view}-toggle"
        class="size-[22px] z-10 hover:brightness-75"
        on:click={() => {
          if (selectedView === "code") {
            selectedView = "viz";
          } else {
            selectedView = "code";
          }
        }}
      >
        <Icon
          size="15px"
          color={view === selectedView && resourceKind
            ? resourceColorMapping[resourceKind]
            : "#9CA3AF"}
        />
      </button>
      <TooltipContent slot="tooltip-content">
        {view === "code" ? "Code view" : "No-code view"}
      </TooltipContent>
    </Tooltip>
  {/each}
  <span
    style:left={selectedView === "code" ? "2px" : "24px"}
    class="toggle size-[22px] pointer-events-none absolute rounded-[4px] z-0 transition-all"
  />
</div>

<style lang="postcss">
  button {
    @apply flex-none flex items-center justify-center rounded-[4px];
    @apply size-[22px] cursor-pointer;
  }

  .toggle {
    @apply bg-surface outline outline-slate-200 outline-[1px];
  }

  .radio {
    @apply h-fit bg-slate-100 p-[2px] rounded-[6px] flex;
  }
</style>
