<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import type { Component, Snippet } from "svelte";

  let {
    href,
    title,
    icon: Icon,
    iconSize = "14px",
    errorMessage = undefined,
    tags,
    subtitle,
    actions,
  }: {
    href: string;
    title: string;
    icon?: Component<{ size: string }>;
    iconSize?: string;
    errorMessage?: string | undefined;
    tags?: Snippet;
    subtitle?: Snippet;
    actions?: Snippet;
  } = $props();

  const GAP_PX = 8; // gap-x-2
  const subtitlePadding = $derived(
    Icon ? `${parseInt(iconSize) + GAP_PX}px` : "0",
  );
</script>

<div class="flex items-center group w-full h-full">
  <a {href} class="flex items-center min-w-0 flex-1 px-4 py-2.5 h-full">
    <div class="flex flex-col gap-y-1 min-w-0 flex-1">
      <div class="flex gap-x-2 items-center min-h-[20px]">
        {#if tags}{@render tags()}{/if}
        {#if Icon}
          <Icon size={iconSize} />
        {/if}
        <span
          class="text-fg-secondary text-sm font-semibold group-hover:text-accent-primary-action truncate"
        >
          {title}
        </span>
        {#if errorMessage}
          <Tag color="red">Error</Tag>
        {/if}
      </div>
      <div
        class="flex gap-x-1 text-fg-tertiary text-xs font-normal min-h-[16px] overflow-hidden"
        style:padding-left={subtitlePadding}
      >
        {#if subtitle}{@render subtitle()}{/if}
      </div>
    </div>
  </a>
  {#if actions}
    <div class="shrink-0 pr-4">
      {@render actions()}
    </div>
  {/if}
</div>
