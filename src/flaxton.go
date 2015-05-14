package main

import (
    "fxConsole"
    "os"
    "runtime"
)

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    fxConsole.RunArguments(os.Args)
}
