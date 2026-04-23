import { writable } from "svelte/store";

export type EditorMode = "code" | "visual";

const STORAGE_KEY = "rill:editor-mode";

function loadInitialMode(): EditorMode {
  if (typeof localStorage === "undefined") return "code";
  const stored = localStorage.getItem(STORAGE_KEY);
  return stored === "visual" ? "visual" : "code";
}

function createEditorModeStore() {
  const store = writable<EditorMode>(loadInitialMode());

  store.subscribe((value) => {
    if (typeof localStorage === "undefined") return;
    localStorage.setItem(STORAGE_KEY, value);
  });

  return store;
}

export const editorMode = createEditorModeStore();
