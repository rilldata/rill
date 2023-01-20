<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/entity";

  export let size = "1em";
  export let status: EntityStatus = EntityStatus.Idle;
  export let bg =
    "linear-gradient(to left, hsla(300, 100%, 50%, .5), hsla(1, 100%, 50%, .5))";
  export let duration = 500;
</script>

<div
  class="status"
  class:running={status === EntityStatus.Running}
  class:idle={status === EntityStatus.Idle}
  style="
		--status-transition: {duration}ms;
		--background: {bg};
		--size: {size};
		width: {size}; height: {size};"
/>

<style>
  div {
    border-radius: 0px;
    transition: border-radius var(--status-transition),
      border-color var(--status-transition);
    border: 0.125rem solid rgba(0, 0, 0, 0);
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
    transition: opacity var(--status-transition),
      border-radius var(--status-transition), transform var(--status-transition);
    background: var(--background);
    /* transform: rotate(0deg); */
  }

  div::after {
    content: " ";
  }

  .running {
    /* border-radius: 0px; */
  }

  .running::before {
    content: " ";
    display: block;
    width: 100%;
    height: 100%;
    opacity: 1;
    border-radius: 0px;
    transition: opacity var(--status-transition),
      border-radius var(--status-transition), transform var(--status-transition);
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
    transition: opacity var(--status-transition),
      border-radius var(--status-transition), transform var(--status-transition);
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
