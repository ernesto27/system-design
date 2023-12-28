
TUTORIAL

1

Explain project,  show examples

Show how works

Show basic diagram or explanation of api

Version using only memory hash without persistence 



In this tutorial we will create a simple key value database using go.
We can think in somethink like redis or etcd but with a very limited set of features.

Why
Because i think make something from scratch is a great way to learn and understand how things works under the hood, 
and also is a great way to learn a new language if you are starting with or have little experience with Go creating 
a real like world project

How it works
The database will be a simple key value store, we will use a hash map to store the data in memory and we will use a file to persist the data on disk.


In order to mantain the project simple we will have a simple http api to interact with it, we will use the standard Go library to create the http server, altouhg use HTTP is an overhead, we will use it because is simple and easy to use instead of create a custom protocol like redis, mysql, etcd, etc.

Example using curl

curl http://localhost:8080/set?key=foo&value=bar

Example get value by key
curl http://localhost:8000/get?key=foo

We  will also create a library to interact with the database from go code

e := db.Connect("localhost", 8080)
e := db.Set("foo", "bar") // nil
e := db.Get("foo") // bar

In this first part of the tutorial we will do this 

- Create go project 
- Create engine package
- Set and Get values from memory
- Add tests














2

Add file  append to persists  data 

use iteration to search values

use hash - byte offset to get value from file on O(1)

tests

3 

restart , restore items

Delete item

Limit files,  use various files for data

Service HTTP db