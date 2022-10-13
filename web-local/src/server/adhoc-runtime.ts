import "../moduleAlias";
import { DatabaseConfig } from "@rilldata/web-local/common/config/DatabaseConfig";
import { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import { DuckDBClient } from "@rilldata/web-local/common/database-service/DuckDBClient";

// TODO: figure out adhoc runtime and add tests in go
(async () => {
  const duckdb = DuckDBClient.getInstance(
    new RootConfig({
      database: new DatabaseConfig({
        spawnRuntime: false,
        spawnRuntimePort: 8080,
      }),
    })
  );
  await duckdb.init();
  await duckdb.execute("SET enable_profiling=QUERY_TREE");
  console.log(
    await duckdb.execute(
      "CREATE OR REPLACE TABLE AdBids AS (SELECT * FROM 'web-local/test/data/AdBids.csv')"
    )
  );
  console.log(
    await duckdb.requestToInstance("metrics-views/AdBids/timeseries", {
      metricsViewName: "AdBids",
      measureNames: ["count(*) as count"],
      timeStart: new Date("2022-01-01").getTime(),
      timeEnd: new Date("2022-03-01").getTime(),
      timeGranularity: "DAY",
    })
  );
  console.log(
    await duckdb.requestToInstance("metrics-views/AdBids/toplist/publisher", {
      metricsViewName: "AdBids",
      measureNames: ["count(*) as count"],
      timeStart: new Date("2022-01-01").getTime(),
      timeEnd: new Date("2022-03-01").getTime(),
      limit: 25,
      filter: {
        include: [
          {
            name: "domain",
            in: ["sports.yahoo.com"],
          },
        ],
        exclude: [],
      },
    })
  );
  console.log(
    await duckdb.requestToInstance("metrics-views/AdBids/toplist/publisher", {
      metricsViewName: "AdBids",
      measureNames: ["count(*) as count"],
      timeStart: new Date("2022-01-01").getTime(),
      timeEnd: new Date("2022-03-01").getTime(),
      limit: 25,
      filter: {
        include: [],
        exclude: [
          {
            name: "publisher",
            in: ["sports.yahoo.com"],
          },
        ],
      },
    })
  );
  console.log(
    await duckdb.requestToInstance("metrics-views/AdBids/toplist/publisher", {
      metricsViewName: "AdBids",
      measureNames: ["count(*) as count"],
      timeStart: new Date("2022-01-01").getTime(),
      timeEnd: new Date("2022-03-01").getTime(),
      limit: 25,
      filter: {
        include: [],
        exclude: [
          {
            name: "domain",
            like: ["%oo%"],
          },
        ],
      },
    })
  );
  console.log(
    await duckdb.requestToInstance("metrics-views/AdBids/totals", {
      metricsViewName: "AdBids",
      measureNames: ["count(*) as count"],
      timeStart: new Date("2022-01-01").getTime(),
      timeEnd: new Date("2022-03-01").getTime(),
      filter: {
        include: [
          {
            name: "domain",
            in: ["sports.yahoo.com"],
          },
        ],
        exclude: [],
      },
    })
  );
  console.log(
    await duckdb.requestToInstance("metrics-views/AdBids/totals", {
      metricsViewName: "AdBids",
      measureNames: ["count(*) as count"],
      timeStart: new Date("2022-01-01").getTime(),
      timeEnd: new Date("2022-03-01").getTime(),
      filter: {
        include: [],
        exclude: [
          {
            name: "publisher",
            in: ["sports.yahoo.com"],
          },
        ],
      },
    })
  );
  console.log(
    await duckdb.requestToInstance("metrics-views/AdBids/totals", {
      metricsViewName: "AdBids",
      measureNames: ["count(*) as count"],
      timeStart: new Date("2022-01-01").getTime(),
      timeEnd: new Date("2022-03-01").getTime(),
      filter: {
        include: [],
        exclude: [
          {
            name: "domain",
            like: ["%oo%"],
          },
        ],
      },
    })
  );
})();
