# KeyValue Database tutorial

In this tutorial we are going to create a simple key value database using go.
We can think in something like redis or etcd but with a very limited set of features and not ready for production.

## Nic, But Why?
Because i think to create something from scratch is a great way to learn and understand how things works under the hood, 
also if you know the basic on Golang and want to learn more about the language,  create a real  project like this is a great way to improve your skills, for that reason i am not go to deeply explaining sintax or basic concepts of the language.


### How it works
The database will be a simple key value store, we will use a hash map to store the key that map to some data,  we will use a file to persist data on disk.

We will use tests to check that everything works as expected after we add or modify the code,  this is very important to prevent bugs and have more trust in our code.
also we will have a simple http api to interact with it, we will use the standard Go library to create the http server, although use HTTP is seems like an overhead, we are going to use it because is simpler and easy to implement instead of create a custom protocol like exits on redis, mysql, etcd or another database , and is pretty easy to test.

Example using curl

We send a json payload with a key, value data

```bash
curl -X POST -H "Content-Type: application/json" -d '{"key": "mykey", "value": "from curl"}' http://localhost:8000/set
```

Example get value by key
```bash
curl http://localhost:8000/get?key=foo
```

Glossary https://github.com/ernesto27/system-design/tree/master/keyvaluedb/tutorial

- Create go project
- Create engine package
- Add tests
- Persist data on disk
- Set data on file 
- Get data from key
- Compact data from file
- Restore data from file on restore/start
- Delete item
- Create HTTP service
- Update database files path


I am not expend much time explaining things abouts golang sintax so i assume that you have go installed, you have basic knowledge of the language and at least create some basic project, if you need a refresh take a look here https://gobyexample.com/.

### Create go project
To start this project create a new folde called "keyvaluedb" go into that and create a new go module with the name keyvaluedb ( you can change this name for the name you want, if you do that remberber imports path where import local files)

```bash
mkdir keyvaluedb && cd keyvaluedb && go mod init keyvaluedb
```

Add a new file main.go and paste or write the following code to check at least that everything is working fine.

main.go
```go
package main


import "fmt"

func main() {
    fmt.Println("Hello world")
}
```

Run the project

```bash
go run main.go 
```

You should see the message Hello world print in the console, if not check your golang installation https://go.dev/doc/install.

### Create engine package

We will create a new package called engine, this package will contain all the logic of the database,  so for that create a new file engine.go and add the following code.


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

This code create a new struct called Engine, this will contain the data of the database, in this case we will use a simple map to store the data in memory, later we will add persistence on a file.

The NewEngine function create a new instance of the Engine struct, the important thing here is notice that we initialize the data map with make, this is a must because if we don't do this, the map will be nil and we will get a panic when we try to add a new key value pair.

The struct has two methods, Set and Get, Set will add a new key value pair to the map and Get will return the value of a key if exists, 
if key not exists return an error.


In order to check , add this on main.go

```go

func main() {
    e := NewEngine()
    e.Set("foo", "bar")
    value, err := e.Get("foo")
    if err != nil {
        panic(err)
    }
    fmt.Println(value)
}

```

Run the project using 

```bash
go run .
```


This create a new instance on Engine,  set a new key value pair and get the value of the key, if the key not exists, panic and print error, otherwise print the value on the console.

### Add tests
Althoug we can use main.go for test the code, is a good idea to start using tests in order to check the code in a more consice, easy and stable way, is true that in this moment seems like an overkill,  but because we will add more features to the code is a good idea to have automated tests from the beggining and check all with a single command.

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

This code create a new test function called Test_SetGetKeyValue, this function has a new instance of the Engine struct, set a new key value pair and get the value of the key, if the value is not the expected return an error, 
we use the testing native native golang package to tests, we are not going to use any external libraries.

Run tests 

```bash
go test
```


### Persist data on disk
At the moment we only store key, value on memory, that works fine, but the problem is that if we for example restart the server or if a crash happened we lost all the data that we previously saved using Set method, in order to prevent that, the key/value data will be save on a file,  this key value will be separate by a space and we differenciate items by a new line,  for example

data.txt

```
foo bar
bar foo
user1 {"id": 1, "name": "ernesto"} 
```

#### Set data on file 
We are goint to update the Set method to save the data on file, the idea is to append the data at end of the file,  using a concept called append only file, this is a common pattern used in  some databases, where information can only be added or appended and not modified or deleted.
https://en.wikipedia.org/wiki/Append-only


engine.go

```go
type Engine struct {
	data map[string]int64
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

In this code we update the Engine structs:
data: change the type of value map to a int64 instead of a string.
file: this property is a pointer to a os.File, this is used to read and write data from a file.
mu: this is for prevent concurrency problems when we write data to the file, 

In the  NewEngine function we open the file in read and write mode and configure to append data to the file when writing, if the file not exists create a new one, we also initialize the mutex property and initialize a map data structure.
In the Set function we use Lock in order to prevent problems when we write data, this is a must if we want to prevent conflicts when multiple clients try to write data to this file in the same moment, we use defer function to unlock the mutex when the Set method finish.
After we use the Seek function to move the cursor to the end of the file, this is because we need to append data to the file.
We use the WriteString function to write the key value pair to the file, we also add a new line at the end of the string, this is because we want to separate the key value pair by a new line.
Finally we now set the value of the map data with the value of the offset return by the Seek function,  remember that with this new approach we need to search the  data value on a file and not on memory.


### Get data from key
Previusly we access to data directly for the map structure on memory, because in the previous step we start to save data in file ,  we need to change the Get method.


engine.go

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

We validate if key exists on data map structure, if not exists, returns an error,
we use the Seek function to move the cursor to the offset of the key that is required by parameter on the method,
for example if we have the key-value "foo bar" saved on disk
we have to obtian the offset for key saved on the data map , plus the len of the key bar (4) plus (1) for the space separator, after that we have the cursor at the start of the value.

```go
_, err := e.file.Seek(e.data[key]+int64(len(key))+1, 0)
```
after call this Seek method, the cursor of the file is at the start of the value of the key, with our example that should be on start of "bar"


Next we create a buffer of 1 byte, this is because we need to read the file byte by byte, we also create a content variable to save the value of the key in a slice of bytes, next we use a for loop to read the file, if we found a new line or reach the end of the file we break the loop, otherwise we append the byte to the content variable,
lastly the end of the method we transform the byte slice to a string and return the value.


Because we are now returning an error on the NewEngine function we should update our tests.
engine_test.go 

```go
func Test_SetGetKeyValue(t *testing.T) {
	e, _ := NewEngine()
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

If now run tests, everything should be working fine.

```bash
go run test
```


### Compact data from file

At this moment code is working well but we have a problem with the data, if we use the Set function multiple times with the same key, the value will be append to the file and multiple entries with the same key will be created, although the map will always return the last value of the key, we need to fix this problem and clean the file for old unused entries.


We need some refactor , first create a Compact function on engine.go

```go

const Seconds = 5

func (e *Engine) CompactFile() {
	for {
		time.Sleep(time.Duration(Seconds) * time.Second)
		fmt.Println("Compacting file...")
		e.mu.Lock()

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
	}
}
```

This method will be run as a background job using a goroutine, like this 

```bash
go e.CompactFile()
```

We get the map of the file using the GetMapFromFile method,  next we truncate the original file,
after that we loop over the map data and use the setRaw function to write the data to the file, finally we move the cursor to the start of the file and unlock the mutex for future uses, if we have any error we must unlock the mutex and continue with the loop.
that works because the map data only have the latests and valid key values pairs.


Add function GetMapFromFile in engine.go

```go

type Item struct {
	Key    string
	Value  string
	Offset int64
}

func (c *Engine) GetMapFromFile() ([]Item, map[string]string) {
	m := make(map[string]string)
	i := []Item{}

	_, err := c.file.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return i, m
	}

	var totalBytesRead int64
	scanner := bufio.NewScanner(c.file)

	for scanner.Scan() {
		line := scanner.Text()
		offset := totalBytesRead
		parts := strings.Split(line, keyValueSeparator)
		if len(parts) >= 2 {
			m[parts[0]] = parts[1]
			i = append(i, Item{
				Key:   parts[0],
				Value: parts[1],
				Offset: offset,
			})
		}
	}

	return i, m
}
```

We created a Item struct for the key, value data, on the GetMapFromFile method we create a map and a slice of Item, after that we move the cursor to the start of the file and we use a scanner to read the file line by line, we split the line by the space separator and save the key, value on the map and the slice of Item, finally we return the slice of Item and the map for later uses.

We need to set the offset value on the Item struct, this is the value of the key map struct the we use to get items from database, we created a totalBytesRead variable that will be used to calculate the offset of the key, to obtain that we sum the len of the line plus one for the new line character, on every iteration of the loop we set the offset value and create a Item struct.


update methods in engine.go

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

On set function we check if the key contains spaces, we must do that because we use a blank space as a separator of key/value on the file, next we call a method call setRaw, that uses this other methods.

saveToFile: this function save data in file and return the offset of the key on the file.
setKey: this function save the key and value offset on the map.


Add this test on engine_test.go

```go
func (c *Engine) GetFileContent(f *os.File) []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := f.Seek(0, 0)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}

	scanner := bufio.NewScanner(f)

	var content []string
	for scanner.Scan() {
		line := scanner.Text()
		content = append(content, line)
	}

	return content
}
```
We also add a method call GetFileContent, this is a helper function that we use to get the content of a file, we use this function on the tests.

To check all this changes create a new test

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
After we set some repeated keys with diferent values, this will create a file with this data

```
key1 value1
key2 value2
key1 latestvalue1
key2 latestvalue2
key3 value3
```

we see that we have some keys repeated, in order to clean that we call the CompactFile function using a goroutine, this runs every 5 seconds.
we wait for that to run and check count of lines/data on the data.txt file, 
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
At this moment the project works fine, we can set, get value and run tests, but we have something to pay attention, if we restart or the server/process crash, we lost all the data on memory, especifically key value saved on map structure.
we need to restore the data that is saved on the file on the map memory in order to Set method workd normally.

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

This method read data from database file and get a map calling the method GetMapFromFile, after that we loop over the map and save the key, value on the map using the setKey method that we create before, like in all methods that write or read a file we use a mutex to prevent concurrency problems.

The Close method close the file when program finish, this is necessary to prevent memory leaks.

Add this tests on engine_test.go.

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
in this test we remove the data.txt file, after that we set some values, close the file and create a new Engine object (this simulates the creation of a new instance after a crash), next we call the Restore function and get the value of a key, if the value is not the expected return an error.




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

First we add a new property on Engine struct called fileDelete, this is a pointer to a os.File, this will be used to save the keys that need to be deleted, we also add a new mutex called muDelete.

Add this new method on engine.go

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
	delete(e.data, key)

	return nil
}

```

In this function we fist lock our delete file mutex, after we move the cursor to the end of the file and write the key that we obtain from the parameter, is the same approach that we use when we save key/value data on the other file.
If everything works fine we delete the key from the map, here we must use the other mutext because we are making changes on the map structure of data,


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
We add remove call for the delete file in order to have a clean state on tests, 
after that we set some values, delete one of them and check that the key not exists on the map, lastly check that the delete file has only one entry.

At the moment we delete the item for the map and add a entry on the delete.txt file,  but  we are not doing anyting with the data.txt file, for that matter we need to create a couple of function that resolves that.

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

This method runs on background and will be called using a goroutine, 
we create a variable secondsDelete that will be used to configure the time that the method will wait to run again,
after we create a loop that uses the mutex created for the delete file, 
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
This method recieved a keys parameters, this is a slice of string that contains the keys that need to be deleted, 
we lock the mutex of the data.txt file, after that we move the cursor to the start of this file, 
we  also create a buffer variable of bytes, this will be used to get the items that not need to be deleted,
we check that looping on the keys parameter and validating key of the data.txt,
if variable found is false we use the writeString function to write the line to the buffer, otherwise we continue with the loop.

After that we have the variable buffer with the data that is not mark to be deleted, we move the cursor to the start of the file, truncate that and finally copy the content of the buffer on the data.txt.

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

This test create some keys and call the deleteKeyFromFile function with two keys that need to be deleted, after that we check that the delete.txt file has only one entry.



### Create HTTP service
At the moment we are testing the code of engine.go using tests, that is great because you can check all the feautures d the project with a single command and gain more confidence in your code when you need to make changes or updates,  but we do not have any service or way to interact with our database from external clients, in order to change that, we are going to create a simple http server, we will expose three endpoint to create, get and delete keys on the database.

On main.go add this code.

```go
package main

import (
	"fmt"
	"net/http"
)

func handlerSet(w http.ResponseWriter, r *http.Request)    {}
func handlerGet(w http.ResponseWriter, r *http.Request)    {}
func handlerDelete(w http.ResponseWriter, r *http.Request) {}

var e *Engine

func main() {
	var err error
	e, err = NewEngine()
	if err != nil {
		panic(err)
	}
	defer e.Close()
	e.Restore()

	go e.CompactFile()
	go e.DeleteFromFile()

	http.HandleFunc("/set", handlerSet)
	http.HandleFunc("/get", handlerGet)
	http.HandleFunc("/delete", handlerDelete)

	address := ":8080"

	fmt.Printf("Server is listening on http://localhost%s\n", address)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

```

In this code we create three functions that will be used as handlers for the http server, at the moment are empty, we will change that in the future.
On main function we first create a instance of the Engine DB ( we use a global e variable to get a more easy access on the handlers) and panic if there is an error, 
after we call the restore function, we have to run this with the service start in order to recover for crashes or error,  this get the data from the file and save it on the map data structure,   
next we call the CompactFile, and DeleteFromFile on background using a goroutine,
CompactFill will remove duplicate values on the file database, 
DeleteFromFIle will remove keys-value from the file database,
Lastly we create the routes on the http server and start on port 8080.



#### Update handlers

We must update the handlers create previously on main.go,  the basic idea is to use the methods of the engine file and respond a JSON in an endpoint http.

main.go
```go

type RequestPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ResponseJson struct {
	Status  string `json:"key"`
	Message string `json:"value"`
}

func handlerSet(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		var rp RequestPayload

		err = json.Unmarshal(body, &rp)
		if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusBadRequest)
			return
		}

		err = e.Set(rp.Key, rp.Value)
		if err != nil {
			responseJSON(w, ResponseJson{
				Status:  "error",
				Message: err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		responseJSON(w, ResponseJson{
			Status:  "success",
			Message: "Key value pair saved successfully.",
		}, http.StatusOK)
	} else {
		fmt.Println("Invalid request method.")
		fmt.Fprintf(w, "Invalid request method.")
	}
}

func responseJSON(w http.ResponseWriter, data interface{}, status int) {
	d, err := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server errror"))
		return
	}

	w.WriteHeader(status)
	w.Write(d)
}

```

We created a few new structs, RequestPayload is needed to decode the JSON payload sent by the client
and Response is used to send a JSON response to the client.
In the function handlerSet we first check the method of the request, if is not a POST request we return an error, 
We read the Body of the request and decode the JSON payload to the RequestPayload struct and check for errors,
after that we call the Set function of the engine file,  we now return a response using a function called responseJSON the we will check next.

ResponseJSON is a helper function that we used to prevent repeat code multiple times, this function receive a interface (basically any struct is valid), if something goes wrong we return a 500 error,
otherwise we set a status code with the value received from the argument, set the content type to application/json and write the JSON response to the client.

We can test this using curl

```bash
go run .

curl -X POST -H "Content-Type: application/json" -d '{"key": "mykey", "value": "myvalue"}' http://localhost:8080/set
```

This should return a json success message and save data on the data.txt database file


#### Get handler

```go

func handlerGet(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		key := r.URL.Query().Get("key")
		value, err := e.Get(key)
		if err != nil {
			responseJSON(w, ResponseJson{
				Status:  "error",
				Message: err.Error(),
			}, http.StatusNotFound)
			return
		}

		responseJSON(w, RequestPayload{
			Key:   key,
			Value: value,
		}, http.StatusOK)
	} else {
		fmt.Println("Invalid request method.")
		fmt.Fprintf(w, "Invalid request method.")
	}
}
```

In this function we first check the method of the request, if is not a GET request we return an error like in the set function, 
next we get the key from the query params and uses the Get function of the engine instance, if something goes wrong we return a 404 error, otherwise we return a JSON response with the key and value.

Test with curl 

```bash
go run .
curl http://localhost:8080/get?key=mykey
```


#### Delete handler

```go
func handlerDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		key := r.URL.Query().Get("key")
		err := e.Delete(key)
		if err != nil {
			responseJSON(w, ResponseJson{
				Status:  "error",
				Message: err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		responseJSON(w, ResponseJson{
			Status:  "success",
			Message: "Key deleted successfully.",
		}, http.StatusOK)
	} else {
		fmt.Println("Invalid request method.")
		fmt.Fprintf(w, "Invalid request method.")
	}
}
```
This function is very similar to the Get function, the only difference is that we use the Delete function of the engine instance instead and change the message response

test in curl 

```bash
go run .
curl -X DELETE http://localhost:8080/delete?key=mykey
```


### Update database files path

Currently we have the project running using tests and also exposing a http server, this works we have some problems with this approach.
We are using the same files for tests and for the server, so when we run a test we are modifying and deleting the data created via API endpoints.
The file is save relative to the current path in which the project is running, this is not a good thing,  because if we start the service from another path we will create a new Database file.

To fix that we we use this approach, 
in tests, we must define the name of the files data and delete and for the server and not tests instances we must use the path of the current user running the project, and save this on a folder called keyvaluedb the lives in .config home user folder.

The .config values is uses for multiple applications to save data,  for example discord, chrome, VirtualBox, etc save files on this place.

Check in your machine with this command

```bash
ls ~/.config
```


update engine.go

```go
type Config struct {
	FileData   string
	FileRemove string
}

var keyValueSeparator = " "

func NewEngine(cfg Config) (*Engine, error) {
	if cfg.FileData == "" && cfg.FileRemove == "" {
		configFolderPath, err := getConfigFolder()
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if _, err := os.Stat(configFolderPath); os.IsNotExist(err) {
			err := os.Mkdir(configFolderPath, 0700)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		}

		cfg.FileData = configFolderPath + "/" + "data.txt"
		cfg.FileRemove = configFolderPath + "/" + "delete.txt"
	}

	file, err := os.OpenFile(cfg.FileData, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file data:", err)
		return nil, err
	}

	fileDelete, err := os.OpenFile(cfg.FileRemove, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file delete:", err)
		return nil, err
	}

	return &Engine{
		data:       make(map[string]int64),
		file:       file,
		fileDelete: fileDelete,
		mu:         sync.Mutex{},
		muDelete:   sync.Mutex{},
	}, nil
}

func getConfigFolder() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	homeDir := currentUser.HomeDir
	configFolder := ".config/keyvaluedb"
	configFolderPath := filepath.Join(homeDir, configFolder)
	return configFolderPath, nil
}

```

We add a Config struct , this will help us for set file data and delete on testing,  NewEngine function now receives a config struct as a parameter,  if the file data and delete is not set ( empty strings default ),  we use the current user path /.config/keyvaluedb to save data of the application, 
we use the function getConfigFolder to get the path of the current user, on NewEngine we check if that folder exists and if not we create it, after that we set the path of the files using the config struct.

We need to update the tests to check this.

```go

var cfg = Config{
	FileData:   "data.txt",
	FileRemove: "delete.txt",
}

func Test_SetGetKeyValue(t *testing.T) {
	e, _ := NewEngine(cfg)
	e.Set("test", "data")
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

We need to update all the calls to NewEngine in order use the cfg variable.

Update main.go

```go
e, err = NewEngine(Config{})

```

In this case we pass a empty config struct, we are going to use that for the http server.

Run the server and save some data

```bash
go run .
curl -X POST -H "Content-Type: application/json" -d '{"key": "mykey", "value": "bar"}' http://localhost:8080/set
```

After that we can check the data.txt file on the path of the current user

```bash
cat ~/.config/keyvaluedb/data.txt
```

Full code github 
https://github.com/ernesto27/system-design/tree/master/keyvaluedb/tutorial


### Conclusion

In this tutorial we finished a very simple implementation of a key value store, although you must use in production a real and stable Database,  is a good exercise to understand how things are made from scatch and understand concepts that help us how to choose hour next database for some project.  
we used a lot of concepts like concurrency, mutex, file write/read, http server, etc that would be very useful in other projects.



