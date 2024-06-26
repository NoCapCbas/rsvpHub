package main

import (
  "fmt"
)

type Rsvp struct {
  Name string 
  Email string
  PhoneNumber string
  WillAttend bool
}

var responses = make([]*Rsvp, 0, 10)

func main() {
  fmt.Println("TODO: add some features")
}
