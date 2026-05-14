package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"github.com/chgrape/storage-app/internal/repository"
)

func fetchList(c http.Client) ([]repository.FileRecord, error) {
	res, err := c.Get("http://localhost:8081/list")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var records []repository.FileRecord
	if err := json.NewDecoder(res.Body).Decode(&records); err != nil {
		return nil, err
	}

	return records, nil
}

func download(c http.Client, id int, dst string) error {
	records, err := fetchList(c)
	if err != nil {
		return err
	}

	if id < 1 || id > len(records) {
		return errors.New("invalid index")
	}
	var record repository.FileRecord

	for i, r := range records {
		if i+1 == id {
			record = r
		}
	}
	if record == (repository.FileRecord{}) {
		return errors.New("file not found")
	}

	res, err := c.Get(fmt.Sprintf("http://localhost:8081/download/%s", record.ID.String()))
	if err != nil {
		return err
	}

	fullpath := filepath.Join(dst, record.Filename)
	file, err := os.Create(fullpath)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return err
	}

	return nil
}

func list(c http.Client) error {
	records, err := fetchList(c)
	if err != nil {
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "#\tNAME\tTYPE\tSIZE\tUPLOADED")
	fmt.Fprintln(w, "-\t----\t----\t----\t--------")

	for i, r := range records {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			i+1,
			r.Filename,
			r.MIMEType,
			strconv.FormatInt(r.Size, 10),
			r.UploadedAt.Format("2006-01-02 15:04"),
		)
	}
	w.Flush()

	return nil
}

func upload(c http.Client, src string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(src))
	if err != nil {
		return err
	}

	if _, err = io.Copy(part, file); err != nil {
		return err
	}
	writer.Close()

	res, err := c.Post("http://localhost:8081/upload", writer.FormDataContentType(), body)

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("upload failed: %s", string(b))
	}
	return nil
}

func main() {
	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)

	dst := downloadCmd.String("dst", "", "Destination of downloaded file")
	id := downloadCmd.Int("id", 0, "Id of file to be downloaded")
	src := uploadCmd.String("src", "", "Source file path")

	if len(os.Args) < 2 {
		fmt.Printf("CLI tool for the storage system. Use either 'download' or 'upload' to interact with the files")
		os.Exit(0)
	}

	c := http.Client{}

	switch os.Args[1] {
	case "upload":
		uploadCmd.Parse(os.Args[2:])
		if *src == "" {
			fmt.Println("Invalid source path or id")
			os.Exit(1)
		}
		info, err := os.Stat(*src)
		if err != nil {
			fmt.Println("Directory doesn't exist")
			os.Exit(1)
		}
		if info.IsDir() == false {
			fmt.Printf("%s is not a directory\n", *src)
			os.Exit(1)
		}
		if err := upload(c, *src); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

	case "list":
		if len(os.Args) > 2 {
			fmt.Printf("List doesn't accept arguments")
			os.Exit(1)
		}
		listCmd.Parse(os.Args[2:])
		if err := list(c); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

	case "download":
		downloadCmd.Parse(os.Args[2:])
		if *dst == "" || *id == 0 {
			fmt.Println("Invalid destination path or id")
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
		if err := download(c, *id, *dst); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	}
}
