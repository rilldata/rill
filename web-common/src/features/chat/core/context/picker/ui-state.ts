import { writable, derived, get } from "svelte/store";

export class ContextPickerUIState {
  public expandedParentsStore = writable({} as Record<string, boolean>);

  public getExpandedStore(id: string) {
    return derived(this.expandedParentsStore, (expandedParents) =>
      Boolean(expandedParents[id]),
    );
  }

  public isExpanded(id: string) {
    return get(this.expandedParentsStore)[id] ?? false;
  }

  public expand(id: string) {
    this.expandedParentsStore.update((expandedParents) => {
      expandedParents[id] = true;
      return expandedParents;
    });
  }

  public collapse(id: string) {
    this.expandedParentsStore.update((expandedParents) => {
      expandedParents[id] = false;
      return expandedParents;
    });
  }

  public toggle(id: string) {
    this.expandedParentsStore.update((expandedParents) => {
      expandedParents[id] = !expandedParents[id];
      return expandedParents;
    });
  }
}
