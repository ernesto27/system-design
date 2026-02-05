const axios = require('axios');

async function fetchData() {
  try {
    console.log('Fetching data from JSONPlaceholder API...\n');

    // GET request example
    const response = await axios.get('https://jsonplaceholder.typicode.com/posts/1');

    console.log('Response Status:', response.status);
    console.log('Response Data:');
    console.log(JSON.stringify(response.data, null, 2));

    console.log('\n---\n');

    // POST request example
    const postData = {
      title: 'Test Post',
      body: 'This is a test post using axios',
      userId: 1
    };

    console.log('Creating a new post...\n');
    const postResponse = await axios.post('https://jsonplaceholder.typicode.com/posts', postData);

    console.log('POST Response Status:', postResponse.status);
    console.log('Created Post:');
    console.log(JSON.stringify(postResponse.data, null, 2));

  } catch (error) {
    console.error('Error occurred:');
    if (error.response) {
      console.error('Status:', error.response.status);
      console.error('Data:', error.response.data);
    } else if (error.request) {
      console.error('No response received:', error.request);
    } else {
      console.error('Error:', error.message);
    }
  }
}

fetchData();
