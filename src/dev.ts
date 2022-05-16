import { spawn } from "node:child_process";

function spawnNpmProcess(scriptName) {
  console.log(`Starting "npm run ${scriptName}"`);
  const childProcess = spawn("npm", ["run", scriptName], {
    stdio: "inherit",
  });
  childProcess.on("error", (err) => {
    console.log(`Script "npm run ${scriptName}" failed.`, err);
  });
  childProcess.on("exit", (code, other) => {
    console.log(
      `Script "npm run ${scriptName}" exited with code ${code}`,
      other
    );
    setImmediate(() => {
      spawnNpmProcess(scriptName);
    });
  });
}

spawnNpmProcess("dev:ui");
spawnNpmProcess("dev:backend");
