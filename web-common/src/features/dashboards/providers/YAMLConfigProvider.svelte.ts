import { type V1Expression } from "@rilldata/web-common/runtime-client";

/**
 * A provider for YAML only configuration. These are only mutable during yaml editing.
 */
export class YAMLConfigProvider {
  public defaultFilters = $state<Record<string, V1Expression | undefined>>({});
  public pinnedFilters = $state<Record<string, boolean>>({});
  public specPinnedFilters = $state<Record<string, boolean>>({});
  public requiredFilters = $state<Record<string, boolean>>({});
  public specRequiredFilters = $state<Record<string, boolean>>({});
  public editable: boolean = false;

  public cleanup: (() => void) | undefined = undefined;

  public update(
    defaultFilters: Record<string, V1Expression | undefined>,
    pinnedFilters: string[],
    requiredFilters: string[],
  ) {
    this.defaultFilters = defaultFilters;

    const pinnedFiltersRec = Object.fromEntries(
      pinnedFilters.map((filter) => [filter, true]),
    );
    this.pinnedFilters = { ...pinnedFiltersRec };
    this.specPinnedFilters = { ...pinnedFiltersRec };

    const requiredFiltersRec = Object.fromEntries(
      requiredFilters.map((filter) => [filter, true]),
    );
    this.requiredFilters = { ...requiredFiltersRec };
    this.specRequiredFilters = { ...requiredFiltersRec };
  }

  public setEditable(newEditable: boolean) {
    this.editable = newEditable;
  }

  public togglePinnedFilter(filter: string) {
    this.pinnedFilters[filter] = !this.pinnedFilters[filter];
  }

  public toggleRequiredFilter(filter: string) {
    this.requiredFilters[filter] = !this.requiredFilters[filter];
  }
}
