<!--
  Renders a planning indicator shown while waiting for the first AI response.
-->
<script lang="ts">
  import Brain from "../../../../../components/icons/Brain.svelte";
  import ShimmerText from "./ShimmerText.svelte";
</script>

<div class="planning-block">
  <div class="planning-header">
    <div class="planning-icon">
      <Brain />
    </div>
    <div class="planning-title">
      <ShimmerText>Planning next moves...</ShimmerText>
    </div>
  </div>
</div>

<style lang="postcss">
  .planning-block {
    @apply w-full max-w-full self-start;
  }

  .planning-header {
    @apply w-full flex items-center gap-1.5 px-1 py-1;
    @apply text-xs text-gray-500;
  }

  .planning-icon {
    @apply flex items-center;
  }

  .planning-title {
    @apply flex-1 text-left font-normal;
  }
</style>
