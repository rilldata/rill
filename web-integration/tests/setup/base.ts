import { mergeTests } from "playwright/test";
import { rillCloud } from "@rilldata/web-common/tests/fixtures/rill-cloud-fixtures";
import { rillDev } from "@rilldata/web-common/tests/fixtures/rill-dev-fixtures";

export const test = mergeTests(rillDev, rillCloud);
