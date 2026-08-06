import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";

export const ADVANCED_FILTER_TAG = "advanced-filter";

export type AdvancedFilterMentionOption =
  | { type: "dimension"; dimension: MetricsViewSpecDimension }
  | { type: "measure"; measure: MetricsViewSpecMeasure };

export type AdvancedFilter = {
  sql: string;
  name: string;
};

export function convertAdvancedFilterToHTML(advancedFilter: AdvancedFilter) {
  return `<${ADVANCED_FILTER_TAG} name="${advancedFilter.name}">${advancedFilter.sql}</${ADVANCED_FILTER_TAG}>`;
}

const htmlTagStartRegex = new RegExp(
  `<${ADVANCED_FILTER_TAG} name=".*?">`,
  "g",
);
const htmlTagEndRegex = new RegExp(`</${ADVANCED_FILTER_TAG}>`, "g");

export function convertHTMLToSql(html: string) {
  return html.replace(htmlTagStartRegex, "").replace(htmlTagEndRegex, "");
}
