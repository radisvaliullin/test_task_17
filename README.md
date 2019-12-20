# test_task_17
The Test Task

## Description
see spin repo readme [README.md](https://github.com/spin-org/thermomatic/blob/master/README.md)  
or
```
docs/thermomatic.md
```

## Run server
```
go run cmd/server/main.go
```

## Test
```
go test ./... -cover
```

## Bench
```
cd internal/server
go test -bench=.
```

#### HTTP Server
http://0.0.0.0:1338
```
GET /readings/:imei
response:
{"imei":"490154203237518","Status":"online","reading":{"Temp":0,"Alt":0,"Lat":0,"Lon":0,"BattLev":0},"time":1576833027211679121}

GET /status/:imei
response:
{"imei":"490154203237518","Status":"online"}
```
