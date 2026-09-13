package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/chgrape/storage-app/services/media-service/internal/repository"
	"github.com/chgrape/storage-app/shared"
	"golang.org/x/sync/errgroup"
)

type authTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type authTransport struct {
	token string
	base  http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}

func fetchList(c http.Client) ([]repository.FileRecord, error) {
	res, err := c.Get("http://localhost:8081/list")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("server returned %d: %s", res.StatusCode,
			strings.TrimSpace(string(body)))
	}

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

	record := records[id-1]

	res, err := c.Get(fmt.Sprintf("http://localhost:8081/download/%s", record.ID.String()))
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("download error %d: %v", res.StatusCode, strings.TrimSpace(string(body)))
	}

	fullpath := filepath.Join(dst, record.Filename)
	file, err := os.Create(fullpath) // #nosec G304
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully uploaded file: %s with size %s at %s\n", record.Filename, shared.FormatSize(record.Size), dst)
	return err
}

func list(c http.Client) error {
	records, err := fetchList(c)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, err = fmt.Fprintln(w, "#\tNAME\tTYPE\tSIZE\tUPLOADED")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, "-\t----\t----\t----\t--------")
	if err != nil {
		return err
	}

	for i, r := range records {
		_, err = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			i+1,
			r.Filename,
			r.MIMEType,
			shared.FormatSize(r.Size),
			r.UploadedAt.Format("2006-01-02 15:04"),
		)
		if err != nil {
			return err
		}
	}

	return w.Flush()
}

func upload(c http.Client, src string) error {
	file, err := os.Open(src) // #nosec G304
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	pipe_r, pipe_w := io.Pipe()
	writer := multipart.NewWriter(pipe_w)

	go func() {
		part, err := writer.CreateFormFile("file", filepath.Base(src))
		if err != nil {
			pipe_w.CloseWithError(err)
			return
		}
		if _, err = io.Copy(part, file); err != nil {
			pipe_w.CloseWithError(err)
			return
		}
		if err = writer.Close(); err != nil {
			pipe_w.CloseWithError(err)
			return
		}
		pipe_w.Close()
	}()

	req, err := http.NewRequest("POST", "http://localhost:8081/upload", pipe_r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-File-Size", fmt.Sprint(info.Size()))

	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed: %s", string(b))
	}

	fmt.Println(string(b))

	return nil
}

func bulk_upload(c http.Client, dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	var filepaths []string
	var totalSize int64

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && path != dir {
			return fs.SkipDir
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			filepaths = append(filepaths, path)
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return err
	}

	var wg errgroup.Group
	wg.SetLimit(4)

	start := time.Now()

	for _, path := range filepaths {
		wg.Go(func() error {
			return upload(c, path)
		})
	}

	if err := wg.Wait(); err != nil {
		return err
	}

	fmt.Printf("uploaded %d files (%s) in %s\n",
		len(filepaths),
		shared.FormatSize(totalSize),
		time.Since(start).Round(time.Millisecond),
	)

	return wg.Wait()
}

func delete(c http.Client, id int) error {
	records, err := fetchList(c)
	if err != nil {
		return err
	}

	if id < 1 || id > len(records) {
		return errors.New("invalid index")
	}

	u := records[id-1].ID

	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:8081/delete/%s", u.String()), nil)
	if err != nil {
		return err
	}

	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed: %s", string(b))
	}

	fmt.Println(string(b))

	return nil
}

func login(c http.Client, username string, password string) error {
	values := url.Values{}
	values.Add("username", username)
	values.Add("password", password)

	res, err := c.PostForm("http://localhost:8081/login", values)
	if err != nil {
		return fmt.Errorf("login failed: %v", err)
	}
	defer res.Body.Close() //nolint: errcheck

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	configPath := filepath.Join(os.Getenv("HOME"), ".tube")
	if err := os.MkdirAll(configPath, 0700); err != nil { // #nosec G703
		return err
	}
	if err := os.WriteFile(filepath.Join(configPath, "config"), body, 0600); err != nil { // #nosec G703
		return err
	}

	fmt.Println("Login successful!")
	return nil

}

func newAuthClient(token string) http.Client {
	return http.Client{
		Timeout: 10 * time.Second,
		Transport: &authTransport{
			token: token,
			base:  http.DefaultTransport,
		},
	}
}

func main() {
	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	loginCmd := flag.NewFlagSet("login", flag.ExitOnError)

	var username string
	var password string

	dst := downloadCmd.String("dst", "", "Destination of downloaded file")
	id := downloadCmd.Int("id", 0, "Id of file to be downloaded")
	deleteId := deleteCmd.Int("id", 0, "Id of file to be deleted")
	src := uploadCmd.String("src", "", "Source file path")
	bulk := uploadCmd.Bool("bulk", false, "Is this a bulk upload?")

	if len(os.Args) < 2 {
		fmt.Println("CLI tool for the storage system")
		fmt.Println("Usage: tube COMMAND")
		fmt.Println(`Commands:
	login			After inputing credentials, authenticates you with the platform
	delete			Deletes media
	upload			Uploads a media file
	download		Downloads a file to the local filesystem
	list			Lists all records of the available files in the storage system`)
		os.Exit(0)
	}

	if os.Args[1] == "login" {
		if err := loginCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("error: couldn't parse command: %v", err)
		}
		fmt.Printf("Username: ")
		fmt.Scan(&username)
		fmt.Printf("Password: ")
		fmt.Scan(&password)

		if err := login(*http.DefaultClient, username, password); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	var tokens authTokens

	config, err := os.Open(filepath.Join(os.Getenv("HOME"), ".tube/config")) // #nosec G703
	if err != nil {
		fmt.Println("unauthorized")
		os.Exit(1)
	}

	err = json.NewDecoder(config).Decode(&tokens)
	if err != nil {
		fmt.Printf("config is wrong: %v", err)
		os.Exit(1)
	}

	c := newAuthClient(tokens.AccessToken)

	switch os.Args[1] {

	case "delete":
		if len(os.Args) == 2 {
			fmt.Println("The 'delete' command is used to delete records from the storage system")
			fmt.Printf(`Usage: tube delete -id int`)
			os.Exit(0)
		}
		if err = deleteCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("error: couldn't parse command: %v", err)
		}
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
		if err = uploadCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("error: couldn't parse command: %v", err)
		}
		if *src == "" {
			fmt.Println("Invalid source path")
			os.Exit(1)
		}
		info, err := os.Stat(*src)
		if err != nil {
			fmt.Println("File doesn't exist")
			os.Exit(1)
		}

		if *bulk {
			if !info.IsDir() {
				fmt.Printf("%s is not a directory\n", *src)
				os.Exit(1)
			}
			if err := bulk_upload(c, *src); err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
		} else {
			if info.IsDir() {
				fmt.Printf("%s is not a file\n", *src)
				os.Exit(1)
			}
			if err := upload(c, *src); err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
		}

	case "list":
		if len(os.Args) > 2 {
			fmt.Printf("List doesn't accept arguments")
			os.Exit(1)
		}
		if err = listCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("error: couldn't parse command: %v", err)
		}
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
		if err = downloadCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("error: couldn't parse command: %v", err)
		}
		if *dst == "" || *id == 0 {
			fmt.Println("Invalid destination path or id")
			os.Exit(1)
		}
		info, err := os.Stat(*dst)
		if err != nil {
			fmt.Println("Directory doesn't exist")
			os.Exit(1)
		}
		if !info.IsDir() {
			fmt.Printf("%s is not a directory\n", *dst)
			os.Exit(1)
		}
		if err := download(c, *id, *dst); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	}
}
