import { Server } from "socket.io";
import chokidar from "chokidar";

import { 
    createStore, 	
    initializeFromSavedState,
	saveToLocalFile,
	resettable,
    addProduce,
    connectStateToSocket,
    addActions } from "../lib/create-store.js";

import * as api from "./duckdb.js";

import { createServerActions, initialState as createInitialState } from "./actions.js"

let socket;

const io = new Server({ cors: {
    origin: "http://localhost:3000",
    methods: ["GET", "POST"]
  } });

const initialState = createInitialState();

const serverActions = createServerActions(api, 
  // hook for user function
  ({ message, type }) => socket.emit("notification", { message, type })
);

const store = createStore(
    //initialState,
    initializeFromSavedState('saved-state')(initialState),
    addProduce(),
    addActions(serverActions),
    resettable(initialState),
    connectStateToSocket(), // emit to socket any state changes via store.subscribe
    saveToLocalFile('saved-state') // let's save to our local file
)

/** Set the frontend db status state. */
api.registerDBRunCallbacks(
  () => { store.setDBStatus("running"); },
  () => { store.setDBStatus("idle"); }
)

console.log('initialized the store.');

io.on("connection", thisSocket => {
    console.log('connected', thisSocket.id);
    socket = thisSocket;
    socket.emit("app-state", store.get());
    store.scanRootForSources();
    store.connectStateToSocket(socket);
  });

  const watcher = chokidar.watch('./**/*.parquet', {
    ignored: /(^|[\/\\])\../, // ignore dotfiles]
    persistent: true
  });

let timeoutId:any;
function watch() {
  if (timeoutId) clearTimeout(timeoutId);
  timeoutId = setTimeout(() => store.scanRootForSources(), 100);
}
watcher
    .on('change', watch)
    .on('add', watch)
    .on('unlink', watch)



io.listen(3001);
