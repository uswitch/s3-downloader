package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	endpointUrl = kingpin.Flag("endpoint-url", "Endpoint URL to connect to S3").Default("").String()
	region = kingpin.Flag("region", "Region to connect to").Default("eu-west-1").String()
	objects = kingpin.Arg("objects", "Format: s3://<bucket>/<path>:<path>(?:<mode>)").Required().Strings()
)

func main() {
	kingpin.Parse()

	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint: endpointUrl,
		Region: region,
	}))

	s3client := s3.New(sess)

	for _, object := range *objects {
		parts := strings.Split(object, ":")
		objectUrlString := parts[0] + ":" + parts[1]
		outPath := parts[2]

		var fileMode os.FileMode

		if len(parts) > 3 {
			if modeNum, err := strconv.ParseUint(parts[3], 8, 64); err != nil {
				panic(err)
			} else {
				fileMode = os.FileMode(modeNum)
			}
		} else {
			fileMode = os.FileMode(0644)
		}

		objectUrl, err := url.Parse(objectUrlString)

		if err != nil {
			panic(err)
		}

		fmt.Printf("Bucket: %s  Path: %s -> %s [%s]\n", objectUrl.Host, objectUrl.Path, outPath, fileMode.String())

		response, err := s3client.GetObject(&s3.GetObjectInput{
			Bucket: &objectUrl.Host,
			Key: &objectUrl.Path,
		})

		if err != nil {
			panic(err)
		}

		objectData, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()

		if err != nil {
			panic(err)
		}

		if err = os.MkdirAll(filepath.Dir(outPath), os.FileMode(0755)); err != nil {
			panic(err)
		}

		if err = ioutil.WriteFile(outPath, objectData, fileMode); err != nil {
			panic(err)
		}
	}
}
