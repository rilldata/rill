import { writable } from "svelte/store";

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

interface Overlay {
  title: string;
  message?: string;
}

export const overlay = writable<Overlay>(null);
