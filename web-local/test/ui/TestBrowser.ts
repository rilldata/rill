import {
  afterAll,
  afterEach,
  beforeAll,
  beforeEach,
  expect,
} from "@jest/globals";
import { expect as playwrightExpect } from "@playwright/test";
import { asyncWait } from "@rilldata/web-local/common/utils/waitUtils";
import path from "node:path";
import { Browser, chromium, Page } from "playwright";

/**
 * Browser interaction abstraction.
 * Has our app specific actions like uploadFile and updateModelSql
 *
 * Use {@link TestBrowser.useTestBrowser} to add hooks in the topmost `describe`.
 */
export class TestBrowser {
  public page: Page;
  private browser: Browser;

  private constructor(
    private readonly testDataPath: string,
    private readonly appUrl: string
  ) {}

  /**
   * Adds the hooks to create browser for the suite and page per test.
   * @param testDataPath Path of test data. All file interactions are relative to this
   * @param appUrl Base url for the UI
   * @returns Instance of {@link TestBrowser} to be used in tests
   */
  public static useTestBrowser(testDataPath: string, appUrl: string) {
    const testBrowser = new TestBrowser(testDataPath, appUrl);

    beforeAll(async () => {
      testBrowser.browser = await chromium.launch({
        headless: false,
        devtools: true,
      });
    });

    beforeEach(async () => {
      testBrowser.page = await testBrowser.browser.newPage();
      await testBrowser.page.goto(appUrl);
    });

    afterEach(() => {
      return testBrowser.page.close();
    });

    afterAll(() => {
      return testBrowser.browser?.close();
    });

    return testBrowser;
  }

  // source action helpers

  /**
   * Used to upload local file as a source
   * @param file File name relative to test data folder
   * @param isDuplicate
   * @param keepBoth
   */
  public async uploadFile(file: string, isDuplicate = false, keepBoth = false) {
    // add table button
    await this.page.locator("button#add-table").click();
    // click local file tab
    await this.page
      .locator(".portal [slot='title'] button:nth-child(4)")
      .click();
    // wait for file chooser while clicking on upload button
    const [fileChooser] = await Promise.all([
      this.page.waitForEvent("filechooser"),
      this.page.locator(".portal .flex-grow .grid button").click(),
    ]);
    // input the `file` after joining with `testDataPath`
    await fileChooser.setFiles([path.join(this.testDataPath, file)]);

    // TODO: infer duplicate
    if (isDuplicate) {
      if (keepBoth) {
        // click on `Keep Both` if `isDuplicate`=true and `keepBoth`=true
        await this.clickModalButton("Keep Both");
      } else {
        // else click on `Replace Existing Source`
        await this.clickModalButton("Replace Existing Source");
      }
    } else {
      // if not duplicate wait and make sure `Duplicate source name` modal is not open
      await asyncWait(100);
      await playwrightExpect(
        this.page.locator(".portal h1", {
          hasText: "Duplicate source name",
        })
      ).toBeHidden();
    }

    await asyncWait(100);
  }

  public async createOrReplaceSource(file: string, name: string) {
    try {
      await this.page.locator(`a[href='/source/${name}']`).waitFor({
        timeout: 100,
      });
      await this.uploadFile(file, true, false);
    } catch (err) {
      await this.uploadFile(file);
    }
    await this.waitForEntity("source", name, true);
  }

  // model action helpers

  public async createModel(name: string) {
    // add model button
    await this.page.locator("button#create-model-button").click();
    await this.waitForEntity("model", "model", true);
    await this.renameEntityUsingTitle(name);
    await this.waitForEntity("model", name, true);
  }

  public async createModelFromSource(source: string) {
    await this.openEntityMenu("source", source);
    await this.clickMenuButton("create new model");
  }

  public async updateModelSql(sql: string) {
    await this.page.locator(".cm-line").first().click();
    // TODO: make this platform independent
    await this.page.keyboard.press("Meta+A");
    await this.page.keyboard.press("Delete");
    await this.page.keyboard.insertText(sql);
  }

  // common action helpers

  public async renameEntityUsingMenu(
    type: string,
    name: string,
    toName: string
  ) {
    // open context menu and click rename
    await this.openEntityMenu(type, name);
    await this.clickMenuButton("rename");

    // wait for rename modal to open
    await this.page
      .locator(".portal h1", {
        hasText: "Rename",
      })
      .waitFor();

    // type new name and submit
    await this.page.locator(".portal input").fill(toName);
    await this.clickModalButton("Change Name");
  }

  public async renameEntityUsingTitle(toName: string) {
    await this.page.locator("#model-title-input").fill(toName);
    await this.page.keyboard.press("Enter");
  }

  public async deleteEntity(type: string, name: string) {
    // open context menu and click rename
    await this.openEntityMenu(type, name);
    await this.clickMenuButton("delete");
  }

  public async modelHasError(hasError: boolean, error = "") {
    try {
      const errorLocator = this.page.locator(".editor-pane .error");
      await errorLocator.waitFor({
        timeout: 100,
      });
      expect(hasError).toBeTruthy();
      const actualError = await errorLocator.textContent();
      expect(actualError).toMatch(error);
    } catch (err) {
      expect(hasError).toBeFalsy();
    }
  }

  // wait helpers

  public async waitForEntity(type: string, name: string, navigated: boolean) {
    await this.page.locator(`a[href='/${type}/${name}']`).waitFor();
    if (navigated) {
      await this.page.locator("input#model-title-input").waitFor();
      await this.page.waitForURL(`${this.appUrl}/${type}/${name}`);
    }
  }

  public async entityNotPresent(type: string, name: string) {
    await asyncWait(100);
    await playwrightExpect(
      this.page.locator(`a[href='/${type}/${name}']`)
    ).toBeHidden();
  }

  private async openEntityMenu(type: string, name: string) {
    const entityLocator = this.page.locator(`a[href='/${type}/${name}']`);
    await entityLocator.hover();
    await this.page
      // get the navigation entry for the entity
      .locator(".navigation-entry-title", {
        has: entityLocator,
      })
      .locator("div.contents div.contents button")
      .click();
  }

  private async clickModalButton(text: string) {
    return this.page
      .locator(".portal button", {
        hasText: text,
      })
      .click();
  }

  private async clickMenuButton(text: string) {
    await this.page
      .locator(".portal button[role='menuitem'] div.text-left div", {
        hasText: new RegExp(text),
      })
      .click();
  }
}
