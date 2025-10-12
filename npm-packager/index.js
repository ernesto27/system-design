const express = require('express');
const moment = require('moment');

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.get('/', (req, res) => {
  const now = moment();
  const threeHoursAgo = moment().subtract(3, 'hours');

  res.json({
    message: 'Hello World!',
    formattedTime: now.format('YYYY-MM-DD HH:mm:ss'),
    relativeTimeExample: threeHoursAgo.fromNow(),
  });
});

app.get('/health', (req, res) => {
  res.json({ status: 'OK', timestamp: new Date().toISOString() });
});

app.listen(PORT, () => {
  console.log(`Server is running on port ${PORT}`);
});