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
        <InfoCircleFilled />
      {:else if titleIcon === "error"}
        <CancelCircle />
      {/if}
    </div>
    <slot name="body" />
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
    @apply w-full max-w-[844px] border border-slate-200;
  }

  .settings-header {
    @apply p-5;
  }

  .settings-title {
    @apply flex flex-row items-center;
    @apply text-lg font-semibold;
  }

  .settings-footer {
    @apply flex flex-row items-center px-5 py-2;
    @apply bg-slate-50 text-slate-500 border-t border-slate-200;
  }
</style>
