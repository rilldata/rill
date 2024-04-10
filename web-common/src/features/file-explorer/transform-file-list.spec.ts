import { V1DirEntry } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { transformFileList } from "./transform-file-list";

describe("transformFileList", () => {
  const testCases = [
    {
      description: "transforms a flat list of files",
      fileList: [
        { path: "file1.yaml", isDir: false },
        { path: "file2.py", isDir: false },
        { path: "file3.md", isDir: false },
      ] as V1DirEntry[],
      expectedStructure: {
        name: "",
        path: "",
        directories: [],
        files: ["file1.yaml", "file2.py", "file3.md"],
      },
    },
    {
      description: "transforms a nested list of files into directories",
      fileList: [
        { path: "dir1", isDir: true },
        { path: "dir1/fileA.sql", isDir: false },
        { path: "dir1/dir4", isDir: true },
        { path: "dir2/dir3", isDir: true },
        { path: "dir2/dir3/fileAB.sql", isDir: false },
      ] as V1DirEntry[],
      expectedStructure: {
        name: "",
        path: "",
        directories: [
          {
            name: "dir1",
            path: "dir1",
            directories: [
              {
                name: "dir4",
                path: "dir1/dir4",
                directories: [],
                files: [],
              },
            ],
            files: ["fileA.sql"],
          },
          {
            name: "dir2",
            path: "dir2",
            directories: [
              {
                name: "dir3",
                path: "dir2/dir3",
                directories: [],
                files: ["fileAB.sql"],
              },
            ],
            files: [],
          },
        ],
        files: [],
      },
    },
    // Additional test cases...
  ];

  testCases.forEach(({ description, fileList, expectedStructure }) => {
    it(description, () => {
      const result = transformFileList(fileList);
      expect(result).toEqual(expectedStructure);
    });
  });
});
