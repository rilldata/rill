<script lang="ts" context="module">
  import TooltipContent from "../components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "../components/tooltip/TooltipShortcutContainer.svelte";
  import Shortcut from "../components/tooltip/Shortcut.svelte";
  import TooltipTitle from "../components/tooltip/TooltipTitle.svelte";
  import { portal } from "../lib/actions/portal";
  import { onMount } from "svelte";

  export type Side = keyof typeof translateVectors;
  export type Align = keyof typeof alignValues;

  const buffer = 8;

  const translateVectors = {
    top: `-50%, calc(-100% - ${buffer}px)`,
    right: `${buffer}px, -50%`,
    bottom: `-50%, ${buffer}px`,
    left: `calc(-100% - ${buffer}px), -50%`,
  };

  function isMac() {
    return window.navigator.userAgent.includes("Macintosh");
  }

  const modifierChar = {
    command: isMac() ? "⌘" : "Ctrl",
    shift: "⇧",
    "shift-command": isMac() ? "⇧ + ⌘" : "Ctrl + Shift",
  };

  const alignValues = {
    center: 0.5,
    start: 0,
    end: 1,
  };
</script>

<script lang="ts">
  export let anchorElement: HTMLElement;
  export let innerWidth: number;
  export let innerHeight: number;
  export let label: string | undefined | null = null;
  export let description: string | null = null;
  export let shortcuts: [string, string][];
  export let side: Side = "right";
  export let align: Align = "center";
  export let skipBoundsCheck = false;

  const sideVectors = {
    top: [alignValues[align], 0],
    right: [1, alignValues[align]],
    bottom: [alignValues[align], 1],
    left: [0, alignValues[align]],
  };

  const {
    height: anchorHeight,
    width: anchorWidth,
    top: anchorTop,
    left: anchorLeft,
  } = anchorElement.getBoundingClientRect();

  let top: number;
  let left: number;
  let container: HTMLElement;
  let hidden = true;
  let xShift = 0;
  let yShift = 0;

  $: top = anchorTop + anchorHeight * sideVectors[side][1];

  $: left = anchorLeft + anchorWidth * sideVectors[side][0];

  onMount(() => {
    if (skipBoundsCheck) {
      hidden = false;
      return;
    }

    const { x, y, width, height } = container.getBoundingClientRect();

    if (x < 0) {
      if (side === "left") {
        side = "right";
      } else {
        xShift = -x + buffer;
      }
    }

    if (x + width > innerWidth) {
      if (side === "right") {
        side = "left";
      } else {
        xShift = x + width - innerWidth - buffer;
      }
    }

    if (y < 0) {
      if (side === "top") {
        side = "bottom";
      } else {
        yShift = -y + buffer;
      }
    }

    if (y + height > innerHeight) {
      if (side === "bottom") {
        side = "top";
      } else {
        yShift = y + height - innerHeight - buffer;
      }
    }

    hidden = false;
  });
</script>

<div
  id="tooltip"
  class="absolute z-50 pointer-events-none"
  class:opacity-0={hidden}
  style:left="{left + xShift}px"
  style:top="{top + yShift}px"
  style:transform="translate({translateVectors[side]})"
  use:portal
  bind:this={container}
>
  <TooltipContent>
    {#if shortcuts.length}
      <TooltipTitle>
        <svelte:fragment slot="name">
          {label}
        </svelte:fragment>

        <svelte:fragment slot="description">
          {#if description}
            {description}
          {/if}
        </svelte:fragment>
      </TooltipTitle>
    {:else}
      {label}
    {/if}

    {#each shortcuts as [modifier, action]}
      <TooltipShortcutContainer>
        <div>
          {action}
        </div>
        <Shortcut>
          {#if modifierChar[modifier]}
            <span style:font-size="11.5px" style:font-family="var(--system)">
              {modifierChar[modifier]} +
            </span>
          {/if}
          Click
        </Shortcut>
      </TooltipShortcutContainer>
    {/each}
  </TooltipContent>
</div>
