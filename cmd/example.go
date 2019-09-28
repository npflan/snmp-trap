package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/types"
)

type arrayStrings []string

var modules arrayStrings
var paths arrayStrings
var debug bool
var oid string

func (a arrayStrings) String() string {
	return strings.Join(a, ",")
}

func (a *arrayStrings) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func main() {
	flag.BoolVar(&debug, "d", false, "Debug")
	flag.Var(&modules, "m", "Module to load")
	flag.Var(&paths, "p", "Path to add")
	flag.StringVar(&oid, "o", "", "oid to lookup")
	flag.Parse()
	Init()
	oid := oid
	if oid == "" {
		panic("No oid")
	} else {
		Subtree(oid)
	}

	Exit()
}

func Init() {
	gosmi.Init()

	for _, path := range paths {
		gosmi.AppendPath(path)
	}

	for _, module := range modules {
		moduleName, err := gosmi.LoadModule(module)
		if err != nil {
			fmt.Printf("Init Error: %s\n", err)
			return
		}
		if debug {
			fmt.Printf("Loaded module %s\n", moduleName)
		}
	}

	if debug {
		path := gosmi.GetPath()
		fmt.Printf("Search path: %s\n", path)
		loadedModules := gosmi.GetLoadedModules()
		fmt.Println("Loaded modules:")
		for _, loadedModule := range loadedModules {
			fmt.Printf("  %s (%s)\n", loadedModule.Name, loadedModule.Path)
		}
	}
	Exit()
}

func Exit() {
	gosmi.Exit()
}

func Subtree(oid string) {
	node, err := gosmi.GetNodeByOID(types.OidMustFromString(oid))
	if err != nil {
		fmt.Printf("Subtree Error: %s\n", err)
		return
	}

	subtree := node.GetSubtree()
	fmt.Println(subtree[0].Node.Oid)
	fmt.Println(subtree[0].Node.OidLen)
	fmt.Println()
	fmt.Println()
	fmt.Println(subtree[0].Node.Name)
	fmt.Println(subtree[0].Node.Description)
	fmt.Println(subtree[0].Node.Access)
	fmt.Println(subtree[0].Node.Status)
	fmt.Println()
	fmt.Println()

	// jsonBytes, _ := json.Marshal(subtree)
	// os.Stdout.Write(jsonBytes)
}
