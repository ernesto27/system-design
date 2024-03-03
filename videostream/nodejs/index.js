import { S3Client, GetObjectCommand, PutObjectCommand } from '@aws-sdk/client-s3';

import { Readable } from 'stream';
import { spawn, execSync} from 'node:child_process';

const s3 = new S3Client({region: 'us-west-2'});
import * as fs from 'fs';

export const handler = async (event) => {
  const srcBucket = event.Records[0].s3.bucket.name;
    
  // Object key may have spaces or unicode non-ASCII characters
  const srcKey    = decodeURIComponent(event.Records[0].s3.object.key.replace(/\+/g, " "));
  const dstBucket = srcBucket + "-resized";
  const dstKey    = "resized-" + srcKey;
  
  try {
    const params = {
      Bucket: srcBucket,
      Key: srcKey
    };
    var response = await s3.send(new GetObjectCommand(params));
    var stream = response.Body;
  
  // Convert stream to buffer to pass to sharp resize function.
    if (stream instanceof Readable) {
      var content_buffer = Buffer.concat(await stream.toArray());
 
      console.log("save file");
      
      const filePath = "/tmp/input.mp4";
      
      try {
        // Write the Buffer content to a file synchronously
        fs.writeFileSync(filePath, content_buffer);
        console.log('File saved successfully.');
        
        const resolution = '640x480';
        const preset = 'slow';
        const crf = '18';
        const outputFile = "/tmp/output.mp4";
                
        try {
          const output = execSync(`ffmpeg -y -i ${filePath} -vf scale=${resolution} -preset ${preset} -crf ${crf} ${outputFile}`, { encoding: 'utf-8' });
      
          console.log(output); // Print ffmpeg output for demonstration purposes
      
            console.log('ffmpeg process completed successfully.');
        } catch (error) {
            console.error(`ffmpeg process exited with error: ${error.stderr}`);
        }
      
              
        
      } catch (err) {
        console.error('Error writing file:', err);
      }
      
      
    } else {
      throw new Error('Unknown object stream type');
    }
  } catch (error) {
    console.log(error);
    return;
  }
};



///////////////////////////////////////////////////////////////////////

// const { spawn } = require('child_process');
// const  os   = require("os");
// const path = require("path")

// const AWS = require('aws-sdk');
// const fs = require('fs');


// const workdir = os.tmpdir();
// const filename = `example-${Date.now().toString()}.mp4`;
// const outputFile = path.join(workdir, filename);
// console.log("🚀 @debug:output", outputFile);

// const inputVideo = 'video.mp4';
// const resolution = '640x480';
// const preset = 'slow';
// const crf = '18';


// // Execute the ffmpeg command
// const ffmpegProcess = spawn('ffmpeg', ['-i', inputVideo, '-vf', `scale=${resolution}`, '-preset', preset, '-crf', crf, outputFile]);

// // Listen for data events on stderr to extract progress information
// ffmpegProcess.stderr.on('data', (data) => {
//     const output = data.toString();
//     console.log(output); // Print stderr output for demonstration purposes
// });

// // Handle process exit
// ffmpegProcess.on('close', (code) => {
//     if (code === 0) {
//         console.log('ffmpeg process completed successfully.');


//         // Set the region and access keys
//         AWS.config.update({
//             region: '',
//             accessKeyId: "",
//             secretAccessKey: ""
//         });

//         // Create a new instance of the S3 class
//         const s3 = new AWS.S3();

//         // Set the parameters for the file you want to upload
//         const params = {
//             Bucket: "videostream-666",
//             Key: 'video.mp4',
//             Body: fs.createReadStream(outputFile)
//         };

//         // Upload the file to S3
//         s3.upload(params, (err, data) => {
//             if (err) {
//                 console.log('Error uploading file:', err);
//             } else {
//                 console.log('File uploaded successfully. File location:', data.Location);
//             }
//         });


//     } else {
//         console.error(`ffmpeg process exited with code ${code}.`);
//     }
// });


