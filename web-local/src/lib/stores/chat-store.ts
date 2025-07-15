import { writable } from "svelte/store";

// Chat widget state
export const chatOpen = writable(false);

// Chat actions
export const chatActions = {
  toggleChat(): void {
    chatOpen.update((isOpen) => !isOpen);
  },
  openChat(): void {
    chatOpen.set(true);
  },
  closeChat(): void {
    chatOpen.set(false);
  },
};
