import { generateBasicSelectors } from "$lib/redux-store/utils/selector-utils";

export const { singleSelector: selectTimeSeriesById } =
  generateBasicSelectors("timeSeries");
