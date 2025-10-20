const bodyParser = require('body-parser');

// JSON Parser - parses application/json content
const jsonParser = bodyParser.json();

// URL-encoded Form Parser - parses application/x-www-form-urlencoded content
const urlencodedParser = bodyParser.urlencoded({ extended: true });

// Text Parser - parses text/plain content
const textParser = bodyParser.text();

// Raw Parser - parses binary data
const rawParser = bodyParser.raw();

// Example usage:
// app.use(jsonParser);        // for JSON requests
// app.use(urlencodedParser);  // for form submissions
// app.use(textParser);        // for plain text
// app.use(rawParser);         // for binary data
