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

<a {href} class="flex items-center group px-4 py-2.5 w-full h-full">
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
  {#if actions}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div
      class="shrink-0"
      onclick={(e) => {
        e.preventDefault();
        e.stopPropagation();
      }}
      role="presentation"
    >
      {@render actions()}
    </div>
  {/if}
</a>
