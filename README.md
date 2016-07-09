# Golang HA Stack
[![Build Status](https://travis-ci.org/fsamin/go-hastack.svg?branch=master)](https://travis-ci.org/fsamin/go-hastack)

Share a thread-safe FIFO Stack across multiple hosts. HA Stack use Redis as backend and helps you to push and pop elements in your stack.

## Open or connect to a Stack
Each stack have a logical name, a is accessible through redis.
```
    stack, err := hastack.Connect("mystack", "localhost:6379", "", 3, 100)
```


## Push
```
stack, _ := hastack.Connect("mystack", "localhost:6379", "", 3, 100)

var data myData{
        ...,
}

s.Push(data)
```

## Pop
```
stack, _ := hastack.Connect("mystack", "localhost:6379", "", 3, 100)
    var data myData{}
err := s.Pop(&data) //can be nil is there is no data in the stack
```

## Blocking Pop
```
stack, _ = hastack.Connect("mystack", "localhost:6379", "", 3, 100)

var data myData{}

err := s.BPop(&data) //This will block the thread until data can be poped
```
