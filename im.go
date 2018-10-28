package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type imageMagic struct {
	requestChan chan imRequest
}

type imRequest struct {
	filename string
	w        io.Writer
	errChan  chan error
}

func initConvert() *imageMagic {
	im := new(imageMagic)
	im.requestChan = make(chan imRequest)

	go func(im *imageMagic) {
		for {
			select {
			case data := <-im.requestChan:
				//create temp file
				tempFile, err := ioutil.TempFile(os.TempDir(), "unoconv-api")
				if err != nil {
					data.errChan <- err
				}
				tempFileName := tempFile.Name() + ".pdf"
				os.Rename(tempFile.Name(), tempFileName)

				//convert to pdf
				cmd := exec.Command("unoconv", "-f", "pdf", "--stdout", data.filename)
				cmd.Stdout = tempFile
				err = cmd.Run()
				if err != nil {
					data.errChan <- err
				}

				//convert to png
				convertedName := fmt.Sprintf("%d_%d.png", time.Now().Unix(), rand.Intn(99999))
				cmd = exec.Command("convert", tempFileName, os.TempDir()+"/%04d_"+convertedName)
				err = cmd.Run()
				if err != nil {
					data.errChan <- err
				}
				os.Remove(tempFileName)

				// find and format converted files
				files, err := filepath.Glob(os.TempDir() + "/*" + convertedName)
				if err != nil {
					data.errChan <- err
				}
				for i, f := range files {
					files[i] = filepath.Base(f)
				}
				output, err := json.Marshal(files)
				if err != nil {
					data.errChan <- err
				}

				// write output
				data.w.Write(output)
				if err != nil {
					data.errChan <- err
				} else {
					data.errChan <- nil
				}
			}
		}
	}(im)
	return im
}

func (c *imageMagic) ConvertPptxToPng(filename string, w io.Writer) error {
	err := make(chan error)

	req := imRequest{
		filename,
		w,
		err,
	}
	im.requestChan <- req

	return <-err
}
