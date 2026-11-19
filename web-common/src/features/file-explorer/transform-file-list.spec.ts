import type { V1DirEntry } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { transformFileList } from "./transform-file-list";

describe("transformFileList", () => {
  const testCases = [
    {
      description: "transforms a flat list of files",
      fileList: [
        { path: "/file1.yaml", isDir: false },
        { path: "/file2.py", isDir: false },
        { path: "/file3.md", isDir: false },
      ] as V1DirEntry[],
      expectedStructure: {
        name: "",
        path: "/",
        directories: [],
        files: ["file1.yaml", "file2.py", "file3.md"],
      },
    },
    {
      description: "transforms a nested list of files into directories",
      fileList: [
        { path: "/dir1", isDir: true },
        { path: "/dir1/fileA.sql", isDir: false },
        { path: "/dir1/dir4", isDir: true },
        { path: "/dir2/dir3", isDir: true },
        { path: "/dir2/dir3/fileAB.sql", isDir: false },
      ] as V1DirEntry[],
      expectedStructure: {
        name: "",
        path: "/",
        directories: [
          {
            name: "dir1",
            path: "/dir1",
            directories: [
              {
                name: "dir4",
                path: "/dir1/dir4",
                directories: [],
                files: [],
              },
            ],
            files: ["fileA.sql"],
          },
          {
            name: "dir2",
            path: "/dir2",
            directories: [
              {
                name: "dir3",
                path: "/dir2/dir3",
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
    {
      description: "lists hidden directories after visible ones at every level",
      // The input mirrors the explorer's alphabetical sort, where dot-prefixed
      // entries come first.
      fileList: [
        { path: "/.claude/skills/analyst/SKILL.md", isDir: false },
        { path: "/.env", isDir: false },
        { path: "/models/.scratch/draft.sql", isDir: false },
        { path: "/models/events.sql", isDir: false },
        { path: "/models/staging", isDir: true },
        { path: "/rill.yaml", isDir: false },
      ] as V1DirEntry[],
      expectedStructure: {
        name: "",
        path: "/",
        directories: [
          {
            name: "models",
            path: "/models",
            directories: [
              {
                name: "staging",
                path: "/models/staging",
                directories: [],
                files: [],
              },
              {
                name: ".scratch",
                path: "/models/.scratch",
                directories: [],
                files: ["draft.sql"],
              },
            ],
            files: ["events.sql"],
          },
          {
            name: ".claude",
            path: "/.claude",
            directories: [
              {
                name: "skills",
                path: "/.claude/skills",
                directories: [
                  {
                    name: "analyst",
                    path: "/.claude/skills/analyst",
                    directories: [],
                    files: ["SKILL.md"],
                  },
                ],
                files: [],
              },
            ],
            files: [],
          },
        ],
        // Only directories are reordered; files keep the caller's order.
        files: [".env", "rill.yaml"],
      },
    },
  ];

  testCases.forEach(({ description, fileList, expectedStructure }) => {
    it(description, () => {
      const result = transformFileList(fileList);
      expect(result).toEqual(expectedStructure);
    });
  });
});
