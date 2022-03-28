import { expect, Page, PlaywrightTestArgs } from "@playwright/test";

import { exec } from "node:child_process";
import { promisify } from "util"
import terminate from "terminate/promise";

import { JestTestLibrary } from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import { PlaywrightSuiteSetup } from "@adityahegde/typescript-test-utils/dist/playwright/PlaywrightSuiteSetup";
import { DataProviderData, TestBase, TestSuiteSetup, TestSuiteParameter } from "@adityahegde/typescript-test-utils";
import { waitUntil } from "$common/utils/waitUtils";
import { isPortOpen } from "$common/utils/isPortOpen";

const execPromise = promisify(exec);

const PORT = 8080;
const DEV_PORT = 3000;
const URL = `http://localhost:${PORT}/`;
const CLI_TEST_FOLDER = 'temp/test-ui';
const DATA_MODELER_CLI = './node_modules/.bin/ts-node-dev --project tsconfig.node.json -- src/cli/data-modeler-cli.ts';
const CLI_TEST_FOLDER_ARG = `--project ${CLI_TEST_FOLDER}`;

let serverStarted = false;

class ServerSetup extends TestSuiteSetup {
  child: import("child_process").ChildProcess;

  setupTest(testSuiteParameter: TestSuiteParameter, testContext: Record<any, any>): Promise<void> {
    return Promise.resolve();
  }
  teardownTest(testSuiteParameter: TestSuiteParameter, testContext: Record<any, any>): Promise<void> {
    return Promise.resolve();
  }
  async teardownSuite(testSuiteParameter: TestSuiteParameter): Promise<void> {
    await terminate(this.child.pid);
    return undefined;
  }

  public async setupSuite(testSuiteParameter: TestSuiteParameter): Promise<void> {
    // Test to see if server is already running on PORT.
    [PORT, DEV_PORT].forEach(async port => {
      if (await isPortOpen(port)) {
        console.error(`Cannot run UI tests, server is already running on ${port}`);
        process.exit(1);
      }
    });

    await execPromise(`mkdir -p ${CLI_TEST_FOLDER}`);
    await execPromise(`rm -rf ${CLI_TEST_FOLDER}/*`);

    await execPromise(`${DATA_MODELER_CLI} init ${CLI_TEST_FOLDER_ARG}`);
    await execPromise(`${DATA_MODELER_CLI} import-table  ${CLI_TEST_FOLDER_ARG} ./data/Users.parquet`);
    await execPromise(`${DATA_MODELER_CLI} import-table  ${CLI_TEST_FOLDER_ARG} ./data/AdImpressions.parquet`);
    await execPromise(`${DATA_MODELER_CLI} import-table  ${CLI_TEST_FOLDER_ARG} ./data/AdBids.parquet`);

    // Run data modeler in the background, logging to stdout.
    this.child = exec(`${DATA_MODELER_CLI} start ${CLI_TEST_FOLDER_ARG}`);
    this.child.stdout.pipe(process.stdout);
    // Watch for server startup in output.
    this.child.stdout.on('data', (data) => {
      if (data.startsWith('Server started at')) {
        serverStarted = true;
      }
    });
    // Terminate if the process exits.
    process.on('exit', async () => {
      await terminate(this.child.pid);
    });
  }
}

type CostOutput = string;
type ErrorOutput = string;

@TestBase.Suite
@TestBase.TestLibrary(JestTestLibrary)
@TestBase.TestSuiteSetup(ServerSetup)
@TestBase.TestSuiteSetup(PlaywrightSuiteSetup)

export class DataModelerTest extends TestBase {

  @TestBase.BeforeSuite()
  public async setup() {
    // Wait for server startup
    await waitUntil(() => serverStarted);
  }

  public queryDataProvider(): DataProviderData<[string, CostOutput]> {
    type Args = [string, CostOutput];

    const LimitZeroQuery = 'SELECT * FROM AdBids LIMIT 0';
    // FIXME the first query loaded doesn't seem to show the cost correctly, this could be a bug in the
    // UI.
    const LimitZeroQueryResult: CostOutput = encodeURI('0 rows, 5 columns');
    const LimitZeroQueryTestData: Args = [LimitZeroQuery, LimitZeroQueryResult];

    const LimitTenQuery = 'SELECT * FROM AdBids LIMIT 10';
    const LimitTenQueryResult: CostOutput = encodeURI('10 rows, 5 columns');
    const LimitTenQueryTestData: Args = [LimitTenQuery, LimitTenQueryResult];

    const LimitRandomQuery = 'SELECT * FROM AdBids LIMIT 1234';
    const LimitRandomQueryResult: CostOutput = encodeURI('1,234 rows, 5 columns');
    const LimitRandomQueryTestData: Args = [LimitRandomQuery, LimitRandomQueryResult];

    const FullSelectQuery = 'SELECT * FROM AdBids';
    const FullSelectResult: CostOutput = encodeURI('10,000 rows, 5 columns')
    const FullSelectQueryTestData: Args = [FullSelectQuery, FullSelectResult];

    return {
      subData: [{
        title: 'Limit 0',
        args: LimitZeroQueryTestData,
      }, {
        title: 'Limit 10',
        args: LimitTenQueryTestData,
      }, {
        title: 'Limit random',
        args: LimitRandomQueryTestData,
      }, {
        title: 'Full Select',
        args: FullSelectQueryTestData,
      }],
    }
  }

  public errorDataProvider(): DataProviderData<[string, ErrorOutput]> {
    type Args = [string, ErrorOutput];

    const CatalogErrorQuery = 'SELECT * FROM xyz';
    const CatalogErrorResult: ErrorOutput = `Catalog Error: Table with name xyz does not exist!
Did you mean \"Users\"?
LINE 1: SELECT * FROM xyz
                      ^`;
    const CatalogErrorTestData: Args = [CatalogErrorQuery, CatalogErrorResult];

    const ParserErrorQuery = 'SELECT FROM AdBids';
    const ParserErrorResult = 'Parser Error: SELECT clause without selection list';
    const ParserErrorTestData: Args = [ParserErrorQuery, ParserErrorResult];

    return {
      subData: [{
        title: 'Catalog Error',
        args: CatalogErrorTestData,
      }, {
        title: 'Parser Error',
        args: ParserErrorTestData,
      }],
    }
  }

  @TestBase.Test()
  public async testActiveInitializeModel({ page }: PlaywrightTestArgs) {
    await page.goto(URL);
    const defaultActiveModel = page.locator("#assets-model-list .collapsible-table-summary-title")
    const modelName = await defaultActiveModel.textContent();
    const count = await defaultActiveModel.count();
    
    // we start with one model.
    expect(count).toBe(1);

    // the model is the selected one.
    const modelTitleElement = await page.inputValue('input#model-title-input');
    expect(modelName.includes(modelTitleElement)).toBeTruthy();
  }

  @TestBase.Test('queryDataProvider')
  public async testCostEstimates(query: string, result: string, { page }: PlaywrightTestArgs) {
    await page.goto(URL);

    await this.execute(page, query);

    const cost = page.locator('.cost-estimate');

    await this.execute(page, query);
    const actualCost = encodeURI(await cost.textContent());
    const expectedCost = result;
    expect(actualCost).toEqual(expectedCost);
  }

  @TestBase.Test()
  public async testNewModelCreation({ page }: PlaywrightTestArgs) {
    await page.goto(URL);

    const oldModelCount = await page.locator('#assets-model-list > div').count();

    await page.click('button#create-model-button');

    await this.delay(300);

    const newModelCount = await page.locator("#assets-model-list > div").count();
    expect(newModelCount).toBe(oldModelCount + 1);

    // get the text of the last model and compare to the title element in the workspace.
    const modelName = await page.locator("#assets-model-list > div").last().textContent();

    // check the modelName against the model title input element.
    const modelTitleElement = await page.inputValue('input#model-title-input');
    expect(modelName.includes(modelTitleElement)).toBeTruthy();
  }

  @TestBase.Test('errorDataProvider')
  public async testInvalidSql(query: string, result: string, { page }: PlaywrightTestArgs) {
    await page.goto(URL);

    const error = page.locator('.error');

    await this.execute(page, query);

    const actualError = await error.textContent();
    expect(actualError).toEqual(result);
  }

  /**
   * Sleep timer. NOTE - waiting for an explicit signal is preferred,
   * but in cases where waiting for a specific amount of time to pass is appropriate,
   * this may be used.
   *
   * @param {number} time - time to sleep in milliseconds.
   * @returns {Promise} - resolves when timeout is reached.
   */
  private delay(time: number): Promise<any> {
    return new Promise(resolve => {
      setTimeout(resolve, time);
    });
  }

  /**
   * Execute SQL using the data modeler UI.
   *
   * This simulates user input by changing the value of a `contenteditable` DIV provided by CodeMirror.
   * TODO: Simulate a virtual keyboard: https://playwright.dev/docs/api/class-keyboard
   *
   * @param page {Page} - Loaded data modeler page.
   * @param sql {string} - SQL to execute.
   */
  private async execute(page: Page, sql: string) {
    const activeLine = page.locator('.cm-activeLine');

    // Click add model button, and select model, if editor is not already visible.
    // TODO this would be more future-proof if the UI added IDs for buttons.
    await page.locator('button#create-model-button').click();
    await page.locator('text=query_0.sql >> nth=1').click();

    await activeLine.fill(sql);

    // FIXME it would be better to get a signal from the UI that it has completed, than
    // to wait an arbitrary amount of time like this.
    await this.delay(500);
  }
}
