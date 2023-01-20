<script lang="ts">
  import {
    LIST_SLIDE_DURATION,
    SIDE_PAD,
  } from "@rilldata/web-local/lib/application-config";
  import { getContext } from "svelte";
  import { tweened } from "svelte/motion";
  import type { Writable } from "svelte/store";

  export let bgClass: string = "";
  export let top: string = undefined;
  export let right = true;

  const navigationWidth = getContext(
    "rill:app:navigation-width-tween"
  ) as Writable<number>;

  const navVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Writable<number>;

  const visibilityTween = getContext(
    "rill:app:inspector-visibility-tween"
  ) as Writable<number>;

  const inspectorWidth = getContext(
    "rill:app:inspector-width-tween"
  ) as Writable<number>;

  let userSetRight = tweened(right ? 1 : 0, {
    duration: LIST_SLIDE_DURATION,
  });
  $: userSetRight.set(right ? 1 : 0);
</script>

<div
  class="box-border fixed {bgClass}"
  style:top
  style:left="{($navigationWidth || 0) * (1 - $navVisibilityTween)}px"
  style:padding-left="{$navVisibilityTween * SIDE_PAD}px"
  style:padding-right="{(1 - $visibilityTween) * SIDE_PAD}px"
  style:right="{$inspectorWidth * $visibilityTween * $userSetRight}px"
>
  <slot />
</div>
