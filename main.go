package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"regexp"
	"time"
)

// sem version = v0.0.1
var regexSemVersion = regexp.MustCompile(`^v\d+\.\d+\.\d+.*`)

type image struct{ Name string }

func deleteRunner(tags chan string) {
	for t := range tags {
		if deleteDigest(t) {
			fmt.Printf("DELETED: %s\n", t)
			continue
		}
		fmt.Printf("FAILED TO DELETED: %s\n", t)
	}
}

func main() {
	images := listImages()
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(images), func(i, j int) { images[i], images[j] = images[j], images[i] })

	del := make(chan string)
	for i := 0; i < 20; i++ {
		go deleteRunner(del)
	}
	for i := range images {
		fmt.Printf("%s\n", images[i].Name)
		tags := listTags(images[i].Name)
		for t := range tags {
			img := fmt.Sprintf("%s@%s", images[i].Name, tags[t].Digest)
			fmt.Println("del <- ", tags[t].Timestamp.Datetime, images[i].Name, tags[t].Tags)
			del <- img
		}
	}
}

func listImages() []image {
	out, err := exec.Command("gcloud", "container", "images", "list", "--format=json").Output()
	if err != nil {
		log.Fatal(err)
	}
	var images []image
	_ = json.Unmarshal(out, &images)
	return images
}

func deleteDigest(tag string) bool {
	// gcloud container images delete --force-delete-tags --quiet
	out, err := exec.Command("gcloud", "container", "images", "delete", "--force-delete-tags", "--quiet", tag).CombinedOutput()
	if err != nil {
		log.Println("gcloud", "container", "images", "delete", "--force-delete-tags", "--quiet", tag)
		log.Printf("%+v: %s\n", err, out)

		return false
	}
	return true
}

func listTags(image string) []tags {
	notAfter := time.Now().AddDate(0, 0, -90).Format("2006-01-02")
	out, err := exec.Command(
		"gcloud", "container", "images", "list-tags", image,
		"--format", "json",
	).Output()
	if err != nil {
		log.Fatal(err, out)
	}

	var res, tags []tags
	_ = json.Unmarshal(out, &tags)

	for i := range tags {
		if len(tags[i].Tags) == 0 || (tags[i].Timestamp.Datetime < notAfter && !regexSemVersion.MatchString(tags[i].Tags[0])) {
			res = append(res, tags[i])
		}
	}
	return res
}

type tags struct {
	/*
	   {
	     "digest": "sha256:e0690a06355e667e436488a21b372b093e2930f09e48adcd0cde56c08dd654e1",
	     "tags": [
	       "srvVrfySwag"
	     ],
	     "timestamp": {
	       "datetime": "2020-06-08 13:00:11+02:00",
	       "day": 8,
	       "fold": 0,
	       "hour": 13,
	       "microsecond": 0,
	       "minute": 0,
	       "month": 6,
	       "second": 11,
	       "year": 2020
	     }
	   }
	*/
	Digest    string
	Tags      []string
	Timestamp struct {
		Datetime string
	}
}
