package main

import (
	"flag"
	"fmt"
	"os"

	kitlog "github.com/go-kit/kit/log"
	opencnam "github.com/oliver006/go-opencnam"
	ocontext "golang.org/x/net/context"
)

func printUsage() {
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("")
	fmt.Println("cli << number>>")
	fmt.Println("")
	fmt.Println("Environment variables:")
	fmt.Println("OPENCNAM_SID, OPENCNAM_TOKEN")
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("not enough parameters")
		printUsage()
		return
	}

	c := opencnam.NewClient(
		os.Getenv("OPENCNAM_SID"),
		os.Getenv("OPENCNAM_TOKEN"),
		"",
		kitlog.NewNopLogger(),
	)

	num := flag.Arg(0)
	res, err := c.NumberInfo(ocontext.Background(), num)
	if err != nil {
		fmt.Printf("error: %s \n", err)
		printUsage()
		return
	}

	fmt.Printf("Number: %s \n", num)

	fmt.Println()
	fmt.Println("Result")
	fmt.Println("======")
	fmt.Println()
	fmt.Printf("Name:    %s\n", res.Name)
	fmt.Printf("Number:  %s\n", res.Number)
	fmt.Printf("Price:   %f\n", res.Price)
	fmt.Printf("Uri:     %s\n", res.Uri)
	fmt.Println()
}
