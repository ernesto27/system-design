const express = require('express');
const app = express();
const port = process.env.PORT || 3000;

app.use(express.json());

app.get('/version', (req, res) => {
  res.json({
    version: '1.0.0',
    name: 'CPU Pressure API',
    timestamp: new Date().toISOString()
  });
});

app.get('/cpu-pressure', (req, res) => {
  const duration = 50000; // default 50 seconds
  const intensity = 10;
  
  const startTime = Date.now();
  let result = 0;
  
  const workers = [];
  for (let i = 0; i < intensity; i++) {
    workers.push(new Promise((resolve) => {
      const workerStart = Date.now();
      while (Date.now() - workerStart < duration) {
        result += Math.random() * Math.PI * Math.sqrt(Math.random() * 1000000);
        for (let j = 0; j < 1000; j++) {
          Math.sin(Math.random()) * Math.cos(Math.random());
        }
      }
      resolve(result);
    }));
  }
  
  Promise.all(workers).then(() => {
    const endTime = Date.now();
    res.json({
      message: 'CPU pressure test completed',
      duration: `${endTime - startTime}ms`,
      requestedDuration: `${duration}ms`,
      intensity: intensity,
      result: result.toString().substring(0, 10)
    });
  });
});

app.get('/health', (req, res) => {
  res.json({ status: 'healthy', uptime: process.uptime() });
});

app.listen(port, () => {
  console.log(`CPU Pressure API running on port ${port}`);
  console.log(`Endpoints:`);
  console.log(`  GET /version - API version info`);
  console.log(`  GET /cpu-pressure?duration=5000&intensity=1 - CPU intensive task`);
  console.log(`  GET /health - Health check`);
});
