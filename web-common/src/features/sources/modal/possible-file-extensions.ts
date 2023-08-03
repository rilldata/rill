export const PossibleFileExtensions = [
  ".parquet",
  ".csv",
  ".tsv",
  ".txt",
  ".json",
  ".ndjson",
];
export const PossibleZipExtensions = [".gz"];
export const AllFileExtensions = [
  ...PossibleFileExtensions,
  ...PossibleFileExtensions.map((extension) =>
    PossibleZipExtensions.map(
      (zippedExtension) => `${extension}${zippedExtension}`
    )
  ).flat(),
];

export function fileHasValidExtension(fileName: string) {
  // Since this check happens pretty rarely and this list is small a brute force check is sufficient
  return AllFileExtensions.some((extension) => fileName.endsWith(extension));
}
