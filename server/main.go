package main

import (
	"github.com/ros-k/job-manager/server/internal"
)

func main() {
	internal.InitializeUsers()
	internal.StartServer()
}
