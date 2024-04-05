<script lang="ts">
  import { debounce } from "../lib/create-debouncer";
  import GlobalTooltip from "./GlobalTooltip.svelte";
  import type { Side, Align } from "./GlobalTooltip.svelte";

  let label: string | undefined | null = null;
  let description: string | null = null;
  let anchorElement: HTMLElement | null = null;
  let shortcuts: [string, string][] = [];
  let align: Align;
  let side: Side;

  let innerHeight: number;
  let innerWidth: number;

  const debouncedHandleTooltip = debounce(handleTooltip, 80);

  const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (
        mutation.target instanceof HTMLElement &&
        mutation.target.getAttribute("data-suppress") === "true"
      ) {
        reset();
      }
    });
  });

  function handleTooltip(
    e: MouseEvent & {
      currentTarget: EventTarget & Window;
    },
  ) {
    if (!(e.target instanceof HTMLElement) || e.target === anchorElement)
      return;

    // Temporary guard so the system can exist simultaneously with the old one
    if (!e.target.hasAttribute("data-tooltip")) return;

    if (e.target.getAttribute("data-suppress") === "true") return;

    label = e.target?.getAttribute("aria-label");

    side = (e.target.getAttribute("data-tooltip-side") ?? "right") as Side;

    align = (e.target.getAttribute("data-tooltip-align") ?? "center") as Align;

    e.target
      .getAttribute("data-actions")
      ?.split(",")
      .forEach((shortcut) => {
        const [modifier, action] = shortcut.split(":");
        shortcuts.push([modifier, action]);
      });

    anchorElement = e.target;

    observer.observe(anchorElement, {
      attributeFilter: ["data-suppress"],
    });

    const onMouseLeave = () => {
      anchorElement?.removeEventListener("mouseleave", onMouseLeave);

      reset();
    };

    anchorElement.addEventListener("mouseleave", onMouseLeave);
  }

  function reset() {
    observer.disconnect();
    label = null;
    anchorElement = null;
    shortcuts = [];
  }
</script>

<svelte:window
  bind:innerHeight
  bind:innerWidth
  on:click={reset}
  on:mousemove={debouncedHandleTooltip}
/>

<slot />

{#if label && anchorElement}
  <GlobalTooltip
    {side}
    {label}
    {align}
    {description}
    {anchorElement}
    {shortcuts}
    {innerHeight}
    {innerWidth}
  />
{/if}
