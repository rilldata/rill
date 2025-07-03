<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  export let size = "1em";
  export let status: EntityStatus = EntityStatus.Idle;
  export let duration = 500;

  $: idle = status === EntityStatus.Idle;
</script>

<div
  class="status bg-gradient-to-b from-primary-500 to-secondary-500"
  class:idle
  style:--status-transition="{duration}ms"
  style:width={size}
  style:height={size}
/>

<style>
  div {
    transform-origin: center;
    border-radius: 1px;
    transition:
      border-radius var(--status-transition),
      border-color var(--status-transition);
    animation: reverse-spin calc(var(--status-transition) * 2) infinite;
  }

  .idle {
    border-radius: 50%;
    border-color: currentColor;
  }

  @keyframes reverse-spin {
    from {
      transform: rotate(360deg);
    }
    to {
      transform: rotate(0deg);
    }
  }
</style>
