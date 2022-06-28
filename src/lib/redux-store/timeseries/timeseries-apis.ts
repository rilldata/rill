import type { ActiveValues } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import * as reduxToolkit from "@reduxjs/toolkit";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { streamingFetchWrapper } from "$lib/util/fetchWrapper";
import type { TimeSeriesResponse } from "$common/database-service/DatabaseTimeSeriesActions";
import { updateTimeSeries } from "$lib/redux-store/timeseries/timeseries-slice";

const { createAsyncThunk } = reduxToolkit;

export const generateTimeSeriesApi = createAsyncThunk(
  `${EntityType.MetricsLeaderboard}/generateTimeSeries`,
  async (
    {
      metricsDefId,
      measures,
      filters,
      pixels,
    }: {
      metricsDefId: string;
      measures: Array<MeasureDefinitionEntity>;
      filters?: ActiveValues;
      pixels?: number;
    },
    thunkAPI
  ) => {
    const stream = streamingFetchWrapper<TimeSeriesResponse>(
      `metrics/${metricsDefId}/time-series`,
      "POST",
      {
        expressionEntries: measures.map((measure) => [
          measure.id,
          measure.expression,
        ]),
        filters,
        pixels,
      }
    );
    for await (const timeSeriesResponse of stream) {
      thunkAPI.dispatch(
        updateTimeSeries(timeSeriesResponse.id, {
          values: timeSeriesResponse.results,
          rollupInterval: timeSeriesResponse.rollupInterval,
          spark: timeSeriesResponse.spark,
        })
      );
    }
  }
);
