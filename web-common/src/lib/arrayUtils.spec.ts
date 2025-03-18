import {
  arrayUnorderedEquals,
  createBatches,
} from "@rilldata/web-common/lib/arrayUtils";
import { describe, expect, it } from "vitest";

describe("createBatches", () => {
  const Input = ["a", "b", "c", "d", "e", "f", "g", "h"];
  const TestCases = [
    {
      batchSize: 3,
      expected: [
        ["a", "b", "c"],
        ["d", "e", "f"],
        ["g", "h"],
      ],
    },
    {
      batchSize: 4,
      expected: [
        ["a", "b", "c", "d"],
        ["e", "f", "g", "h"],
      ],
    },
    {
      batchSize: 10,
      expected: [["a", "b", "c", "d", "e", "f", "g", "h"]],
    },
  ];

  for (const { batchSize, expected } of TestCases) {
    it(`input=${Input.join(",")} batchSize=${batchSize}`, () => {
      expect(createBatches(Input, batchSize)).toEqual(expected);
    });
  }
});

describe("arrayUnorderedEquals", () => {
  const TestCases = [
    {
      src: [1, 2, 3],
      tar: [3, 2, 1],
      equals: true,
    },
    {
      src: [1, 2, 3],
      tar: [1, 2],
      equals: false,
    },
    {
      src: [1, 2, 3],
      tar: [1, 2, 1],
      equals: false,
    },
  ];
  for (const { src, tar, equals } of TestCases) {
    it(`arrayUnorderedEquals(${src.join(",")}, ${tar.join(",")})`, () => {
      expect(arrayUnorderedEquals(src, tar)).toEqual(equals);
    });
  }
});
