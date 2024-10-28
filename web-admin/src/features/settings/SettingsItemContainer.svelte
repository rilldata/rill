<script lang="ts">
  import InfoCircleFilled from "@rilldata/web-common/components/icons/InfoCircleFilled.svelte";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";

  export let title: string;
  export let titleIcon: "none" | "info" | "error" = "none";
</script>

<div class="plan-container">
  <div class="plan-title">
    <span>{title}</span>
    {#if titleIcon === "info"}
      <InfoCircleFilled />
    {:else if titleIcon === "error"}
      <CancelCircle />
    {/if}
  </div>
  <slot name="description" />
  {#if $$slots["contact"] || $$slots["action"]}
    <div class="plan-footer">
      <slot name="contact" />
      {#if $$slots["action"]}
        <div class="grow"></div>
      {/if}
      <slot name="action" />
    </div>
  {/if}
</div>

<style lang="postcss">
  .plan-container {
    @apply w-[800px] border border-slate-200 m-5;
  }

  .plan-title {
    @apply flex flex-row items-center;
    @apply text-lg font-semibold;
  }

  .plan-footer {
    @apply flex flex-row items-center;
    @apply bg-slate-50 text-slate-500 border-t border-slate-200;
  }
</style>
