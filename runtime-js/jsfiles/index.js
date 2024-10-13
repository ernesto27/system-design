// const fs = require('fs');
// const path = require('path');

// // Asynchronous version
// function createDirectoryAsync(dirPath) {
//   fs.mkdir(path.join(__dirname, dirPath), { recursive: true }, (err) => {
//     if (err) {
//       console.error('Error creating directory asynchronously:', err);
//     } else {
//       console.log(`Directory '${dirPath}' created asynchronously`);
//     }
//   });
// }

// createDirectoryAsync('test');


// import { createServer } from 'node:http';
// const server = createServer((req, res) => {
//   res.writeHead(200, { 'Content-Type': 'text/plain' });
//   res.end('Hello World!\n');
// });
// // starts a simple http server locally on port 3000
// server.listen(3000, '127.0.0.1', () => {
//   console.log('Listening on 127.0.0.1:3000');
// });

// const os = require('os');

// console.log('Platform:', os.platform());
// console.log('Arch:', os.arch());



// console.log({ "name": "ernesto", "age": 20 }, {"uuid": 322});

// console.log(process.env.USER);
// console.log(process.env.UAAA);

// console.log(__dirname);
// console.log(__filename);


// const myModule = require('./module');
// console.log(myModule.myFunction());


// const http = require('http');
// const server = http.createServer((req, res) => {
//   //res.writeHead(500, { 'Content-Type': 'text/plain' });
//   // res.end('Hello World!\n')
//   // res.json({
//   //   "name": "ernesto",
//   //   "age": 20
//   // })
//   res.writeHead(200, { 'Content-Type': 'application/json' });
//   res.end(JSON.stringify(req));
// });
// // starts a simple http server locally on port 3000
// server.listen(3000, '127.0.0.1', () => {
//   console.log('Listening on 127.0.0.1:3000');
// });


// let count = 0;
// let intervalId = setInterval(() => {
//     count++;
//     console.log(count);
//     if (count >= 3) {
//         clearInterval(intervalId);
//     }
// }, 100);




fetch('https://dummyjson.com/test')
    // .then(response => response.json())
    .then(data => console.log(data))
    .catch(error => console.error('Fetch error:', error));
