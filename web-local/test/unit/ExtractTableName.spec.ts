import {
  extractFileExtension,
  extractTableName,
  sanitizeEntityName,
} from "@rilldata/web-local/lib/util/extract-table-name";
import { DataProviderData, TestBase } from "@adityahegde/typescript-test-utils";
import { JestTestLibrary } from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";

@TestBase.Suite
@TestBase.TestLibrary(JestTestLibrary)
export class ExtractTableNameSpec extends TestBase {
  public tablePathTestData(): DataProviderData<[string, [string, string]]> {
    const getVariations = (
      fileName,
      expectedFileName,
      expectedExtension = "parquet"
    ) => {
      const expectedFileAndExtension = [expectedFileName, expectedExtension];
      return {
        title: `fileName=${fileName}`,
        subData: [
          {
            args: [`path/to/file/${fileName}`, expectedFileAndExtension],
          },
          {
            args: [`/path/to/file/${fileName}`, expectedFileAndExtension],
          },
          {
            args: [`./path/to/file/${fileName}`, expectedFileAndExtension],
          },
          {
            args: [fileName, expectedFileAndExtension],
          },
          {
            args: [`/${fileName}`, expectedFileAndExtension],
          },
          {
            args: [`./${fileName}`, expectedFileAndExtension],
          },
        ],
      } as DataProviderData<[string, [string, string]]>;
    };

    return {
      subData: [
        getVariations("table.parquet", "table"),
        getVariations("table.v1.parquet", "table_v1"),
        getVariations("table", "table", ""),
      ],
    };
  }

  @TestBase.Test("tablePathTestData")
  public shouldExtractAnSanitiseTableName(
    tablePath: string,
    [expectedTableName]: [string, string]
  ) {
    expect(sanitizeEntityName(extractTableName(tablePath))).toBe(
      expectedTableName
    );
  }

  @TestBase.Test("tablePathTestData")
  public shouldExtractExtension(
    tablePath: string,
    [, expectedExtension]: [string, string]
  ) {
    expect(extractFileExtension(tablePath)).toBe(expectedExtension);
  }
}
