package main

import "fmt"

// declare types for the AppStatus and ReplicaResponse
// call /app/status to get CPU and replica count.
// calculate the replica new count in a way that CPU <.80
//		inc replica will dec CPU and dec replica will inc CPU
// call /app/replicas to PUT the new replica count

func main() {

	fmt.Println("Idea to Code")
}
