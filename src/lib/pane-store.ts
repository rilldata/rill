import { writable } from "svelte/store";
import { tweened } from "svelte/motion";
import { cubicInOut as easing } from "svelte/easing";

export const panes = tweened({
    left: 400,
    right: 400
  }, { duration: 50}
);

export const assetVisibilityTween = tweened(0, { duration: 400, easing });
export const inspectorVisibilityTween = tweened(0, { duration: 400, easing });

export const assetsVisible = writable(true);
assetsVisible.subscribe((tf) => {
    assetVisibilityTween.set(tf ? 0 : 1);
})

export const inspectorVisible = writable(true);
inspectorVisible.subscribe((tf) => {
    inspectorVisibilityTween.set(tf ? 0 : 1);
})

