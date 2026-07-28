import {
  createQueryServiceResolveCanvas,
  type V1CanvasSpec,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { SvelteMap } from "svelte/reactivity";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

export class CanvasSpecProvider {
  public canvasSpec: V1CanvasSpec | undefined = $state(undefined);
  public metricsViewSpecs: SvelteMap<string, V1MetricsViewSpec> = $state(
    new SvelteMap(),
  );
  public metricsViewNames: string[] = $state([]);

  public cleanup: () => void;

  public constructor(runtimeClient: RuntimeClient, canvasName: string) {
    const resolveCanvasQuery = createQueryServiceResolveCanvas(runtimeClient, {
      canvas: canvasName,
    });

    this.cleanup = resolveCanvasQuery.subscribe((resolveCanvasResp) => {
      this.canvasSpec =
        resolveCanvasResp.data?.canvas?.canvas?.state?.validSpec;
      this.metricsViewSpecs = new SvelteMap(
        Object.entries(
          resolveCanvasResp.data?.referencedMetricsViews ?? {},
        ).map(([mv, res]) => [mv, res.metricsView?.state?.validSpec ?? {}]),
      );
      this.metricsViewNames = [...this.metricsViewSpecs.keys()];
    });
  }
}
