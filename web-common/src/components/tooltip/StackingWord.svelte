<script lang="ts">
  import { onMount } from "svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let key: "command" | "shift";

  let span: HTMLSpanElement;

  onMount(() => {
    const unsubscribe = eventBus.on(`${key}-click`, () => {
      if (!span) return;
      span.animate(
        [
          {
            transform: "translateY(2px) translateX(2px)",
            boxShadow: "-2px -2px 0px var(--color-gray-600)",
          },
        ],
        { duration: 250, easing: "ease-in-out" },
      );
    });

    return unsubscribe;
  });
</script>

<span bind:this={span}><slot /></span>

<style lang="postcss">
  @reference "tailwindcss";

  span {
    @apply inline-block;
    border-radius: 2px;
    position: relative;
  }
</style>
