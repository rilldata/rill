/**
 * 
 * state we need
 * metrics:Metric[]
 * dimensions:Dimension[]
 * 
 */

// FIXME: this should be part of the api spec, not here locally!
import yaml from 'js-yaml';

 import { getByID } from "../dataset/index.js";
 import type { ExploreConfiguration, DataModelerState, MetricsModel } from "../../lib/types"
 import { rollupQuery } from '../explore-api.js';
 import { guidGenerator } from "../../lib/util/guid.js";


 export function createExploreConfigurationActions(api) {
     return {
         createExploreConfiguration({ metricsModelID }) {
            return async (dispatch, getState) => {
                // I would love to generate a preview of all this stuff.
                const state = getState();
                const configID = guidGenerator();
                const metricsModel = getByID(state.metricsModels, metricsModelID) as MetricsModel;
                dispatch((draft:DataModelerState) => {
                    const config = {
                        modelID: metricsModelID,
                        name: metricsModel.name,
                        id: configID,
                        currentMetricLeaderboard: metricsModel?.parsedSpec?.metrics[0]?.name,
                        activeMetrics: metricsModel?.parsedSpec?.metrics,
                        activeDimensions: metricsModel?.parsedSpec?.dimensions,
                        selectedMetrics: [],
                        selectedDimensions: [],
                        preview: {}
                    }
                    draft.exploreConfigurations = [config];
                })
                dispatch(this.generateTimeseries({ id: configID }));
                
                metricsModel?.parsedSpec?.dimensions?.forEach((dimension) => {
                    // run the API call?
                    // get currentLeaderboard
                    const metric = metricsModel?.parsedSpec?.metrics[0];
                    api.getTopKAndCardinality(metricsModel.parsedSpec.table, dimension.field, `${metric.function}(${metric.field})`)
                    //api.generateExploreLeaderboard({
                    //     table: metricsModel?.parsedSpec?.table,
                    //     leaderboardMetric: 'count(*)',
                    //     dimension: dimension.field,
                    //     timeField: metricsModel?.parsedSpec?.timeField
                    // })
                        .then(summary => {
                            dispatch((draft:DataModelerState) => {
                                const config = getByID(draft.exploreConfigurations, configID) as ExploreConfiguration;
                                if (!('dimensionBoard' in config.preview)) {
                                    config.preview.dimensionBoard = {};
                                }
                                config.preview.dimensionBoard[dimension.field] = summary;
                            })
                        })
                })
            }
        },

        generateTimeseries({ id, }) {
            // this needs to be an explore configuration id.
            return async (dispatch, getState) => {
                // get configuration first.
                const state = getState();
                const thisExploreConfiguration =  getByID(state.exploreConfigurations, id) as ExploreConfiguration;
                const underlyingMetricsModel = getByID(state.metricsModels, thisExploreConfiguration.modelID) as MetricsModel;

                // let's look at the selected metrics?
                api.generateExploreTimeseries({
                    table: underlyingMetricsModel?.parsedSpec?.table,
                    // FIXME: is this right?
                    metrics: thisExploreConfiguration.activeMetrics, 
                    dimensions: thisExploreConfiguration.selectedDimensions,
                    timeField: underlyingMetricsModel.parsedSpec.timeField, 
                    timeGrain: underlyingMetricsModel.parsedSpec.timeGrain
                }).then((timeSeries) => {
                    dispatch((draft:DataModelerState) => {
                        const exploreConfiguration = getByID(draft.exploreConfigurations, id) as ExploreConfiguration;
                        if (!('preview' in exploreConfiguration)) {
                            exploreConfiguration.preview = {};
                        }
                        exploreConfiguration.preview.timeSeries = timeSeries;
                    })
                })
            }
        },

        deleteExploreConfiguration({ id }) {
            return ((draft:DataModelerState) => {
                draft.exploreConfigurations = draft.exploreConfigurations.filter(config => config.id !== id);
            })
        }
    }
}


