import { execSync, spawn } from "node:child_process";

export default function () {
  execSync("export RILL_EXTERNAL_RUNTIME=true");
  global.runtimeProcess = spawn("./dist/runtime/runtime", [], {
    env: {
      ...process.env,
      RILL_RUNTIME_ENV: "production",
      RILL_RUNTIME_LOG_LEVEL: "warn",
      RILL_RUNTIME_HTTP_PORT: "8081",
      RILL_RUNTIME_GRPC_PORT: "9081",
    },
    stdio: "inherit",
    shell: true,
  });
}
