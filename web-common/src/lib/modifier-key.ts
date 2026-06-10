import { writable, type Readable } from "svelte/store";

// Tracks whether the platform's "modify" key (CMD on macOS, Ctrl elsewhere)
// is currently held. Subscribe via `$modifierHeld` for reactive UI; read
// the value at the moment of action from the original MouseEvent's
// `metaKey || ctrlKey` so a fast click that arrives before the keydown
// listener fires still picks up the held key.
const _modifierHeld = writable(false);

if (typeof window !== "undefined") {
  window.addEventListener("keydown", (e) => {
    if (e.metaKey || e.ctrlKey) _modifierHeld.set(true);
  });
  window.addEventListener("keyup", (e) => {
    if (!e.metaKey && !e.ctrlKey) _modifierHeld.set(false);
  });
  window.addEventListener("blur", () => _modifierHeld.set(false));
}

export const modifierHeld = _modifierHeld as Readable<boolean>;
