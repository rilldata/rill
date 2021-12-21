import { Server } from "socket.io";
import { 
    createStore, 	
    initializeFromSavedState,
	saveToLocalFile,
	resettable,
    addProduce,
    connectStateToSocket,
    addActions } from "../src/lib/create-store.js";

import * as api from "./duckdb.mjs";

import { createServerActions, emptyQuery } from "./actions.js"

let socket;

const io = new Server({   cors: {
    origin: "http://localhost:3000",
    methods: ["GET", "POST"]
  } });

const initialState = {
    queries: [emptyQuery()]
}

const store = createStore(
    initializeFromSavedState('saved-state')(initialState),
    addProduce(),
    addActions(createServerActions(api)),
    resettable(initialState),
    connectStateToSocket(), // emit to socket any state changes via store.subscribe
    saveToLocalFile('saved-state') // let's save to our local file
)

const actions = createServerActions();
console.log('initialized the store.');

io.on("connection", thisSocket => {
    console.log('connected', thisSocket.id);
    socket = thisSocket;
    socket.emit("app-state", store.get());
    store.connectStateToSocket(socket);
  });

io.listen(3001);
