import { Server } from "socket.io";
import { 
    createStore, 	
    initializeFromSavedState,
	saveToLocalFile,
	resettable,
    addProduce,
    connectStateToSocket,
    addActions } from "../lib/create-store.js";

import * as api from "./duckdb.js";

import { createServerActions, emptyQuery, initialState as createInitialState } from "./actions.js"

let socket;

const io = new Server({ cors: {
    origin: "http://localhost:3000",
    methods: ["GET", "POST"]
  } });

const initialState = createInitialState();

const serverActions = createServerActions(api, ({ message, type }) => socket.emit("notification", { message, type }));

const store = createStore(
    initialState,
    //initializeFromSavedState('saved-state')(initialState),
    addProduce(),
    addActions(serverActions),
    resettable(initialState),
    connectStateToSocket(), // emit to socket any state changes via store.subscribe
    saveToLocalFile('saved-state') // let's save to our local file
)

console.log('initialized the store.');

io.on("connection", thisSocket => {
    console.log('connected', thisSocket.id);
    socket = thisSocket;
    socket.emit("app-state", store.get());
    //console.log(Object.keys(store))
    store.scanRootForSources();
    //store.produce(serverActions(undefined, undefined).scanRootForSources());
    //store.produce(store.scanForRootSources());
    store.connectStateToSocket(socket);
  });



io.listen(3001);
