curl -X POST -H "Content-Type: application/json" -d '{"key": "mykey", "value": "from curl"}' http://localhost:8080/set

curl -X DELETE http://localhost:8080/delete?key=111
