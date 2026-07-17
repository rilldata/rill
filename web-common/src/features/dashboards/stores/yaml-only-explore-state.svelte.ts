import {
  ArrayRuneStore,
  InMemoryRuneStore,
} from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";

export interface YAMLOnlyExploreState {
  editable: boolean;
  pinnedFilters: ArrayRuneStore<string>;
  requiredFilters: ArrayRuneStore<string>;
}

export class ReadonlyYAMLExploreState implements YAMLOnlyExploreState {
  public readonly editable = false;
  public pinnedFilters: ArrayRuneStore<string>;
  public requiredFilters: ArrayRuneStore<string>;

  public constructor() {
    this.pinnedFilters = new ArrayRuneStore(new InMemoryRuneStore([]));
    this.requiredFilters = new ArrayRuneStore(new InMemoryRuneStore([]));
  }

  public sync(exploreSpec: V1ExploreSpec) {
    this.pinnedFilters.setter(exploreSpec.defaultPreset?.pinnedFilters ?? []);
    this.requiredFilters.setter(
      exploreSpec.defaultPreset?.requiredFilters ?? [],
    );
  }
}
