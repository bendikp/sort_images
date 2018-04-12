package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rwcarlsen/goexif/exif"
	"gopkg.in/h2non/filetype.v1"
)

// Image holds information about an image
type Image struct {
	name  string
	path  string
	year  int
	month int
	day   int
}

func main() {
	sourceFlag := flag.String("source", "./", "Folder with unorganised images. Must be an existing folder.")
	destFlag := flag.String("destination", "", "Folder to move the images into. The folder is created if it does not exist. (Required)")
	dryRunFlag := flag.Bool("dry-run", true, "Set to false to actually make changes.")
	required := []string{"destination"}
	flag.Parse()

	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })

	// make sure the required flag are set
	for _, req := range required {
		if !seen[req] {
			fmt.Fprintf(os.Stderr, "Missing required -%s argument/flag\n", req)
			flag.PrintDefaults()
			os.Exit(2)
		}
	}

	// check if source is a folder that exists
	file, err := os.Open(*sourceFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Something went wrong while trying to open %s:\n%s\n", *sourceFlag, err.Error())
		os.Exit(1)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Something went wrong while reading info about %s: %v\n", *sourceFlag, err.Error())
		os.Exit(2)
	}
	if !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "%s is not a directory!\n", *sourceFlag)
		os.Exit(2)
	}

	// Suppress warnings about variables not being used
	_ = *sourceFlag
	_ = *destFlag
	_ = *dryRunFlag

	// Traverse source and find all images in the directory
	images, err := getImages(*sourceFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking the path %q: %v\n", *sourceFlag, err)
	}

	// Find out when each image is taken
	for idx := range images {
		err := getDate(&images[idx])
		if err != nil {
			continue
		}
	}

	// Create folder structure
	for _, image := range images {
		newPath := filepath.Join(*destFlag, fmt.Sprintf("%04d/%02d/%02d", image.year, image.month, image.day))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			err := os.MkdirAll(newPath, 0755)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Something went wrong while creating the directory %s: %v\n", newPath, err)
				os.Exit(1)
			}
		}
	}

	// Copy images to date folders
	for _, image := range images {
		newPath := filepath.Join(*destFlag, fmt.Sprintf("%04d/%02d/%02d/%s", image.year, image.month, image.day, image.name))
		err := copyFile(image.path, newPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Something went wrong while copying the image %s to %s: %v\n", image.path, newPath, err)
			continue
		}
	}
}

func getDate(image *Image) error {
	f, err := os.Open(image.path)
	defer f.Close()
	if err != nil {
		return err
	}

	x, err := exif.Decode(f)
	if err != nil {
		return err
	}

	date, err := x.DateTime()
	if err != nil {
		return err
	}

	image.year = date.Year()
	image.month = int(date.Month())
	image.day = date.Day()
	return nil
}

func getImages(path string) ([]Image, error) {
	var images []Image
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		head := make([]byte, 261)
		file.Read(head)
		if filetype.IsImage(head) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			images = append(images, Image{
				path: absPath,
				name: info.Name(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return images, nil
}

func copyFile(src, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !sfi.Mode().IsRegular() {
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return err
		}
	}
	err = copyFileContents(src, dst)
	return err
}

func copyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return err
}
