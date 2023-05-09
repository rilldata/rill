<!-- @component
Combines a Navigation element with a slot for a WorkspaceContainer.
BasicLayout is the backbone of the Rill application.
-->
<script lang="ts">
  import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
  import { setContext } from "svelte";
  import { tweened } from "svelte/motion";
  import {
    SURFACE_DRAG_DURATION,
    SURFACE_SLIDE_DURATION,
    SURFACE_SLIDE_EASING,
  } from "./config";
  import Navigation from "./navigation/Navigation.svelte";

  /** navigation element layout*/
  const navigationLayout = localStorageStore("navigation-layout", {
    value: 400,
    visible: true,
  });

  const navigationWidth = tweened($navigationLayout.value || 400, {
    duration: SURFACE_DRAG_DURATION,
  });

  export const navVisibilityTween = tweened(
    $navigationLayout?.visible ? 0 : 1,
    {
      duration: SURFACE_SLIDE_DURATION,
      easing: SURFACE_SLIDE_EASING,
    }
  );

  navigationLayout.subscribe((state) => {
    navigationWidth.set(state.value);
    navVisibilityTween.set(state.visible ? 0 : 1);
  });

  setContext("rill:app:navigation-layout", navigationLayout);
  setContext("rill:app:navigation-width-tween", navigationWidth);
  setContext("rill:app:navigation-visibility-tween", navVisibilityTween);
</script>

<main>
  <Navigation />
  <slot />
</main>
