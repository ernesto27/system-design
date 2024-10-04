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



