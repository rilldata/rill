import { describe, expect, it } from "@jest/globals";
import {
  initBlankDashboardYAML,
  MetricsInternalRepresentation,
} from "@rilldata/web-common/features/metrics-views/metrics-internal-store";

function createEmptyRepresentation() {
  const internalRepresentation = new MetricsInternalRepresentation(
    initBlankDashboardYAML("AdBids"),
    () => {
      // no-op
    }
  );
  internalRepresentation.bindStore(() => {
    // no-op
  });
  return internalRepresentation;
}

// TODO: add more exhaustive tests
describe("Metrics Internal Store", () => {
  it("Add remove dimensions", () => {
    const internalRepresentation = createEmptyRepresentation();

    internalRepresentation.addNewDimension();
    expect(internalRepresentation.internalYAML)
      .toEqual(`# Visit https://docs.rilldata.com/references/project-files to learn more about Rill project files.

display_name: "AdBids"
model: ""
default_time_range: ""
smallest_time_grain: ""
timeseries: ""
measures: []
dimensions:
  - label: ""
    property: ""
    description: ""
`);
    internalRepresentation.updateDimension(0, "label", "Publisher");
    internalRepresentation.updateDimension(0, "property", "publisher");
    expect(internalRepresentation.internalYAML)
      .toEqual(`# Visit https://docs.rilldata.com/references/project-files to learn more about Rill project files.

display_name: "AdBids"
model: ""
default_time_range: ""
smallest_time_grain: ""
timeseries: ""
measures: []
dimensions:
  - label: Publisher
    property: publisher
    description: ""
`);
  });

  it("Add remove measures", () => {
    const internalRepresentation = createEmptyRepresentation();

    internalRepresentation.addNewMeasure();
    expect(internalRepresentation.internalYAML)
      .toEqual(`# Visit https://docs.rilldata.com/references/project-files to learn more about Rill project files.

display_name: "AdBids"
model: ""
default_time_range: ""
smallest_time_grain: ""
timeseries: ""
measures:
  - label: ""
    expression: ""
    name: measure
    description: ""
    format_preset: humanize
dimensions: []
`);

    internalRepresentation.updateMeasure(0, "expression", "count(*)");
    internalRepresentation.updateMeasure(0, "name", "impressions");
    internalRepresentation.updateMeasure(0, "label", "Total Impressions");
    expect(internalRepresentation.internalYAML)
      .toEqual(`# Visit https://docs.rilldata.com/references/project-files to learn more about Rill project files.

display_name: "AdBids"
model: ""
default_time_range: ""
smallest_time_grain: ""
timeseries: ""
measures:
  - label: Total Impressions
    expression: count(*)
    name: impressions
    description: ""
    format_preset: humanize
dimensions: []
`);

    internalRepresentation.addNewMeasure();
    expect(internalRepresentation.internalYAML)
      .toEqual(`# Visit https://docs.rilldata.com/references/project-files to learn more about Rill project files.

display_name: "AdBids"
model: ""
default_time_range: ""
smallest_time_grain: ""
timeseries: ""
measures:
  - label: Total Impressions
    expression: count(*)
    name: impressions
    description: ""
    format_preset: humanize
  - label: ""
    expression: ""
    name: measure
    description: ""
    format_preset: humanize
dimensions: []
`);

    internalRepresentation.addNewMeasure();
    expect(internalRepresentation.internalYAML)
      .toEqual(`# Visit https://docs.rilldata.com/references/project-files to learn more about Rill project files.

display_name: "AdBids"
model: ""
default_time_range: ""
smallest_time_grain: ""
timeseries: ""
measures:
  - label: Total Impressions
    expression: count(*)
    name: impressions
    description: ""
    format_preset: humanize
  - label: ""
    expression: ""
    name: measure
    description: ""
    format_preset: humanize
  - label: ""
    expression: ""
    name: measure_1
    description: ""
    format_preset: humanize
dimensions: []
`);

    internalRepresentation.updateMeasure(1, "name", "measure_2");
    internalRepresentation.addNewMeasure();
    expect(internalRepresentation.internalYAML)
      .toEqual(`# Visit https://docs.rilldata.com/references/project-files to learn more about Rill project files.

display_name: "AdBids"
model: ""
default_time_range: ""
smallest_time_grain: ""
timeseries: ""
measures:
  - label: Total Impressions
    expression: count(*)
    name: impressions
    description: ""
    format_preset: humanize
  - label: ""
    expression: ""
    name: measure_2
    description: ""
    format_preset: humanize
  - label: ""
    expression: ""
    name: measure_1
    description: ""
    format_preset: humanize
  - label: ""
    expression: ""
    name: measure
    description: ""
    format_preset: humanize
dimensions: []
`);
  });
});
