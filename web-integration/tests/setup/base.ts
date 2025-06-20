import { mergeTests } from "playwright/test";
import { rillCloud } from "web-integration/tests/fixtures/rill-cloud-fixtures";
import { rillDev } from "web-integration/tests/fixtures/rill-dev-fixtures";

export const test = mergeTests(rillDev, rillCloud);
