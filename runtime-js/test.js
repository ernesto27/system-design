// console.log("Start");

// readFile("main.go", (err, data) => {
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

const options = {
    encoding: 'utf8',
    mode: 0o777,
    flag: 'w'
};

writeFile('fromgo.txt', 'Hello, World!', options, (err) => {
    if (err) {
      console.error('Error writing to file:', err);
    } else {
      console.log('File written successfully');
    }
  });

console.log("End");