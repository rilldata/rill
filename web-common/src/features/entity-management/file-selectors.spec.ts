import {
  useDirectoryNamesInDirectorySelector,
  useFileNamesInDirectorySelector,
} from "@rilldata/web-common/features/entity-management/file-selectors";
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

describe("useDirectoryNamesInDirectorySelector", () => {
  it("returns an empty array if no files are present", () => {
    const data = { files: undefined };
    const directoryPath = "/";
    const result = useDirectoryNamesInDirectorySelector(data, directoryPath);
    expect(result).toEqual([]);
  });

  it("filters out non-directory entries", () => {
    const data = {
      files: [
        { path: "/a/file.txt", isDir: false },
        { path: "/a/b", isDir: true },
      ],
    };
    const directoryPath = "/a";
    const result = useDirectoryNamesInDirectorySelector(data, directoryPath);
    expect(result).toEqual(["b"]);
  });

  it("excludes directories not directly under the specified path", () => {
    const data = {
      files: [
        { path: "/a/b", isDir: true },
        { path: "/a/b/c", isDir: true },
      ],
    };
    const directoryPath = "/a";
    const result = useDirectoryNamesInDirectorySelector(data, directoryPath);
    expect(result).toEqual(["b"]);
  });

  it("does not include the same directory or nested deeper levels", () => {
    const data = {
      files: [
        { path: "/a", isDir: true },
        { path: "/a/b", isDir: true },
        { path: "/a/b/c", isDir: true },
      ],
    };
    const directoryPath = "/a";
    const result = useDirectoryNamesInDirectorySelector(data, directoryPath);
    expect(result).toEqual(["b"]);
  });

  it("sorts directory names alphabetically and case-insensitively", () => {
    const data = {
      files: [
        { path: "/a/Zeta", isDir: true },
        { path: "/a/alpha", isDir: true },
        { path: "/a/Beta", isDir: true },
      ],
    };
    const directoryPath = "/a";
    const result = useDirectoryNamesInDirectorySelector(data, directoryPath);
    expect(result).toEqual(["alpha", "Beta", "Zeta"]);
  });
});
