// An ad-hoc "ephemeral measure" defined by the user for an explore dashboard,
// derived from existing metrics view measures via an arithmetic expression
// (e.g. Profit = revenue - cost). It is computed server-side via the metrics
// APIs' `expression` measure compute, which restricts expressions to
// references to existing measures, numeric literals, basic arithmetic and a
// small allowlist of functions. Definitions live on the explore state, are
// shared by all views (leaderboards, charts, pivot, ...) and are encoded in
// the `ephemeral` URL param so shared links reproduce them.
export type EphemeralMeasureDef = {
  // Query alias, e.g. "profit". Must not collide with metrics view field names.
  name: string;
  // Display name shown in the UI, e.g. "Profit".
  displayName: string;
  // Arithmetic expression over existing measure names, e.g. "revenue - cost".
  expression: string;
  // Optional FormatPreset for rendering values; defaults to humanize.
  formatPreset?: string;
};
