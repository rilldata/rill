import { ChildProcess, exec, spawn } from "child_process";
import { promisify } from "util";

export const execAsync = promisify(exec);

export type SpawnAndMatchResult = {
  process: ChildProcess;
  match: RegExpMatchArray;
};

export async function spawnAndMatch(
  command: string,
  args: string[],
  pattern: RegExp,
  options: {
    cwd?: string;
    timeoutMs?: number;
    additionalEnv?: NodeJS.ProcessEnv;
  } = {},
): Promise<SpawnAndMatchResult> {
  const { timeoutMs = 30000 } = options;

  return new Promise((resolve, reject) => {
    const childProcess = spawn(command, args, {
      stdio: ["inherit", "pipe", "inherit"],
      cwd: options.cwd,
      env: {
        ...process.env,
        ...(options.additionalEnv ?? {}),
      },
    });

    const timeout = setTimeout(() => {
      childProcess.kill();
      reject(new Error(`Timeout waiting for regex match: ${pattern}`));
    }, timeoutMs);

    childProcess.stdout.on("data", (data: Buffer | string) => {
      const output = data.toString();
      console.log(output);
      const match = output.match(pattern);
      if (match) {
        clearTimeout(timeout);
        resolve({ process: childProcess, match });
      }
    });

    childProcess.on("error", (err) => {
      clearTimeout(timeout);
      reject(err);
    });

    childProcess.on("exit", (code) => {
      clearTimeout(timeout);
      reject(
        new Error(
          `Process "${command} ${args.join(" ")}" exited with code ${code} before finding match`,
        ),
      );
    });
  });
}
