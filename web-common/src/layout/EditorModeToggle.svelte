<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { Code2Icon, LayoutDashboardIcon } from "lucide-svelte";
  import { editorMode, type EditorMode } from "./editor-mode-store";

  const options: Array<{
    value: EditorMode;
    icon: typeof Code2Icon;
    label: string;
  }> = [
    { value: "code", icon: Code2Icon, label: "Code" },
    { value: "visual", icon: LayoutDashboardIcon, label: "Visual" },
  ];
</script>

<div class="radio relative" role="radiogroup" aria-label="Editor mode">
  {#each options as { value, icon: Icon, label } (value)}
    <Tooltip activeDelay={700} distance={8}>
      <button
        aria-label="Switch to {label.toLowerCase()} mode"
        aria-checked={$editorMode === value}
        role="radio"
        class="size-[22px] z-10 hover:brightness-75 p-0"
        onclick={() => editorMode.set(value)}
      >
        <Icon size="15px" />
      </button>
      <TooltipContent slot="tooltip-content">
        {label} mode
      </TooltipContent>
    </Tooltip>
  {/each}
  <span
    style:left={$editorMode === "code" ? "2px" : "24px"}
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

  .radio {
    @apply h-fit bg-surface-subtle border p-0.5 rounded-[6px] flex w-fit;
  }
</style>
