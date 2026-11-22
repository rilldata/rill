import type { View } from "svelte-vega";

/** The most fits a change of the plot's size is followed by before the chart is left as it is. */
const MAX_REFITS = 3;

/**
 * Keeps a chart's plot fitted around its axes and legends as they are drawn.
 *
 * Vega fits a plot to its view once per resize (`autosize: fit`): it lays the chart out with the
 * `width` and `height` signals holding the view's size, measures the room the axes and legends
 * take, shrinks the plot by that much, and lays the chart out again. That second layout is not
 * measured. Whatever is laid out from `width` or `height` changes in it: a legend whose column
 * count follows the plot's width wraps onto more rows, and a category axis turns its labels once
 * its bands narrow. Those guides then take more room than the plot left them, so a legend above
 * the plot lost its title and first rows past the top of the chart, and one below it, or an
 * axis's turned labels, ran past the bottom.
 *
 * So each time the plot's size changes, the chart is fitted once more, around the guides as that
 * size laid them out, until the size holds. A chart nothing of which follows its size takes one
 * extra layout pass per resize.
 */
export function refitOnPlotResize(view: View): void {
  // Refits since the plot's size last held.
  let refits = 0;
  let scheduled = false;

  const refit = () => {
    scheduled = false;
    if (refits >= MAX_REFITS) {
      // A layout that never settles is left where it is, until the next change of size.
      refits = 0;
      return;
    }
    refits++;
    const width = view.width();
    const height = view.height();
    // A fit that changes the plot's size sets the signals in a run of its own, which schedules
    // the next refit and runs it before this call returns.
    view.resize().run();
    if (view.width() === width && view.height() === height) refits = 0;
  };

  // The listeners are called while the dataflow runs, when it cannot be run again; the refit
  // waits for that run to end, and for Vega's own fit, which takes priority 1, to have applied.
  const schedule = () => {
    if (scheduled) return;
    scheduled = true;
    view.runAfter(refit, true);
  };
  view.addSignalListener("width", schedule);
  view.addSignalListener("height", schedule);
}
