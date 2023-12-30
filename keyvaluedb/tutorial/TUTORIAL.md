

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

The struct has two methods, Set and Get, Set will add a new key value pair to the map and Get will return the value of a key if exists, 
if key not exists return an error.


add this on main.go

```bash

func main() {
    e := NewEngine()
    e.Set("foo", "bar")
    value, err := e.Get("foo")
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(value)

}

```

Run the project using 

go run main.go

this should return the value "bar"

### Add tests

Althoug we can use main.go for test the code,  is a good idea to create a test in order to check the code in a more professional way, is true that in this moment seems like a overkilll,  but because we will add more features to the code is a good idea to have automated tests at the beggining.

Create a new file engine_test.go and add the following code

```bash
package main

import "testing"

func Test_SetGetKeyValue(t *testing.T) {
	e := NewEngine()
	e.Set("foo", "bar")
	value, err := e.Get("foo")
	if err != nil {
		t.Error(err)
	}
	if value != "bar" {
		t.Error("value should be bar")
	}

	_, err = e.Get("notfound")
	if err == nil {
		t.Error("should return error")
	}

}

```

This code create a new test function called Test_SetGetKeyValue, this function create a new instance of the Engine struct, set a new key value pair and get the value of the key, if the value is not the expected return an error, also check that if we try to get a key that not exists return an error.


### Persist data on disk
At the moment we only store key, value on memory, that works fine, but the problem is that if we restart the server we lost all the data, in order to persist the data we will save the key, value pair on file separate by a space and distict items by a new line, for example

data.txt

```
foo bar
bar foo
```

#### Set data on file 

Update the Set method to save the data on file, the idea is to append the data at end of the file,  using something called append only file, this is a common pattern used in databases like redis, etcd, etc.
https://en.wikipedia.org/wiki/Append-only


```bash

type Engine struct {
	data map[string]string
	file *os.File
	mu   sync.Mutex
}

var keyValueSeparator = " "

func NewEngine() (*Engine, error) {
	file, err := os.OpenFile("data.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file data:", err)
		return nil, err
	}

	return &Engine{
		data: make(map[string]string),
		file: file,
		mu:   sync.Mutex{},
	}, nil
}

func (e *Engine) Set(key, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	_, err := e.file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		return err
	}

	_, err = e.file.WriteString(key + keyValueSeparator + value + "\n")
	if err != nil {
		fmt.Println("Error appending text:", err)
		return err
	}

	e.data[key] = value
	return nil
}
```

In this code we update the Engine structs with two new properties.
file: this property is a pointer to a os.File, this is used to read and write data from a file.
mu: this is for prevent concurrency problems when we write data to the file, 

In the  NewEngine function we open the file in read and write mode, and configure to append data to the file when writing. 
if the file not exists create a new one, we also initialize the mutex property.

In the Set function we use Lock in order to prevent problems when we write data to the file on this critical section of code, we use defer to unlock the mutex when the function finish.

After we use the Seek function to move the cursor to the end of the file, this is because we want to append data to the file.

Finally we use the WriteString function to write the key value pair to the file, we also add a new line at the end of the string, this is because we want to separate the key value pair by a new line.















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