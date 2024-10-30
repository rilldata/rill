import type { SvelteComponent } from "svelte";
import { writable } from "svelte/store";

interface Overlay {
  title: string;
  detail?: {
    component: typeof SvelteComponent<any>;
    props: Record<string, unknown>;
  };
}

class OverlayStore {
  private timeout: NodeJS.Timeout;
  private isCleared: boolean = false;
  private overlayStore = writable<Overlay | null>(null);
  public subscribe = this.overlayStore.subscribe;

  public set(overlay: Overlay | null) {
    this.isCleared = false;
    this.overlayStore.set(overlay);
  }

  public setDebounced(overlay: Overlay | null, delay: number = 300) {
    this.isCleared = false;
    clearTimeout(this.timeout);
    this.timeout = setTimeout(() => {
      if (!this.isCleared) {
        this.overlayStore.set(overlay);
      }
    }, delay);
  }

  public clear() {
    this.isCleared = true;
    this.overlayStore.set(null);
  }
}

export const overlay = new OverlayStore();
