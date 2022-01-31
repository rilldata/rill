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

import { createDataModelerActions, initialState as createInitialState } from "./actions.js"

let socket;

const io = new Server({ cors: {
    origin: "http://localhost:3000",
    methods: ["GET", "POST"]
  } });

const initialState = createInitialState();

const serverActions = createDataModelerActions(
  // plug in the duckdb API
  api, 
  // hook for user function
  ({ message, type }) => socket.emit("notification", { message, type })
);

const store = createStore(
    initializeFromSavedState('saved-state')(initialState),
    // add our main production method. This gives us the store.produce function.
    addProduce(), 
    // after we've added our produce and thunks, let's add a bunch of thunks
    // to our store.
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
//store.scanRootForSources();

io.on("connection", thisSocket => {
    console.log('connected', thisSocket.id);
    socket = thisSocket;

    socket.emit("app-state", store.get());
    store.connectStateToSocket(socket);
  });

  const watcher = chokidar.watch('./**/*.parquet', {
    ignored: /(^|[\/\\])\../, // ignore dotfiles]
    persistent: true
  });

let timeoutId:any;
function watch() {
  if (timeoutId) clearTimeout(timeoutId);
  timeoutId = setTimeout(() => store.scanRootForSources(), 1000);
}
watcher
    .on('change', watch)
    .on('add', watch)
    .on('unlink', watch)



io.listen(3001);
