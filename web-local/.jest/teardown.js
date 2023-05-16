import { spawn } from "child_process";

export default async () => {
    const childProcess = spawn("rm", ["-rf", "rill-e2e-test"], {
        stdio: "inherit",
        shell: true,
        cwd: "./.jest"
    });

    return new Promise((resolve, reject) => {
        childProcess.on("close", resolve);
        childProcess.on("error", reject);
    })
}
