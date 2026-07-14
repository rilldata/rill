<script lang="ts">
  import type { Snippet } from "svelte";
  import InfoCircleFilled from "@rilldata/web-common/components/icons/InfoCircleFilled.svelte";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";

  let {
    title,
    titleIcon = "none",
    children,
    contact,
    action,
  }: {
    title: string;
    titleIcon?: "none" | "info" | "error";
    children?: Snippet;
    contact?: Snippet;
    action?: Snippet;
  } = $props();
</script>

<div class="settings-container">
  <div class="settings-header">
    <div class="settings-title">
      <span>{title}</span>
      {#if titleIcon === "info"}
        <InfoCircleFilled className="text-yellow-500" size="14px" />
      {:else if titleIcon === "error"}
        <CancelCircle className="text-red-600" size="14px" />
      {/if}
    </div>
    <div class="settings-body">
      {#if children}
        {@render children()}
      {/if}
    </div>
  </div>
  {#if contact || action}
    <div class="settings-footer">
      {#if contact}
        {@render contact()}
      {/if}
      {#if action}
        <div class="grow"></div>
        {@render action()}
      {/if}
    </div>
  {/if}
</div>

<style lang="postcss">
  .settings-container {
    @apply w-full border text-fg-secondary rounded-sm bg-surface-background;
  }

  .settings-header {
    @apply p-5;
  }

  .settings-title {
    @apply flex flex-row gap-x-2 items-center mb-2;
    @apply text-lg font-semibold text-fg-primary;
  }

  .settings-body {
    @apply text-sm text-fg-tertiary;
  }

  .settings-footer {
    @apply flex flex-row items-center px-5 py-2;
    @apply bg-surface-subtle text-fg-tertiary text-sm border-t;
  }
</style>
