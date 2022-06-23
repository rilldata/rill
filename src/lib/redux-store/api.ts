import * as reduxQuery from "@reduxjs/toolkit/query";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

const { createApi, fetchBaseQuery } = reduxQuery;

export const api = createApi({
  reducerPath: "api",
  baseQuery: fetchBaseQuery({
    // TODO: support hostname configuration
    baseUrl: "http://localhost:8080/api/",
  }),
  tagTypes: [
    EntityType.MetricsDefinition,
    EntityType.MeasureDefinition,
    EntityType.DimensionDefinition,
  ] as Array<EntityType>,
  endpoints: () => ({}),
});

export const defaultTransformFunction = (response) => response.data;
export const defaultProvidesFunction = (entityType: EntityType) => (result) =>
  result ? [{ type: entityType, id: result.id }] : [];
export const defaultArrayProvidesFunction =
  (entityType: EntityType) => (result) =>
    result
      ? [...result.map(({ id }) => ({ type: entityType, id })), entityType]
      : [];
export const defaultInvalidatesFunction =
  (entityType: EntityType) => (result, error, arg) =>
    [{ type: entityType, id: arg.id }];
