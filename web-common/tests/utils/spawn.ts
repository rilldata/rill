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
    const processForCommand = spawn(command, args, {
      stdio: ["inherit", "pipe", "inherit"],
      cwd: options.cwd,
      env: {
        ...process.env,
        ...(options.additionalEnv ?? {}),
      },
    });

    const timeout = setTimeout(() => {
      processForCommand.kill();
      reject(new Error(`Timeout waiting for regex match: ${pattern}`));
    }, timeoutMs);

    processForCommand.stdout.on("data", (data: Buffer | string) => {
      const output = data.toString();
      console.log(output);
      const match = output.match(pattern);
      if (match) {
        clearTimeout(timeout);
        resolve({ process: processForCommand, match });
      }
    });

    processForCommand.on("error", (err) => {
      clearTimeout(timeout);
      reject(err);
    });

    processForCommand.on("exit", (code) => {
      clearTimeout(timeout);
      reject(
        new Error(
          `Process "${command} ${args.join(" ")}" exited with code ${code} before finding match`,
        ),
      );
    });
  });
}
