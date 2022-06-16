import { DataProviderData, TestBase } from "@adityahegde/typescript-test-utils";
import { JestTestLibrary } from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import {
  extractFileExtension,
  extractSourceName,
  sanitizeSourceName,
} from "$lib/util/extract-source-name";

@TestBase.Suite
@TestBase.TestLibrary(JestTestLibrary)
export class ExtractSourceNameSpec extends TestBase {
  public sourcePathTestData(): DataProviderData<[string, [string, string]]> {
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
        getVariations("source.parquet", "source"),
        getVariations("source.v1.parquet", "source_v1"),
        getVariations("source", "source", ""),
      ],
    };
  }

  @TestBase.Test("sourcePathTestData")
  public shouldExtractAnSanitiseSourceName(
    sourcePath: string,
    [expectedSourceName]: [string, string]
  ) {
    expect(sanitizeSourceName(extractSourceName(sourcePath))).toBe(
      expectedSourceName
    );
  }

  @TestBase.Test("sourcePathTestData")
  public shouldExtractExtension(
    sourcePath: string,
    [, expectedExtension]: [string, string]
  ) {
    expect(extractFileExtension(sourcePath)).toBe(expectedExtension);
  }
}
