import { spawn } from "child_process";

export default async () => {
    const buildArgs = ["setup-e2e-rill.sh"];
    if (process.env.SKIP_UI_BUILD === "true") buildArgs.push("-s");

    const childProcess = spawn("bash", buildArgs, {
        stdio: "inherit",
        shell: true,
        cwd: "./.jest"
    });

    return new Promise((resolve, reject) => {
        childProcess.on("close", resolve);
        childProcess.on("error", reject);
    })
}
