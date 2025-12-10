<script lang="ts">
  export let label: string;
</script>

<div class="info-row">
  <dt class="info-label">{label}</dt>
  <dd class="info-value">
    <slot />
  </dd>
</div>

<style lang="postcss">
  .info-row {
    @apply flex items-center;
    @apply border-b border-slate-200;
    @apply min-h-[44px];
  }

  .info-row:last-of-type {
    @apply border-b-0;
  }

  .info-label {
    @apply w-[140px] flex-shrink-0;
    @apply px-4 py-3;
    @apply text-sm font-medium text-gray-600;
    @apply bg-slate-50;
    @apply border-r border-slate-200;
    @apply whitespace-nowrap;
  }

  .info-value {
    @apply flex-1 px-4 py-3;
    @apply text-sm;
    @apply m-0; /* Reset default dd margin */
  }
</style>
