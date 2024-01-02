# KeyValue Database tutorial

In this tutorial we will create a simple key value database using go.
We can think in something like redis or etcd but with a very limited set of features.

### But Why?
Because i think make something from scratch is a great way to learn and understand how things works under the hood, 
also if you want to learn a new language or do you have a little experience with Go creating 
a real like world project is a great way to learn the language.

### How it works
The database will be a simple key value store, we will use a hash map to store the data in memory and we will use a file to persist the data on disk.


We will use tests to check that everything works as expected after we add or modify the code.
also we will have a simple http api to interact with it, we will use the standard Go library to create the http server, although use HTTP is seems like  an overhead, we are going to use it because is simpler and easy to implement instead of create a custom protocol like redis, mysql, etcd or another database.

Example using curl

We send a json payload with a key, value dataW

```bash
curl -X POST -H "Content-Type: application/json" -d '{"key": "mykey", "value": "from curl"}' http://localhost:8000/set
```

Example get value by key
```bash
curl http://localhost:8000/get?key=foo
```

We  will also create a library to interact with the database from go code

e := db.Connect("localhost", 8080)
e := db.Set("foo", "bar") // nil
e := db.Get("foo") // bar

In this first part of the tutorial we will do this 

- Create go project
- Create engine package
- Add tests
- Persist data on disk
- Set data on file 
- Get data from key
- Compact data from file

I assume that you have go installed and you have basic knowledge of the language  and at least create some basic project , if not please check the official documentation https://golang.org/doc/

### Create go project

Create a new folder, go to the folder and create a new go module with the name keyvaluedb ( you can change this name for the name you want, is not important at this moment)

mkdir keyvaluedb && cd keyvaluedb && go mod init keyvaluedb

Create a new file main.go and add the following code to check that everything is working

main.go
```go
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
```go

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

```go

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

```go
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


```go
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



### Get data from key

Previusly we access to data from key using a map, this is because we saved the the value on memory, this works fine but because now we save the data on a file we search the data here.

We will continue uses a map to store key,  but we will change the type of the value,  instead of save the string value,  we will save the byte offset of the value on the file, in this way we can access directly to the value of the key using 0(1) time complexity,  another option could be to loop over the file and search the value, but this will be O(n) time complexity and if we a lot of entries this could be pretty slow and inefficient.

we have to make this changes
engine.go

```go
type Engine struct {
	data map[string]int64
	file *os.File
	mu   sync.Mutex
}

func NewEngine() (*Engine, error) {
	file, err := os.OpenFile("data.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file data:", err)
		return nil, err
	}

	return &Engine{
		data: make(map[string]int64),
		file: file,
		mu:   sync.Mutex{},
	}, nil
}

func (e *Engine) Set(key, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	offset, err := e.file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		return err
	}

	_, err = e.file.WriteString(key + keyValueSeparator + value + "\n")
	if err != nil {
		fmt.Println("Error appending text:", err)
		return err
	}

	e.data[key] = offset
	return nil
}

```

In this code we change the type of the data map from string to int64 in order to save the offset of key, 
also in the Set we obtaint the offset using the Seek function and save it on the map after write the data to the file,
we use the io.SeekEnd parameter to move the cursor to the end of the file, remember that we always want to append data to end of the file,
with this approach if you use set multiple times with the same key the value will be append to the file and multiple entries with the same 
key will be created, we are going to fix that problem later.


Wc must update the Get function on engine.go

```go
func (e *Engine) Get(key string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.data[key]; !ok {
		return "", fmt.Errorf("key not found")
	}

	_, err := e.file.Seek(e.data[key]+int64(len(key))+1, 0)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		return "", err
	}

	buffer := make([]byte, 1)
	var content []byte

	for {
		n, err := e.file.Read(buffer)
		if err != nil {
			fmt.Println("Error reading file:", err)
			break
		}

		if n == 0 {
			break
		}

		if buffer[0] == '\n' {
			break
		}

		content = append(content, buffer[0])
	}
	return string(content), nil
}
```

We use the Seek function to move the cursor to the offset of the key, we also add a little tricky logic that we will next,
for example if we have the key-value "foo bar",
we have to obtaint the offset for key saved on the data map , plus the len of the key bar (4) plus (1) for the space separator, after that we have the cursor at the start of the value.

```bash
_, err := e.file.Seek(e.data[key]+int64(len(key))+1, 0)
```

Next we create a buffer of 1 byte, this is because we need to read the file byte by byte, we also create a content variable to save the value of the key, we use a for loop to read the file, if we found a new line or we reach the end of the file we break the loop, if not we append the byte to the content variable.

If we run test, everything should be working fine.

```bash
go run test
```


### Compact data from file

At this moment we have a problem, if we use the Set function multiple times with the same key, the value will be append to the file and multiple entries with the same key will be created, we need to fix this problem, although the map will always return the last value of the key, we need to clean the file for old entries.


We need some refactor , first create a Compact function on engine.go

```go

const Seconds = 5

func (e *Engine) CompactFile() {
	for {
		time.Sleep(time.Duration(Seconds) * time.Second)
		fmt.Println("Compacting file...")
		e.mu.Lock()

		tempFile, err := os.OpenFile("temp.txt", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println("Error creating backup file:", err)
			e.mu.Unlock()
			continue
		}

		_, err = io.Copy(tempFile, e.file)
		if err != nil {
			fmt.Println("Error copying file contents to backup file:", err)
			e.mu.Unlock()
			tempFile.Close()
			continue
		}

		_, m := e.GetMapFromFile()

		err = e.file.Truncate(0)
		if err != nil {
			fmt.Println(err)
			e.mu.Unlock()
			continue
		}

		for k, v := range m {
			e.setRaw(k, v)
		}

		e.file.Seek(0, 0)
		e.mu.Unlock()
		tempFile.Close()
	}
}
```

This function will executes as a background job using a goroutine,  for example

```bash
go e.CompactFile()
```

This function will execute every 5 seconds,  first we create a new file called temp.txt, this file will be used as a backup of the original file, we copy the content of the original file to the temp file, after that we get the map of the file using the GetMapFromFile function, after that we truncate the file, after that we loop over the map and use the setRaw function to write the data to the file, finally we move the cursor to the start of the file and unlock the mutex for future uses, if we have any error we must unlock the mutex and continue with the loop.


Add function GetMapFromFile in engine.go

```go

type Item struct {
	Key    string
	Value  string
}

func (c *Engine) GetMapFromFile() ([]Item, map[string]string) {
	m := make(map[string]string)
	i := []Item{}

	_, err := c.file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return i, m
	}

	scanner := bufio.NewScanner(c.file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, keyValueSeparator)
		if len(parts) >= 2 {
			m[parts[0]] = parts[1]
			i = append(i, Item{
				Key:   parts[0],
				Value: parts[1],
			})
		}
	}

	return i, m
}
```

We created a Item struct for the key, value data,  on the GetMapFromFile function we create a map and a slice of Item, after that we move the cursor to the start of the file, after that we use a scanner to read the file line by line, we split the line by the space separator and save the key, value on the map and the slice of Item, finally we return the slice of Item and the map for later uses.

add this functions and refactor code on engine.go

```go
func (e *Engine) Set(key string, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if strings.Contains(key, " ") {
		return fmt.Errorf("key cannot contain spaces")
	}

	return e.setRaw(key, value)
}

func (e *Engine) setRaw(key string, value string) error {
	offset, err := e.saveToFile(key, value)
	if err != nil {
		return err
	}

	e.setKey(key, offset)
	return nil
}

func (e *Engine) setKey(key string, value int64) {
	e.data[key] = value
}

func (c *Engine) saveToFile(key string, value string) (int64, error) {
	offset, err := c.file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		return 0, err
	}

	_, err = c.file.WriteString(key + keyValueSeparator + value + "\n")
	if err != nil {
		fmt.Println("Error appending text:", err)
		return 0, err
	}

	return offset, nil
}
```

On set function we check if the key contains spaces, we must do that because we use a blank space as a separator of key, value on the file, 
after we call a function call setRaw, that uses other functions .

saveToFile: this function save data in file and return the offset of the key on the file.
setKey: this function save the key and value offset on the map.

To check this create a new test

engine_test.go

```go
func TestEngine_Compact(t *testing.T) {
	os.Remove("data.txt")
	v1 := "latestvalue1"
	v2 := "latestvalue2"
	e, _ := NewEngine()
	e.Set("key1", "value1")
	e.Set("key2", "value2")
	e.Set("key1", v1)
	e.Set("key2", v2)
	e.Set("key3", "value3")

	go e.CompactFile()

	time.Sleep((Seconds + 3) * time.Second)
	if len(e.GetFileContent(e.file)) != 3 {
		t.Errorf("Expected %d, but got %d", 3, len(e.GetFileContent(e.file)))
	}

}
```

First we remove the data.txt file, later we will fix that and uses another specific file for testing, 
After we set some repeated with diferent values, this will create a file wit this data

```
key1 value1
key2 value2
key1 latestvalue1
key2 latestvalue2
key3 value3
```

we see that we have key and key repeated, in order to clean that we call the CompactFile function using a goroutine, this runs every 5 seconds.
we wait for that to run and check count of lines/data on the data.file.

that should be 3, and have this data.


```
key1 latestvalue1
key2 latestvalue2
key3 value3
```

Run test

```
go test

```

### Restore data from file on restore/start

At this moment the project works fine, we can set, get value and run tests withour a problem, but we have a problem, if we restart or the server/process crash, we lost all the data on memory, especifically on the map.
we need to restore the data that is saved on the filed to the map.

add this function on engine.go

```go
func (e *Engine) Restore() {
	e.mu.Lock()
	defer e.mu.Unlock()

	items, _ := e.GetMapFromFile()

	for _, v := range items {
		e.setKey(v.Key, v.Offset)
	}
}

func (c *Engine) Close() {
	c.file.Close()
}


```

This function read data from database file and get a map calling the function GetMapFromFile, after that we loop over the map and save the key, value on the map,  also add a function that close the file descriptor to prevent memory leaks.

We need to call the function restore after get a Engine object, add this tests on engine_test.go

```go
func TestEngine_Restore(t *testing.T) {
	os.Remove("data.txt")
	e, _ := NewEngine()

	e.Set("key1_restore", "value1")
	e.Set("key2_restore", "value2")

	e.Close()

	e, _ = NewEngine()
	e.Restore()
	k, _ := e.Get("key1_restore")

	if k != "value1" {
		t.Errorf("Expected %s, but got %s", "value1", k)
	}
}
```
in this test we remove the data.txt file, after that we set some values, close the file and create a new Engine object (this simulates the creation of a new instance after crash), after that we call the Restore function and get the value of a key, if the value is not the expected return an error.
Check this test also removing the call to e.Restore, you should see an error because the key not exists on the map.



### Delete item
Next feature will be an option to delete keys, the API usage is this

```go
e := NewEngine()
e.Set("foo", "bar")
e.Delete("foo")
```

We need to create another file in order to track all keys that need to be remove, this job will be execute in an asyncronous way, we will use a goroutine to do that.

We need to make updates on engine.go

```go
type Engine struct {
	data       map[string]int64
	file       *os.File
	fileDelete *os.File
	mu         sync.Mutex
	muDelete   sync.Mutex
}

var keyValueSeparator = " "

func NewEngine() (*Engine, error) {
	file, err := os.OpenFile("data.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file data:", err)
		return nil, err
	}

	fileDelete, err := os.OpenFile("delete.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file delete:", err)
		return nil, err
	}

	return &Engine{
		data:       make(map[string]int64),
		file:       file,
		fileDelete: fileDelete,
		mu:         sync.Mutex{},
	}, nil
}

```

First we add a new property on Engine struct called fileDelete, this is a pointer to a os.File, this file will be used to save the keys that need to be deleted, we also add a new mutex called muDelete, this is needed to prevent concurrency problems when we write data to the file.


Also add this new functions

```go
func (e *Engine) Delete(key string) error {
	e.muDelete.Lock()
	defer e.muDelete.Unlock()
	_, err := e.fileDelete.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking file:", err)
		return err
	}

	_, err = e.fileDelete.WriteString(key + "\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.m, key)

	return nil
}

```

In this function we fist lock our delete file mutex, after that we move the cursor to the end of the file and write the key that we obtaint for the parameter.
If everything works fine we delete the key from the map, using the other mutex.
We use defer keyword to unlock the mutex, this assures that the unlock will be called after the function finish.

Add test for this 

```go

func TestEngine_DeleteKey(t *testing.T) {
	os.Remove("data.txt")
	os.Remove("remove.txt")
	e, _ := NewEngine()

	e.Set("key1_delete", "value1")
	e.Set("key2_delete", "value2")

	err := e.Delete("key1_delete")
	if err != nil {
		panic(err)
	}

	k, _ := e.Get("key1_delete")

	if k != "" {
		t.Errorf("Expected %s, but got %s", "", k)
	}

	if len(e.GetFileContent(e.fileDelete)) != 1 {
		t.Errorf("Expected %d, but got %d", 1, len(e.GetFileContent(e.file)))
	}
}

```
We add remove for the delete file in order to have a clean state on tests, 
after that we set some values, delete one of them and check that the key not exists on the map, lastly check that the delete file has one entry.

At the moment we delete the item for the map and add a entry on the delete.txt file,  but at the moment we are not doing anyting with the data.txt file, for that matter we need to create a couple of function that resolves that.

```go

const secondsDelete = 5
func (e *Engine) DeleteFromFile() {
	for {
		time.Sleep(secondsDelete * time.Second)
		fmt.Println("Deleting from file...")
		e.muDelete.Lock()

		_, err := e.fileDelete.Seek(0, 0)
		if err != nil {
			fmt.Println(err)
			e.muDelete.Unlock()
			continue
		}

		scanner := bufio.NewScanner(e.fileDelete)

		content := []string{}
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				content = append(content, line)
			}
		}

		err = e.deleteKeyFromFile(content)
		if err != nil {
			fmt.Println(err)
			e.muDelete.Unlock()
			continue
		}

		err = e.fileDelete.Truncate(0)
		if err != nil {
			fmt.Println(err)
			e.muDelete.Unlock()
			continue
		}

		e.muDelete.Unlock()
	}
}
```

This function runs on background and will be called using a goroutine, 
we create a variable secondsDelete that will be used to configure the time that the function will wait to run again,
After we create a loop that uses the mutex created for the delete file, 
read the file line by line and save the content on a slice of string, after  we call a function called deleteKeyFromFile ( next to analyse ) that will make the changes in order to delete the keys found on the data.txt
finally we truncate the delete file and unlock the mutex.

```go
func (c *Engine) deleteKeyFromFile(keys []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var bs []byte
	buf := bytes.NewBuffer(bs)

	scanner := bufio.NewScanner(c.file)
	for scanner.Scan() {
		l := scanner.Text()

		parts := strings.Split(l, keyValueSeparator)
		if len(parts) >= 2 {
			found := false
			for _, k := range keys {
				if parts[0] == k {
					found = true
					break
				}
			}

			if !found {
				buf.WriteString(l + "\n")
			}
		}
	}

	_, err = c.file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = c.file.Truncate(0)
	if err != nil {
		fmt.Println(err)
		return err
	}


	_, err = buf.WriteTo(c.file)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
```
This is a long function,  we explain step by step what we do here.
the keys parameter is a slice of string that contains the keys that need to be deleted, 
we lock the mutex of the data.txt file, after that we move the cursor to the start of this file, 
we  also create a buffer variable of bytes, this will be used to get the items that not need to be deleted,
we check that looping on the keys parameter and validating key of the data.txt,
if variable found is false we use the writeString function to write the line to the buffer, otherwise we continue with the loop.

After that we have the variable buffer with the data that not be deleted, we move the cursor to the start of the file, truncate the file and finally copy the content of the buffer on the data.txt.

add this test

```go
func TestEngine_DeleteKeyFromFile(t *testing.T) {
	os.Remove("data.txt")
	os.Remove("delete.txt")
	e, _ := NewEngine()

	e.Set("key1_delete", "value1")
	e.Set("key2_delete", "value2")
	e.Set("key3_delete", "value3")

	e.deleteKeyFromFile([]string{"key2_delete", "key3_delete"})

	if len(e.GetFileContent(e.file)) != 1 {
		t.Errorf("Expected %d, but got %d", 1, len(e.GetFileContent(e.file)))
	}
}
```

This test create some keys and call the deleteKeyFromFile function with two keys that need to be deleted, after that we check that the data.txt file has only one entry.
























2

Add file  append to persists  data 

use iteration to search values

use hash - byte offset to get value from file on O(1)

tests


restart , restore items

Delete item

Service HTTP db

save data files on user/.config/  folder