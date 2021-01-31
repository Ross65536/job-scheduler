package main

import (
	"github.com/ros-k/job-manager/server/internal"
)

func main() {
	internal.AddUser("user1", "XlG15tRINdWTAm5oZ/mhikbEiwf75w0LJUVek0ROhY4=")
	internal.AddUser("user2", "oAtCvE6Xcu07f2PmjoOjq8kv6X2XTgh3s37suKzKHLo=")
	internal.StartServer()
}
