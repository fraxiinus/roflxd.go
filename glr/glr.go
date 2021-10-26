package glr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/1lann/lol-replay/recording"
)

// This is the debug function used in 1lann/lol-replay
func New(path string, verbose bool) (*Glr, error) {
	// attempt to open file
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("failed to open file:", err)
		return nil, err
	}

	old, err := recording.NewRecording(file)
	if err != nil {
		fmt.Println("failed to read glr:", err)
		file.Close()
		return nil, err
	}

	var r = Glr{}
	if verbose {
		fmt.Print("reading game metadata...")
	}
	r.HasGameMetadata = old.HasGameMetadata()
	r.HasUserMetadata = old.HasUserMetadata()
	r.IsComplete = old.IsComplete()
	if verbose {
		fmt.Print("OK\n")
	}

	if verbose {
		fmt.Print("reading game info...")
	}
	r.GameInfo = old.RetrieveGameInfo()
	if verbose {
		fmt.Print("OK\n")
	}

	buf := new(bytes.Buffer)

	if verbose {
		fmt.Print("reading game metadata...")
	}
	old.RetrieveGameMetadataTo(buf)
	json.Unmarshal(buf.Bytes(), &r.Metadata)
	if verbose {
		fmt.Print("OK\n")
	}

	if verbose {
		fmt.Print("reading chunk info...")
	}
	buf.Reset()
	old.RetrieveFirstChunkInfo().WriteTo(buf)
	json.Unmarshal(buf.Bytes(), &r.FirstChunkInfo)

	buf.Reset()
	old.RetrieveLastChunkInfo().WriteTo(buf)
	json.Unmarshal(buf.Bytes(), &r.LastChunkInfo)
	if verbose {
		fmt.Print("OK\n")
	}

	if verbose {
		fmt.Print("reading chunks...")
	}
	for i := 1; old.HasChunk(i); i++ {
		buf.Reset()
		old.RetrieveChunkTo(i, buf)

		var c = Chunk{
			Length: buf.Len(),
			Data:   buf.Bytes(),
		}
		r.Chunks = append(r.Chunks, c)
	}
	if verbose {
		fmt.Printf("read %v chunks\n", len(r.Chunks))
	}

	if verbose {
		fmt.Print("reading keyframes...")
	}
	for i := 1; old.HasKeyFrame(i); i++ {
		buf.Reset()
		old.RetrieveKeyFrameTo(i, buf)

		var k = Keyframe{
			Length: buf.Len(),
			Data:   buf.Bytes(),
		}
		r.Keyframes = append(r.Keyframes, k)
	}
	if verbose {
		fmt.Printf("read %v keyframes\n", len(r.Keyframes))
	}

	return &r, nil
}

func (r Glr) WriteToJson(path string) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	} else {
		outFile, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return err
		}
		outFile.Write(b)
	}
	return nil
}

func (r Glr) WriteToRofl(path string) error {
	return nil
}
