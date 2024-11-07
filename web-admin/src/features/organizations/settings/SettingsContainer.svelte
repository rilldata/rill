<script lang="ts">
  import InfoCircleFilled from "@rilldata/web-common/components/icons/InfoCircleFilled.svelte";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  export let title: string;
  export let titleIcon: "none" | "info" | "error" = "none";
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
      <slot name="body" />
      {#if $$slots["default"]}
        <slot />
      {/if}
    </div>
  </div>
  {#if $$slots["contact"] || $$slots["action"]}
    <div class="settings-footer">
      <slot name="contact" />
      {#if $$slots["action"]}
        <div class="grow"></div>
      {/if}
      <slot name="action" />
    </div>
  {/if}
</div>

<style lang="postcss">
  .settings-container {
    @apply w-full max-w-[844px] border border-slate-200 text-slate-700;
  }

  .settings-header {
    @apply p-5;
  }

  .settings-title {
    @apply flex flex-row gap-x-2 items-center mb-2;
    @apply text-lg font-semibold;
  }

  .settings-body {
    @apply text-sm text-slate-800;
  }

  .settings-footer {
    @apply flex flex-row items-center px-5 py-2;
    @apply bg-slate-50 text-slate-500 text-sm border-t border-slate-200;
  }
</style>
