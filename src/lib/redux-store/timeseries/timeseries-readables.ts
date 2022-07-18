import { createReadableFactoryWithSelector } from "$lib/redux-store/svelte-readables-wrapper";
import { store } from "$lib/redux-store/store-root";
import { selectTimeSeriesById } from "$lib/redux-store/timeseries/timeseries-selectors";

export const getTimeSeriesById = createReadableFactoryWithSelector(
  store,
  selectTimeSeriesById
);
