
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

We send a json payload with a key, value dataW

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

I assume that you have go installed and you have basic knowledge of the language  and at least create some basic project , if not please check the official documentation https://golang.org/doc/

### Create go project

Create a new folder, go to the folder and create a new go module with the name keyvaluedb ( you can change this name for the name you want, is not important at this moment)

mkdir keyvaluedb && cd keyvaluedb && go mod init keyvaluedb

Create a new file main.go and add the following code to check that everything is working

main.go
```bash
package main


import "fmt"

func main() {
    fmt.Println("Hello world")
}
```

Run the project

go run main.go

You should see the message Hello world print in the console.

### Create engine package

We will create a new package called engine, this package will contain all the logic of the database,  create a new file engine.go and add the following code.
This file is part of the main package.

engine.go
```bash

package main

import (
    "errors"
    "sync"
)

type Engine struct {
    data map[string]string
}

func NewEngine() *Engine {
    return &Engine{
        data: make(map[string]string),
    }
}

func (e *Engine) Set(key, value string) error {
    e.data[key] = value
    return nil
}

func (e *Engine) Get(key string) (string, error) {
    value, ok := e.data[key]
    if !ok {
        return "", errors.New("key not found")
    }
    return value, nil
}

```

This code create a new struct called Engine, this struct will contain the data of the database, in this case we will use a simple map to store the data in memory, we will add persistence on a file later.

The NewEngine function will create a new instance of the Engine struct, the important thing here is notice that we initialize the data map with make, this is important because if we don't do this the map will be nil and we will get a panic when we try to add a new key value pair.

The struct has two methods, Set and Get, Set will add a new key value pair to the map and Get will return the value of a key if exists.





















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