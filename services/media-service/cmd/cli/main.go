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

	"github.com/chgrape/storage-app/services/media-service/internal/repository"
	"github.com/google/uuid"
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

	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed: %s", string(b))
	}

	fmt.Println(string(b))

	return nil
}

func delete(c http.Client, id int) error {
	records, err := fetchList(c)
	if err != nil {
		return err
	}

	if id < 1 || id > len(records) {
		return errors.New("invalid index")
	}
	var u uuid.NullUUID

	for i, r := range records {
		if i+1 == id {
			u.UUID = r.ID
			u.Valid = true
		}
	}

	if !u.Valid {
		return errors.New("record not found")
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:8081/delete/%s", u.UUID.String()), nil)
	if err != nil {
		return err
	}

	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed: %s", string(b))
	}

	fmt.Println(string(b))

	return nil
}

func main() {
	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)

	dst := downloadCmd.String("dst", "", "Destination of downloaded file")
	id := downloadCmd.Int("id", 0, "Id of file to be downloaded")
	deleteId := deleteCmd.Int("id", 0, "Id of file to be deleted")
	src := uploadCmd.String("src", "", "Source file path")

	if len(os.Args) < 2 {
		fmt.Println("CLI tool for the storage system\n")
		fmt.Println("Usage: tube COMMAND\n")
		fmt.Println(`Commands:
	delete			Deletes media
	upload			Uploads a media file
	download		Downloads a file to the local filesystem
	list			Lists all records of the available files in the storage system`)
		os.Exit(0)
	}

	c := http.Client{}

	switch os.Args[1] {
	case "delete":
		if len(os.Args) == 2 {
			fmt.Println("The 'delete' command is used to delete records from the storage system")
			fmt.Printf(`Usage: tube delete -id int`)
			os.Exit(0)
		}
		deleteCmd.Parse(os.Args[2:])
		if *deleteId == 0 {
			fmt.Println("Invalid id")
			os.Exit(1)
		}
		if err := delete(c, *deleteId); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

	case "upload":
		if len(os.Args) == 2 {
			fmt.Println("The 'upload' command is used to upload media files from the local filesystem to the storage system")
			fmt.Printf(`Usage: tube upload -src string`)
			os.Exit(0)
		}
		uploadCmd.Parse(os.Args[2:])
		if *src == "" {
			fmt.Println("Invalid source path")
			os.Exit(1)
		}
		info, err := os.Stat(*src)
		if err != nil {
			fmt.Println("File doesn't exist")
			os.Exit(1)
		}
		if info.IsDir() == true {
			fmt.Printf("%s is not a file\n", *src)
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
		if len(os.Args) == 2 {
			fmt.Println("The 'download' command is used to download media files from the storage system into a destination directory on the local filesystem.")
			fmt.Printf(`Usage: tube download -id int -dst string`)
			os.Exit(0)
		}
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
