import { rillDev } from "@rilldata/web-common/tests/fixtures/rill-dev-fixtures";

export const test = rillDev.extend({
  page: async ({ rillDevPage }, use) => {
    await use(rillDevPage);
  },
});
