import { PostgresTestContainer } from "./postgres.ts";

const pg = new PostgresTestContainer();
void (async () => {
  await pg.start();
  await pg.seedAdBids();
  await pg.seedAdImpressions();
  await new Promise((resolve) => setTimeout(resolve, 10000));
  await pg.stop();
})();
