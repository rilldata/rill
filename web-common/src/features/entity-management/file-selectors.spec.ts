import { useFileNamesInDirectorySelector } from "@rilldata/web-common/features/entity-management/file-selectors";
import { describe, expect, it } from "vitest";

describe("useFileNamesInDirectorySelector", () => {
  it("returns an empty array if no files are present", () => {
    const data = { files: undefined };
    const directoryPath = "/";
    const result = useFileNamesInDirectorySelector(data, directoryPath);
    expect(result).toEqual([]);
  });

  it("filters out files not in the immediate directory", () => {
    const data = {
      files: [
        { path: "/a/file1.txt", isDir: false },
        { path: "/a/b/file2.txt", isDir: false },
        { path: "/b/file3.txt", isDir: false },
      ],
    };
    const directoryPath = "/a";
    const result = useFileNamesInDirectorySelector(data, directoryPath);
    expect(result).toEqual(["file1.txt"]);
  });

  it("filters out directories and includes only files in the immediate directory", () => {
    const data = {
      files: [
        { path: "/a/file1.txt", isDir: false },
        { path: "/a/b", isDir: true },
      ],
    };
    const directoryPath = "/a";
    const result = useFileNamesInDirectorySelector(data, directoryPath);
    expect(result).toEqual(["file1.txt"]);
  });

  it("ensures files from subdirectories are not included", () => {
    const data = {
      files: [
        { path: "/a/file1.txt", isDir: false },
        { path: "/a/b/file2.txt", isDir: false },
      ],
    };
    const directoryPath = "/a";
    const result = useFileNamesInDirectorySelector(data, directoryPath);
    expect(result).toEqual(["file1.txt"]);
  });

  it("sorts filenames alphabetically and case-insensitively", () => {
    const data = {
      files: [
        { path: "/a/zeta.txt", isDir: false },
        { path: "/a/alpha.txt", isDir: false },
        { path: "/a/beta.txt", isDir: false },
      ],
    };
    const directoryPath = "/a";
    const result = useFileNamesInDirectorySelector(data, directoryPath);
    expect(result).toEqual(["alpha.txt", "beta.txt", "zeta.txt"]);
  });
});
