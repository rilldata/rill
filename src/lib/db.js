import sqlite3 from 'better-sqlite3';
// import express from 'express';
// import cors from 'cors';

// const app = express();
// app.use(express.json());
// app.use(cors());
export default sqlite3('../../microfiche.db');
