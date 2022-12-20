import { glob } from "glob";

export function getParquetFiles(sourcePath: string): Promise<Array<string>> {
  return new Promise((resolve, reject) => {
    glob(
      `${process.cwd()}/${sourcePath}/**/*.parquet`,
      {
        ignore: [
          "./node_modules/",
          "./.svelte-kit/",
          "./build/",
          "./src/",
          "./tsc-tmp",
        ],
      },
      (err, output) => {
        if (err) reject(err);
        else resolve(output);
      }
    );
  });
}
