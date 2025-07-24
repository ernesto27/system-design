# LET CODE SYSTEM

API Endpoints
GET /problems?start={start}&end={end}

GET /problems/:problem_id


schema tables

USERS 
- id (int, primary key)
- email (string, unique)


PROBLEMS
- id (int, primary key)
- title (string)
- description (text)
- difficulty (string)
- constraints (text)
- examples (text) 


