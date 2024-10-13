// console.log("Start");

// readFile("main.go", (err, data) => {
//     data.method;
//     if (err) {
//         console.log("Error reading file:", err);
//     } else {
//         console.log("File content length:", data.length);;
//     }
//     console.log("Callback executed");

// });

// setTimeout(() => {
//     console.log("SET TIMEOUT CALLBACK");
// }, 2000);

// setInterval(() => {
//     console.log("SET INTERVAL CALLBACK");
// }, 1000);

// // setTimeout(() => {
// //     console.log("SET TIMEOUT CALLBACK");
// // }, 1000);

// mkdir("test", (err, data) => {
//     console.log(data);
//     if (err) {
//         console.log("Error creating directory:", err);
//     } else {
//         console.log("Directory created:", data);
//     }
// });

// const options = {
//     encoding: 'utf8',
//     mode: 0o777,
//     flag: 'w'
// };

// writeFile('fromgo.txt', 'Hello, World!', options, (err) => {
//     if (err) {
//       console.error('Error writing to file:', err);
//     } else {
//       console.log('File written successfully');
//     }
//   });

// console.log("End");


// console.log(platform());
// console.log(arch());


// let count = 0;
// let intervalId = setInterval(() => {
//     count++;
//     console.log(count);
//     if (count >= 3) {
//         console.log("clearing interval 1");
//         clearInterval(intervalId);
//     }
// }, 1000);

// let count2 = 0;
// let intervalId2 = setInterval(() => {
//     count2++;
//     console.log(count2);
//     if (count2 >= 10) {
//         console.log("clearing interval 2");
//         clearInterval(intervalId2);
//     }
// }, 2000);


// let count3   = 0;
// let intervalId3 = setInterval(() => {
//     count3++;
//     console.log(count3);
//     if (count3 >= 3) {
//         console.log("clearing interval 3");
//         clearInterval(intervalId3);
//     }
// }, 500);


// let count4 = 0;
// let intervalId4 = setInterval(() => {
//     count4++;
//     console.log(count4);
// }, 1000);

// console.log({ "name": "ernesto", "age": 20 }, {"uuid": 322});
// console.log(process.env.USER);
// console.log(process.env.GOPATH);

// console.log(__dirname);
// console.log(__filename);


// const myModule = require("./module.js");
// const myModulea = require("./module.js");
// const myModuleaa = require("./module.js");
// console.log(myModule.myFunction());
// console.log(myModule.myFunction2());

// const myModule2 = require("./module2.js");
// console.log(myModule2.moduleTwoFunction());


// const server = createServer((req, res) => {
//     // res.writeHead(200, { 'Content-Type': 'application/json' });
//     // // res.end('This is a test\n');
//     // res.json({
//     //     "name": "ernesto",
//     //     "age": 20
        
//     // });
//     res.end(req.method)
// });

// server.listen(6000, '127.0.0.1', () => {
//     console.log('Listening on 127.0.0.1:3000');
// });

// throw new Error("test dfdfd");



fetch('https://dummyjson.com/test')
    .then(response => response.json())
    // .then(data => console.log(data))
    .catch(error => console.error('Fetch error:', error));