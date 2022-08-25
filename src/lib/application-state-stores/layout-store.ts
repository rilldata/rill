import { cubicInOut as easing } from "svelte/easing";
import { tweened } from "svelte/motion";
import { writable } from "svelte/store";

export const SURFACE_SLIDE_DURATION = 400;
export const SURFACE_SLIDE_EASING = easing;

export const SURFACE_DRAG_DURATION = 50;

export const layout = tweened(
  {
    assetsWidth: 401,
    inspectorWidth: 401,
    modelPreviewHeight: 400,
  },
  { duration: SURFACE_DRAG_DURATION }
);

export const SIDE_PAD = 32;

export const assetVisibilityTween = tweened(0, {
  duration: SURFACE_SLIDE_DURATION,
  easing: SURFACE_SLIDE_EASING,
});
export const inspectorVisibilityTween = tweened(0, {
  duration: SURFACE_SLIDE_DURATION,
  easing: SURFACE_SLIDE_EASING,
});
export const modelPreviewVisibilityTween = tweened(0, {
  duration: SURFACE_SLIDE_DURATION,
  easing: SURFACE_SLIDE_EASING,
});

export const assetsVisible = writable(true);
assetsVisible.subscribe((tf) => {
  assetVisibilityTween.set(tf ? 0 : 1);
});

export const inspectorVisible = writable(true);
inspectorVisible.subscribe((tf) => {
  inspectorVisibilityTween.set(tf ? 0 : 1);
});

export const modelPreviewVisible = writable(true);
modelPreviewVisible.subscribe((tf) => {
  modelPreviewVisibilityTween.set(tf ? 0 : 1);
});

export const importOverlayVisible = writable(false);

export const quickStartDashboardOverlay = writable({
  show: false,
  sourceName: "",
  timeDimension: "",
});
export function showQuickStartDashboardOverlay(
  sourceName: string,
  timeDimension: string
) {
  quickStartDashboardOverlay.set({
    show: true,
    sourceName,
    timeDimension,
  });
}
export function resetQuickStartDashboardOverlay() {
  quickStartDashboardOverlay.set({
    show: false,
    sourceName: "",
    timeDimension: "",
  });
}
