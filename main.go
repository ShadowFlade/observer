package main
import (
    "fmt"
    "github.com/shirou/gopsutil/v3/process"
)

func main() {
    processes, _ := process.Processes()
    for _, process := range processes {
        name, _ := process.Name()
        fmt.Println(name)
    }
}
