import { DataProviderData, TestBase } from "@adityahegde/typescript-test-utils";
import { JestTestLibrary } from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import { PlaywrightSuiteSetup } from "@adityahegde/typescript-test-utils/dist/playwright/PlaywrightSuiteSetup";
import { expect, Page, PlaywrightTestArgs } from "@playwright/test";
import {
  TestServerSetup,
  TestServerSetupParameter,
} from "../utils/ServerSetup";

const PORT = 8080;
const URL = `http://localhost:${PORT}/`;

type CostOutput = string;
type ErrorOutput = string;

@TestBase.ParameterizedSuite([
  {
    cliFolder: "temp/test-ui",
    serverPort: PORT,
    uiPort: 3000,
  } as TestServerSetupParameter,
])
@TestBase.TestLibrary(JestTestLibrary)
@TestBase.TestSuiteSetup(TestServerSetup)
@TestBase.TestSuiteSetup(PlaywrightSuiteSetup)
export class DataModelerTest extends TestBase<TestServerSetupParameter> {
  public queryDataProvider(): DataProviderData<[string, CostOutput]> {
    type Args = [string, CostOutput];

    const LimitZeroQuery = "SELECT * FROM AdBids LIMIT 0";
    // FIXME the first query loaded doesn't seem to show the cost correctly, this could be a bug in the
    // UI.
    const LimitZeroQueryResult: CostOutput = encodeURI("0 rows, 5 columns");
    const LimitZeroQueryTestData: Args = [LimitZeroQuery, LimitZeroQueryResult];

    const LimitTenQuery = "SELECT * FROM AdBids LIMIT 10";
    const LimitTenQueryResult: CostOutput = encodeURI("10 rows, 5 columns");
    const LimitTenQueryTestData: Args = [LimitTenQuery, LimitTenQueryResult];

    const LimitRandomQuery = "SELECT * FROM AdBids LIMIT 1234";
    const LimitRandomQueryResult: CostOutput = encodeURI(
      "1,234 rows, 5 columns"
    );
    const LimitRandomQueryTestData: Args = [
      LimitRandomQuery,
      LimitRandomQueryResult,
    ];

    const FullSelectQuery = "SELECT * FROM AdBids";
    const FullSelectResult: CostOutput = encodeURI("10,000 rows, 5 columns");
    const FullSelectQueryTestData: Args = [FullSelectQuery, FullSelectResult];

    return {
      subData: [
        {
          title: "Limit 0",
          args: LimitZeroQueryTestData,
        },
        {
          title: "Limit 10",
          args: LimitTenQueryTestData,
        },
        {
          title: "Limit random",
          args: LimitRandomQueryTestData,
        },
        {
          title: "Full Select",
          args: FullSelectQueryTestData,
        },
      ],
    };
  }

  public errorDataProvider(): DataProviderData<[string, ErrorOutput]> {
    type Args = [string, ErrorOutput];

    const CatalogErrorQuery = "SELECT * FROM xyz";
    const CatalogErrorResult: ErrorOutput = `Catalog Error: Table with name xyz does not exist!`;
    const CatalogErrorTestData: Args = [CatalogErrorQuery, CatalogErrorResult];

    const ParserErrorQuery = "SELECT FROM AdBids";
    const ParserErrorResult =
      "Parser Error: SELECT clause without selection list";
    const ParserErrorTestData: Args = [ParserErrorQuery, ParserErrorResult];

    return {
      subData: [
        {
          title: "Catalog Error",
          args: CatalogErrorTestData,
        },
        {
          title: "Parser Error",
          args: ParserErrorTestData,
        },
      ],
    };
  }

  @TestBase.Test()
  public async testActiveInitializeModel({ page }: PlaywrightTestArgs) {
    await page.goto(URL);
    const defaultActiveModel = page.locator(
      "#assets-model-list .collapsible-table-summary-title"
    );
    const count = await defaultActiveModel.count();

    // we start with one model.
    expect(count).toBe(1);
  }

  // @TestBase.Test('queryDataProvider')
  // public async testCostEstimates(query: string, result: string, { page }: PlaywrightTestArgs) {
  //   await page.goto(URL);

  //   await this.execute(page, query);

  //   const cost = page.locator('.cost-estimate');

  //   await this.execute(page, query);
  //   const actualCost = encodeURI(await cost.textContent());
  //   const expectedCost = result;
  //   expect(actualCost).toEqual(expectedCost);
  // }

  @TestBase.Test()
  public async testNewModelCreation({ page }: PlaywrightTestArgs) {
    await page.goto(URL);

    const oldModelCount = await page
      .locator("#assets-model-list > div")
      .count();

    await page.click("button#create-model-button");

    await this.delay(300);

    const newModelCount = await page
      .locator("#assets-model-list > div")
      .count();
    expect(newModelCount).toBe(oldModelCount + 1);

    // get the text of the last model and compare to the title element in the workspace.
    const modelName = (
      await page.locator("#assets-model-list > div").last().textContent()
    ).replace(/\s/g, "");

    // check the modelName against the model title input element.
    const modelTitleElement = await page.inputValue("input#model-title-input");
    expect(modelName.includes(modelTitleElement)).toBeTruthy();
  }

  @TestBase.Test("errorDataProvider")
  public async testInvalidSql(
    query: string,
    result: string,
    { page }: PlaywrightTestArgs
  ) {
    await page.goto(URL);

    const error = page.locator(".error").first();

    await this.execute(page, query);

    const actualError = await error.textContent();
    expect(actualError).toContain(result);
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
    return new Promise((resolve) => {
      setTimeout(resolve, time);
    });
  }

  /**
   * Execute SQL using the Rill Developer UI.
   *
   * This simulates user input by changing the value of a `contenteditable` DIV provided by CodeMirror.
   * TODO: Simulate a virtual keyboard: https://playwright.dev/docs/api/class-keyboard
   *
   * @param page {Page} - Loaded Rill Developer page.
   * @param sql {string} - SQL to execute.
   */
  private async execute(page: Page, sql: string) {
    const activeLine = page.locator(".cm-activeLine").first();

    await activeLine.fill(sql);

    // FIXME it would be better to get a signal from the UI that it has completed, than
    // to wait an arbitrary amount of time like this.
    await this.delay(500);
  }
}
