package downloader

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	humanize "github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
)

// FromURL downloads file at downloadURl and return path to downloaded file.
func FromURL(downloadURL string) (string, error) {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", err
	}
	return FromReader(resp.Body, resp.ContentLength)
}

// FromReader downloads file from reader and return path to downloaded file.
func FromReader(reader io.Reader, length int64) (string, error) {
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}

	pr := progressReader(reader, length)
	if _, err := io.Copy(tmp, pr); err != nil {
		return "", err
	}

	tmp.Close()
	if r, ok := reader.(io.ReadCloser); ok {
		r.Close()
	}
	return tmp.Name(), nil
}

func progressReader(r io.Reader, totalSize int64) io.Reader {
	return &ioprogress.Reader{
		Reader: r,
		Size:   totalSize,
		DrawFunc: ioprogress.DrawTerminalf(os.Stdout, func(progress, total int64) string {
			progressStr := humanize.Bytes(uint64(progress))
			if total > 0 {
				return fmt.Sprintf(
					"  Downloading: %s/%s",
					progressStr,
					humanize.Bytes(uint64(total)),
				)
			}
			return fmt.Sprintf(
				"  Downloading: %s",
				progressStr,
			)
		}),
	}
}
