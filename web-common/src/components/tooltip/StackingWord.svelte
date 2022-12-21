<script lang="ts">
  import { getContext } from "svelte";
  import transientBooleanStore from "../../lib/transient-boolean-store";

  export let isStacked = false;
  export let key: "command" | "shift";

  // NOTE: Using these two different contexts is tech debt. Ideally, we would have one "click-action-callbacks".
  // We have to refactor `shift-click-action.ts` to account for multiple kinds of key clicks.
  let keyCallbacks;
  if (key === "command") {
    keyCallbacks = getContext("rill:app:ui:command-click-action-callbacks");
  } else if (key === "shift") {
    keyCallbacks = getContext("rill:app:ui:shift-click-action-callbacks");
  }

  let keyClicked = transientBooleanStore();
  // if a parent component upstream triggers the shift-click action,
  // let's flip our transientBooleanStore to create the animation.
  if (keyCallbacks) {
    keyCallbacks.addCallback(() => keyClicked.flip());
  }
</script>

<span
  class="inline-block shiftable"
  class:keyClicked={!isStacked && $keyClicked}
  class:stacked={isStacked}><slot /></span
>

<style>
  .shiftable {
    padding-left: 2px;
    margin-right: -2px;
    transform: translateY(0px) translateX(-2px);
    transition: transform 200ms;
  }
  .keyClicked {
    animation: pulse 250ms;
    border-radius: 2px;
    position: relative;
    mix-blend-mode: screen;
    background-blend-mode: screen;
  }
  @keyframes pulse {
    0%,
    100% {
      transform: translateY(0px) translateX(-2px);
    }
    50% {
      transform: translateY(2px) translateX(2px);
      box-shadow: -1px -1px 0px rgba(100, 100, 100, 1),
        -2px -2px 0px rgba(75, 75, 75, 1), -3px -3px 0px rgba(50, 50, 50, 1);
    }
  }

  .stacked {
    transform: translateY(2px) translateX(2px);
    box-shadow: -1px -1px 0px rgba(100, 100, 100, 1),
      -2px -2px 0px rgba(75, 75, 75, 1), -3px -3px 0px rgba(50, 50, 50, 1);
  }
</style>
