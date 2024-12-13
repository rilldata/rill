import { type Page } from "@playwright/test";
import { exec, spawn } from "child_process";
import { promisify } from "util";

const execAsync = promisify(exec);

export async function cliLogin(page: Page) {
  // Run the login command
  const loginProcess = spawn("rill", ["login"], {
    stdio: ["inherit", "pipe", "inherit"],
  });

  // Capture the verification URL from the CLI output
  let verificationUrl = "";
  loginProcess.stdout.on("data", (data) => {
    const output = data.toString();
    const match = output.match(
      /Open this URL in your browser to confirm the login: (.*)\n/,
    );
    if (match) {
      verificationUrl = match[1];
    }
  });
  while (!verificationUrl) {
    await new Promise((resolve) => setTimeout(resolve, 100));
  }

  // Manually navigate to the verification URL
  await page.goto(verificationUrl);

  // Click the confirm button
  await page.getByRole("button", { name: /confirm/i }).click();

  // Wait for the process to complete
  await new Promise((resolve, reject) => {
    loginProcess.on("close", (code) => {
      if (code === 0) resolve(null);
      else reject(new Error(`Process exited with code ${code}`));
    });
  });
}

export async function cliLogout() {
  await execAsync("rill logout");
}
