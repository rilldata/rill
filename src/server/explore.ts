import { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { CATEGORICALS } from "$lib/duckdb-data-types";

import { ActionQueueOrchestrator } from "$common/priority-action-queue/ActionQueueOrchestrator";



function getAvailableDimensions({
  // when this is a table, the dimensions
  // are just the varchar values.
  entityType,
  entityID
}) {
  if (entityType === EntityType.Table) {
    // this is where we return something?
    const dimensions = this.dataModelerStateService.getEntityById(
      EntityType.Table,
      StateType.Derived,
      entityID
    ).profile.filter(column => CATEGORICALS.has(column.type))
      .map(column => column.name);
    // clear queue
    this.queue.clearQueue();
    this.socket.emit('getAvailableDimensions', { dimensions });
  } else {
    // ????
  }
}

async function getBigNumber({
  filters, entityType,
  entityID, expression = 'count(*)'
}) {
  const table = this.dataModelerStateService.getEntityById(
    EntityType.Table,
    StateType.Persistent,
    entityID
  );

  const [output] = await this.queue.enqueue({ priority: 0, id: 'ok' }, "bigNumber",
    [table.tableName, expression, filters]);

  this.socket.emit("getBigNumber", { value: output.value, metric: expression, filters })

}


async function getDimensionLeaderboard({
  dimensionName,
  filters,
  expression = 'count(*)',
  timeScaleHeuristic = undefined,
  // table or model?
  entityType,
  // the entity ID.
  entityID
}) {
  // whew, ok. Let's see if we can make this happen.
  // get the table name from this entity type / id.
  // ugh, just compute for all values. Who cares. This is a demo!

  const table = this.dataModelerStateService.getEntityById(
    EntityType.Table,
    StateType.Persistent,
    entityID
  );
  if (!table) return;
  const value = await this.queue.enqueue({ priority: 1, id: 'ok' }, "leaderboard",
    [table.tableName, dimensionName, expression, filters]);
  // this whole thing sucks so bad.
  this.socket.emit("getDimensionLeaderboard", { dimensionName, values: value, filters })
}

function dimensionSelectionsToFilterPredicates(selections) {
  return Object.keys(selections).map(field => {
    return selections[field].map(([value, filterType]) => {
      if (filterType === 'include') return `"${field}" = '${value}'`
    }).join(' OR ')
  }).join(' AND ')
}

/** This is the actionService for the priority queue orchetrator */
const queryMap = {
  leaderboard(db, [table, column, expression, predicates]) {
    // remove predicates for this specific dimension.
    const filteredPredicates = { ...predicates };
    delete filteredPredicates[column];
    const whereClause = predicates && Object.keys(filteredPredicates).length ? `AND ${dimensionSelectionsToFilterPredicates(filteredPredicates)}` : '';
    console.log(
      `
      SELECT ${expression} as value, "${column}" as label from "${table}"
      WHERE "${column}" IS NOT NULL ${whereClause}
      GROUP BY "${column}"
      ORDER BY value desc
      LIMIT 15
    `
    )
    return db.execute(`
      SELECT ${expression} as value, "${column}" as label from "${table}"
      WHERE "${column}" IS NOT NULL ${whereClause}
      GROUP BY "${column}"
      ORDER BY value desc
      LIMIT 15
    `)
  },
  bigNumber(db, [table, expression, predicates]) {
    const whereClause = predicates && Object.keys(predicates).length ? `WHERE ${dimensionSelectionsToFilterPredicates(predicates)}` : '';
    console.log(`
    SELECT ${expression} as value from "${table}"
    ${whereClause};
`)
    return db.execute(`
      SELECT ${expression} as value from "${table}"
      ${whereClause};
  `)
  }
}

const exploreAPI = [
  getAvailableDimensions,
  getDimensionLeaderboard,
  getBigNumber
]





export function initializeExploreSocketEndpoints(socket, dataModelerService, dataModelerStateService) {
  /** add all the explore API endpoints. Let's bind the dataModelerService to the functions
   * so that we can call this.datamodelerService & other things to tap into the existing server
   * infrastructure.
   */
  /** FIXME: why is it so hard to get access to the action queue? */
  const db = dataModelerService.databaseService.databaseClient;
  const exploreQueries = {
    dispatch(action, args) {
      return queryMap[action](db, args);
    }
  }
  const queue = new ActionQueueOrchestrator(exploreQueries);

  exploreAPI.forEach((api) => {
    socket.on(api.name, api.bind({ dataModelerService, dataModelerStateService, socket, queue, db }))
  });
}