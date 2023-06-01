import { transformFileList } from "./transform-file-list";

const testCases = [
  {
    fileList: ["file1.yaml", "file2.py", "file3.md"],
    expectedStructure: {
      name: "",
      expanded: true,
      directories: [],
      files: ["file1.yaml", "file2.py", "file3.md"],
    },
  },
  {
    fileList: ["dir1/fileA.sql", "dir2/dir3/fileAB.sql"],
    expectedStructure: {
      name: "",
      expanded: true,
      directories: [
        {
          name: "dir1",
          expanded: false,
          directories: [],
          files: ["fileA.sql"],
        },
        {
          name: "dir2",
          expanded: false,
          directories: [
            {
              name: "dir3",
              expanded: false,
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
    fileList: ["dir1/fileX.js", "dir1/dir2/fileY.py", "dir3/fileZ.md"],
    expectedStructure: {
      name: "",
      expanded: true,
      directories: [
        {
          name: "dir1",
          expanded: false,
          directories: [
            {
              name: "dir2",
              expanded: false,
              directories: [],
              files: ["fileY.py"],
            },
          ],
          files: ["fileX.js"],
        },
        {
          name: "dir3",
          expanded: false,
          directories: [],
          files: ["fileZ.md"],
        },
      ],
      files: [],
    },
  },
];

// Test the transformFileList function
for (const testCase of testCases) {
  const { fileList, expectedStructure } = testCase;
  const directoryStructure = transformFileList(fileList);
  console.log("File List:", fileList);
  console.log("Expected Directory Structure:", expectedStructure);
  console.log("Actual Directory Structure:", directoryStructure);
  console.log("--------------------------------------------");
}
