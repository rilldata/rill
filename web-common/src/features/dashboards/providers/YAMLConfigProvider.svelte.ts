import type { V1CanvasSpec } from "@rilldata/web-common/runtime-client";

/**
 * A provider for YAML only configuration. These are only mutable during yaml editing.
 */
export class YAMLConfigProvider {
  public pinnedFilters = $state<Record<string, boolean>>({});
  public specPinnedFilters = $state<Record<string, boolean>>({});
  public requiredFilters = $state<Record<string, boolean>>({});
  public specRequiredFilters = $state<Record<string, boolean>>({});

  public constructor(public readonly editable: boolean = false) {}

  public syncWithCanvas(canvasSpec: V1CanvasSpec) {
    const pinnedFilters = canvasSpec.pinnedFilters
      ? Object.fromEntries(
          canvasSpec.pinnedFilters.map((filter) => [filter, true]),
        )
      : {};
    this.pinnedFilters = { ...pinnedFilters };
    this.specPinnedFilters = { ...pinnedFilters };

    const requiredFilters = canvasSpec.requiredFilters
      ? Object.fromEntries(
          canvasSpec.requiredFilters.map((filter) => [filter, true]),
        )
      : {};
    this.requiredFilters = { ...requiredFilters };
    this.specRequiredFilters = { ...requiredFilters };
  }

  public togglePinnedFilter(filter: string) {
    this.pinnedFilters[filter] = !this.pinnedFilters[filter];
  }

  public toggleRequiredFilter(filter: string) {
    this.requiredFilters[filter] = !this.requiredFilters[filter];
  }
}
