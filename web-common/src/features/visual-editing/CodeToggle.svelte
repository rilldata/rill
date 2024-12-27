<script lang="ts">
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "../entity-management/resource-icon-mapping";
  import type { ResourceKind } from "../entity-management/resource-selectors";
  import { Code2Icon } from "lucide-svelte";

  export let selectedView: "code" | "no-code" | "split" = "no-code";
  export let resourceKind: ResourceKind;

  $: viewOptions = [
    { view: "code", icon: Code2Icon },
    { view: "no-code", icon: resourceIconMapping[resourceKind] },
  ];
</script>

<div class="radio" role="radiogroup">
  {#each viewOptions as { view, icon: Icon } (view)}
    <input
      tabindex="0"
      type="radio"
      id={view}
      name="view"
      class="screenreader-only"
      value={view}
      checked={view === selectedView}
      bind:group={selectedView}
    />

    <label aria-label={view} for={view} title={view}>
      <Icon
        size="15px"
        color={view === selectedView && resourceKind
          ? resourceColorMapping[resourceKind]
          : "#9CA3AF"}
      />
    </label>
  {/each}
</div>

<style lang="postcss">
  label {
    @apply flex-none flex items-center justify-center rounded-[4px];
    @apply size-[22px] cursor-pointer;
  }

  .screenreader-only {
    position: absolute;
    clip: rect(1px, 1px, 1px, 1px);
    padding: 0;
    border: 0;
    height: 1px;
    width: 1px;
    overflow: hidden;
  }

  input:checked + label {
    @apply bg-white outline outline-slate-200 outline-[1px];
  }

  .radio {
    @apply h-fit bg-slate-100 p-[2px] rounded-[6px] flex;
  }
</style>
