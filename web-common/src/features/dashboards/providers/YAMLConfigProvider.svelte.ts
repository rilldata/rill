import {
  createQueryServiceResolveCanvas,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

/**
 * A provider for YAML only configuration. These are only mutable during yaml editing.
 */
export class YAMLConfigProvider {
  public defaultFilters = $state<Record<string, V1Expression | undefined>>({});
  public pinnedFilters = $state<Record<string, boolean>>({});
  public specPinnedFilters = $state<Record<string, boolean>>({});
  public requiredFilters = $state<Record<string, boolean>>({});
  public specRequiredFilters = $state<Record<string, boolean>>({});

  public cleanup: (() => void) | undefined = undefined;

  public constructor(public readonly editable: boolean = false) {}

  public togglePinnedFilter(filter: string) {
    this.pinnedFilters[filter] = !this.pinnedFilters[filter];
  }

  public toggleRequiredFilter(filter: string) {
    this.requiredFilters[filter] = !this.requiredFilters[filter];
  }
}

export class CanvasConfigProvider extends YAMLConfigProvider {
  public constructor(
    runtimeClient: RuntimeClient,
    canvasName: string,
    editable: boolean = false,
  ) {
    super(editable);

    const resolveCanvasQuery = createQueryServiceResolveCanvas(runtimeClient, {
      canvas: canvasName,
    });
    this.cleanup = resolveCanvasQuery.subscribe((resolveCanvasResp) => {
      const canvasSpec =
        resolveCanvasResp.data?.canvas?.canvas?.state?.validSpec ?? {};

      this.defaultFilters = Object.fromEntries(
        Object.entries(canvasSpec.defaultPreset?.filterExpr ?? {}).map(
          ([mv, sqlFilter]) => [mv, sqlFilter.expression],
        ),
      );

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
    });
  }
}
