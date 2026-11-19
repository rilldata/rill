import type { View } from "svelte-vega";

/**
 * Narrows `datasets` to the ones `view` defines.
 *
 * svelte-vega pushes every entry of its `data` prop into the view with `view.change`, which
 * throws `Unrecognized data set: <name>` when the compiled spec has no data source of that
 * name. It does so inside an async function nobody awaits, so the chart keeps rendering while
 * the page gets an unhandled rejection on every data update. A spec that does not read a
 * dataset simply does not receive it.
 */
export function pickViewDatasets(
  view: View,
  datasets: Record<string, unknown>,
): Record<string, unknown> {
  const known: Record<string, unknown> = {};
  for (const [name, values] of Object.entries(datasets)) {
    if (hasDataset(view, name)) known[name] = values;
  }
  return known;
}

// Vega offers no way to ask whether a dataset exists; reading it throws when it does not.
function hasDataset(view: View, name: string): boolean {
  try {
    view.data(name);
    return true;
  } catch {
    return false;
  }
}
