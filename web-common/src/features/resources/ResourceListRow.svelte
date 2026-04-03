<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import type { ComponentType } from "svelte";

  export let href: string;
  export let title: string;
  export let icon: ComponentType;
  export let iconSize: string = "14px";
  export let errorMessage: string | undefined = undefined;
</script>

<a
  {href}
  class="flex items-center group px-4 py-2.5 w-full h-full"
>
  <div class="flex flex-col gap-y-1 min-w-0 flex-1">
    <div class="flex gap-x-2 items-center min-h-[20px]">
      <svelte:component this={icon} size={iconSize} />
      <span
        class="text-fg-secondary text-sm font-semibold group-hover:text-accent-primary-action truncate"
      >
        {title}
      </span>
      {#if errorMessage}
        <Tag color="red">Error</Tag>
      {/if}
      <slot name="tags" />
    </div>
    <div
      class="flex gap-x-1 text-fg-tertiary text-xs font-normal min-h-[16px] overflow-hidden pl-[22px]"
    >
      <slot name="subtitle" />
    </div>
  </div>
  <button
    class="shrink-0 text-fg-secondary hover:text-fg-primary grid place-items-center rounded-sm p-1 opacity-0 group-hover:opacity-100 transition-opacity"
    on:click|preventDefault|stopPropagation={() => {}}
  >
    <ThreeDot size="16px" />
  </button>
</a>
