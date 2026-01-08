import { writable, derived, get } from "svelte/store";

export class ContextPickerUIState {
  private expandedParentsStore = writable({} as Record<string, boolean>);

  public getExpandedStore(parentKey: string) {
    return derived(this.expandedParentsStore, (expandedParents) =>
      Boolean(expandedParents[parentKey]),
    );
  }

  public isExpanded(parentKey: string) {
    return get(this.expandedParentsStore)[parentKey] ?? false;
  }

  public expand(parentKey: string) {
    this.expandedParentsStore.update((expandedParents) => {
      expandedParents[parentKey] = true;
      return expandedParents;
    });
  }

  public collapse(parentKey: string) {
    this.expandedParentsStore.update((expandedParents) => {
      expandedParents[parentKey] = false;
      return expandedParents;
    });
  }

  public toggle(parentKey: string) {
    this.expandedParentsStore.update((expandedParents) => {
      expandedParents[parentKey] = !expandedParents[parentKey];
      return expandedParents;
    });
  }
}
