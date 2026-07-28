import {
  ArrayRuneStore,
  InMemoryRuneStore,
} from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";

export class YAMLOnlyExploreState {
  public pinnedFilters: ArrayRuneStore<string>;
  public requiredFilters: ArrayRuneStore<string>;

  public constructor(public readonly editable: boolean) {
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

export type YAMLOnlyExploreConfig = {
  pinnedFilters: string[];
  requiredFilters: string[];
};
