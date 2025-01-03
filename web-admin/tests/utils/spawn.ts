import { ChildProcess, spawn } from "child_process";

export type SpawnAndMatchResult = {
  process: ChildProcess;
  match: RegExpMatchArray;
};

export async function spawnAndMatch(
  command: string,
  args: string[],
  pattern: RegExp,
  options: {
    timeoutMs?: number;
  } = {},
): Promise<SpawnAndMatchResult> {
  const { timeoutMs = 30000 } = options;

  return new Promise((resolve, reject) => {
    const process = spawn(command, args, {
      stdio: ["inherit", "pipe", "inherit"],
    });

    const timeout = setTimeout(() => {
      process.kill();
      reject(new Error(`Timeout waiting for regex match: ${pattern}`));
    }, timeoutMs);

    process.stdout.on("data", (data: Buffer | string) => {
      const output = data.toString();
      const match = output.match(pattern);
      if (match) {
        clearTimeout(timeout);
        resolve({ process, match });
      }
    });

    process.on("error", (err) => {
      clearTimeout(timeout);
      reject(err);
    });

    process.on("exit", (code) => {
      clearTimeout(timeout);
      reject(
        new Error(`Process exited with code ${code} before finding match`),
      );
    });
  });
}
