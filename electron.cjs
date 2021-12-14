//const { app, BrowserWindow } = import('electron');
//const path = import('path');
//import { app, BrowserWindow } from 'electron';
//import path from 'path';

const { app, BrowserWindow } = require('electron');
const path = require('path');
const exec = require('child_process');
function serve() {
	let server;
	function toExit() {
		if (server) server.kill(0);
	}

	//if (server) return;
	server = exec.spawn('node', ['./server/server.duckdb.js'], {
		stdio: ['ignore', 'inherit', 'inherit'],
		shell: true
	});

	process.on('SIGTERM', toExit);
	process.on('exit', toExit);
}

const port = process.env.PORT || 3000;
//import('./server/server.duckdb.js').catch(console.log);
serve();
app.on('ready', () => {
	const mainWindow = new BrowserWindow();
	//mainWindow.loadFile(path.join(__dirname, 'build/index.html'));
	mainWindow.loadURL(`http://localhost:${port}`);
	mainWindow.webContents.openDevTools();
	//import('./server/server.duckdb');
});
