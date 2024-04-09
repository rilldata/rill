import { describe, expect, it } from "vitest";
import { transformFileList } from "./transform-file-list";

describe("transformFileList", () => {
  const testCases = [
    {
      description: "transforms a flat list of files",
      fileList: ["file1.yaml", "file2.py", "file3.md"],
      expectedStructure: {
        name: "",
        path: "",
        directories: [],
        files: ["file1.yaml", "file2.py", "file3.md"],
      },
    },
    {
      description: "transforms a nested list of files into directories",
      fileList: ["dir1/fileA.sql", "dir2/dir3/fileAB.sql"],
      expectedStructure: {
        name: "",
        path: "",
        directories: [
          {
            name: "dir1",
            path: "dir1",
            directories: [],
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
