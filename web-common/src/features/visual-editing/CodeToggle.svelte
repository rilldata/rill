<script lang="ts" context="module">
  import type { ComponentType } from "svelte";
  import type { WorkspaceView } from "@rilldata/web-common/layout/workspace/workspace-stores";

  export type ViewOption = {
    view: WorkspaceView;
    icon: ComponentType;
    label: string;
  };
</script>

<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { resourceIconMapping } from "../entity-management/resource-icon-mapping";
  import type { ResourceKind } from "../entity-management/resource-selectors";
  import { Code2Icon } from "lucide-svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let selectedView: WorkspaceView = "viz";
  export let resourceKind: ResourceKind;
  export let views: ViewOption[] | undefined = undefined;

  $: viewOptions = views ?? [
    { view: "code", icon: Code2Icon, label: "Code view" },
    { view: "viz", icon: resourceIconMapping[resourceKind], label: "No-code view" },
  ];

  $: selectedIndex = Math.max(
    0,
    viewOptions.findIndex(({ view }) => view === selectedView),
  );

  const ariaLabels: Partial<Record<WorkspaceView, string>> = {
    code: "Switch to code editor",
    viz: "Switch to visual editor",
    explore: "Switch to explore editor",
  };
</script>

<div class="radio relative">
  {#each viewOptions as { view, icon: Icon, label } (view)}
    <Tooltip activeDelay={700} distance={8}>
      <button
        aria-label={ariaLabels[view] ?? `Switch to ${view} editor`}
        id="{view}-toggle"
        class="size-[22px] z-10 hover:brightness-75 p-0"
        onclick={() => {
          selectedView = view;
        }}
      >
        <Icon size="15px" />
      </button>
      <TooltipContent slot="tooltip-content">
        {label}
      </TooltipContent>
    </Tooltip>
  {/each}
  <span
    style:left="{2 + selectedIndex * 22}px"
    class="toggle size-[22px] pointer-events-none absolute rounded-[4px] z-0 transition-[left]"
  ></span>
</div>

<style lang="postcss">
  button {
    @apply flex-none flex items-center justify-center rounded-[4px];
    @apply size-[22px] cursor-pointer;
  }

  .toggle {
    @apply bg-surface-hover;
  }

  .toggle:hover {
    @apply bg-surface-hover;
  }

  .radio {
    @apply h-fit bg-surface-subtle border p-0.5 rounded-[6px] flex;
  }
</style>
