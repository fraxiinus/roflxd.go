package main

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/fraxiinus/roflxd/glr"
	"github.com/fraxiinus/roflxd/rofl"
)

type replayFile interface {
	WriteToJson(path string) error
	WriteToRofl(path string) error
}

func main() {
	// Define command line arguments
	inputPtr := flag.String("input", "", "required - replay file to read")
	modePtr := flag.String("mode", "json", "optional - output mode (json, verify)")
	verbosePtr := flag.Bool("v", false, "optional - display verbose parse logs")
	OutPtr := flag.String("output", "", "optional - output file name")
	flag.Parse()

	// if input file argument isn't blank
	if len(*inputPtr) > 0 {
		var r replayFile
		var err error

		// get the file extension, create appropriate file
		ext := filepath.Ext(*inputPtr)
		switch ext {
		case ".rofl":
			fmt.Println("detected rofl file")
			r, err = rofl.New(*inputPtr, *verbosePtr)
		case ".glr":
			fmt.Println("detected glr file")
			r, err = glr.New(*inputPtr, *verbosePtr)
		default:
			msg := "unsupported file type: " + ext
			err = errors.New(msg)
		}

		// handle errors
		if err != nil {
			fmt.Println("failed to read file:", err)
		} else {
			fmt.Println("successfully read file")

			// optionally write output to file
			if len(*OutPtr) > 0 {
				err := WriteOutput(*modePtr, *OutPtr, r)
				if err != nil {
					fmt.Println("failed to write file:", err)
				} else {
					fmt.Println("successfully wrote file:", *OutPtr)
				}
			}
		}
	} else {
		flag.Usage()
	}
}

func WriteOutput(mode string, path string, r replayFile) error {
	switch mode {
	case "json":
		r.WriteToJson(path)
	case "rofl":
		r.WriteToRofl(path)
	default:
		return errors.New("unsupported output mode")
	}

	return nil
}
