import { VegaSpec } from "svelte-vega";
import { compile, TopLevelSpec } from "vega-lite";
import { Signal } from "vega-typings";

export class VegaSignalManager {
  private compiledSpec: VegaSpec;

  constructor(private sanitizedVegaLiteSpec: TopLevelSpec) {
    this.compiledSpec = compile(this.sanitizedVegaLiteSpec).spec;
  }

  public updateVegaSpec() {
    const originalSignals = this.compiledSpec.signals || [];
    const updatedSignals = originalSignals.map(this.updateExistingSignals);

    updatedSignals.push(this.createBrushEndSignal());
    updatedSignals.push(this.createBrushClearSignal());

    return {
      ...this.compiledSpec,
      signals: updatedSignals,
    };
  }

  private updateExistingSignals = (signal: Signal): Signal => {
    switch (signal.name) {
      case "brush_x":
        return this.updateBrushXSignal(signal);
      case "brush_ts":
        return this.updateBrushTsSignal(signal);
      default:
        return signal;
    }
  };

  private updateBrushXSignal(signal: Signal): Signal {
    return {
      ...signal,
      value: [],
      on: [
        { events: { signal: "brush_clear" }, update: "[0, 0]" },
        ...(signal.on || []),
      ],
    };
  }

  private updateBrushTsSignal(signal: Signal): Signal {
    return {
      ...signal,
      on: [
        { events: { signal: "brush_clear" }, update: "null" },
        ...(signal.on || []),
      ],
    };
  }

  private createBrushEndSignal(): Signal {
    return {
      name: "brush_end",
      on: [
        {
          events: {
            source: "scope",
            type: "pointerup",
          },
          update: { signal: "brush" },
        },
        {
          events: {
            source: "scope",
            type: "pointerdown",
          },
          update: { signal: "brush" },
        },
        // When user pointerups outside of chart
        // TODO: revisit this
        // {
        //   events: {
        //     source: "window",
        //     type: "pointerup",
        //   },
        //   update: { signal: "brush" },
        // },
        // {
        //   events: {
        //     source: "window",
        //     type: "pointerdown",
        //   },
        //   update: { signal: "brush" },
        // },
      ],
    };
  }

  private createBrushClearSignal(): Signal {
    return {
      name: "brush_clear",
      on: [
        {
          events: {
            source: "window",
            type: "keydown",
            filter: ["event.key === 'Escape'"],
          },
          update: { signal: "brush" },
        },
      ],
    };
  }
}
