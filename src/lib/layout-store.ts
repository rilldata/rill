import { writable } from "svelte/store";
import { tweened } from "svelte/motion";
import { cubicInOut as easing } from "svelte/easing";

export const SURFACE_SLIDE_DURATION = 400;
export const SURFACE_SLIDE_EASING = easing;

export const SURFACE_DRAG_DURATION = 50;

export const layout = tweened({
    assetsWidth: 400,
    inspectorWidth: 400,
  }, { duration: SURFACE_DRAG_DURATION }
);



export const assetVisibilityTween = tweened(0, { duration: SURFACE_SLIDE_DURATION, easing: SURFACE_SLIDE_EASING });
export const inspectorVisibilityTween = tweened(0, { duration: SURFACE_SLIDE_DURATION, easing: SURFACE_SLIDE_EASING });

export const assetsVisible = writable(true);
assetsVisible.subscribe((tf) => {
    assetVisibilityTween.set(tf ? 0 : 1);
})

export const inspectorVisible = writable(true);
inspectorVisible.subscribe((tf) => {
    inspectorVisibilityTween.set(tf ? 0 : 1);
})

