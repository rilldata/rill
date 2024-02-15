<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  export let size = "1em";
  export let status: EntityStatus = EntityStatus.Idle;

  export let duration = 500;
</script>

<div
  class="status bg-gradient-to-b from-primary-500 to-secondary-500"
  class:running={status === EntityStatus.Running}
  class:idle={status === EntityStatus.Idle}
  style="
    --status-transition: {duration}ms;
    --size: {size};
    width: {size};
    height: {size};"
/>

<style>
  div {
    border-radius: 0px;
    transition:
      border-radius var(--status-transition),
      border-color var(--status-transition);
    animation: spin calc(var(--status-transition) * 2) infinite;
    background-color: transparent;
  }

  div::before {
    content: " ";
    display: block;
    width: 100%;
    height: 100%;
    opacity: 0;
    border-radius: 0px;
    transition:
      opacity var(--status-transition),
      border-radius var(--status-transition),
      transform var(--status-transition);
    background: var(--background);
    /* transform: rotate(0deg); */
  }

  div::after {
    content: " ";
  }

  .running::before {
    content: " ";
    display: block;
    width: 100%;
    height: 100%;
    opacity: 1;
    border-radius: 0px;
    transition:
      opacity var(--status-transition),
      border-radius var(--status-transition),
      transform var(--status-transition);
  }

  .idle {
    border-radius: 50%;
    border-color: currentColor;
  }

  .idle::before {
    content: " ";
    display: block;
    width: 100%;
    height: 100%;
    opacity: 0;
    border-radius: 10rem;
    transform: rotate(-180deg);
    transition:
      opacity var(--status-transition),
      border-radius var(--status-transition),
      transform var(--status-transition);
  }

  @keyframes spin {
    0% {
      transform: rotate(360deg);
    }
    100% {
      transform: rotate(0deg);
    }
  }
</style>
