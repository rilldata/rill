import { spawn, execSync } from "child_process";
import chokidar from "chokidar";
// spin up server

function spinUpServer() {
    console.log('spinning up server');
    // first transpile!
    execSync("npx tsc");
    const c = spawn("node", ["tsc-tmp/server/websocket.js"]);

    c.stdout.on('data', data => {
        console.log(`stdout: ${data}`);
      });

      c.stderr.on('data', data => {
        console.error(`stderr: ${data}`);
      });
    
    // listen for file changes
    c.on('error', (error) => {
        console.error(`error: ${error.message}`);
      });
      
      c.on('close', (code, other) => {
        console.log(`child process exited with code ${code}`, other);
      });

    return c;
}

let child = spinUpServer();

const watcher = chokidar.watch('./src/server/*', {
    ignored: /(^|[\/\\])\../, // ignore dotfiles]
    persistent: true
  });
watcher
    .on('change', (file) => {
        console.log('change', file);
        child.kill();
        child = spinUpServer();
    })
