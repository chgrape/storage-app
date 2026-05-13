package cli

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	download := flag.NewFlagSet("download", flag.ExitOnError)
	// upload := flag.NewFlagSet("upload", flag.ExitOnError)

	dst := download.String("dst", "", "Destination of downloaded file")
	// src := upload.String("src", "", "Source file path")

	if len(os.Args) < 2 {
		fmt.Printf("CLI tool for the storage system. Use either 'download' or 'upload' to interact with the files")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "download":
		download.Parse(os.Args[2:])
		if *dst == "" {
			fmt.Println("Invalid path")
			os.Exit(1)
		}
		info, err := os.Stat(*dst)
		if err != nil {
			fmt.Println("Directory doesn't exist")
			os.Exit(1)
		}
		if info.IsDir() == false {
			fmt.Printf("%s is not a directory\n", *dst)
			os.Exit(1)
		}
		fmt.Println(*dst)
	}
}
