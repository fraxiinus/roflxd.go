package rofl

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/fraxiinus/roflxd/helper"
)

var magic = []byte{0x52, 0x49, 0x4F, 0x54, 00, 00}

func New(path string, verbose bool) (*Rofl, error) {
	// attempt to open file
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("failed to open file:", err)
		return nil, err
	}

	// create new Rofl struct
	newRofl := Rofl{}

	// magic number check
	err = newRofl.CheckMagicNumber(file)
	if err != nil {
		fmt.Println("magic number check: FAIL,", err)
		return nil, err
	} else {
		if verbose {
			fmt.Println("magic number check: OK")
		}
	}

	// read signature
	_, err = file.Read(newRofl.Signature[:])
	if err != nil {
		fmt.Println("failed to read signature:", err)
		return nil, err
	}
	if verbose {
		fmt.Printf("file signature: %x\n", newRofl.Signature)
	}

	// read lengths
	err = newRofl.ReadLengthFields(file, verbose)
	if err != nil {
		fmt.Println("failed to read length fields:", err)
		return nil, err
	}

	// read metadata
	buffer := make([]byte, newRofl.Lengths.Metadata)
	_, err = file.Read(buffer)
	if err != nil {
		fmt.Println("failed to read metadata:", err)
		return nil, err
	}

	newRofl.Metadata = MetadataJson{}
	err = json.Unmarshal(buffer, &newRofl.Metadata)
	if err != nil {
		fmt.Println("failed to unmarshal metadata json:", err)
	}
	if verbose {
		fmt.Printf("metadata: %v\n", newRofl.Metadata)
	}

	err = json.Unmarshal([]byte(newRofl.Metadata.StatsJSON), &newRofl.Metadata.Stats)
	if err != nil {
		fmt.Println("failed to unmarshal stats json:", err)
	}
	if verbose {
		fmt.Printf("metadata stats json: %v\n", newRofl.Metadata.Stats)
	}

	// read payload headers
	err = newRofl.ReadPayloadHeaders(file, verbose)
	if err != nil {
		fmt.Println("failed to read payload headers:", err)
		return nil, err
	}

	// read chunks headers
	err = newRofl.ReadChunksHeaders(file, verbose)
	if err != nil {
		fmt.Println("failed to read chunk headers:", err)
		return nil, err
	}

	// read keyframe headers
	err = newRofl.ReadKeyframeHeaders(file, verbose)
	if err != nil {
		fmt.Println("failed to read keyframe headers:", err)
		return nil, err
	}

	// get current offset
	newRofl.headerEnd, _ = file.Seek(0, 1)
	if verbose {
		fmt.Println("header end offset:", newRofl.headerEnd)
	}

	// read chunk data
	for i := 0; i < len(newRofl.Chunks); i++ {
		_, err = newRofl.Chunks[i].ReadChunk(file, newRofl.headerEnd)
		if err != nil {
			fmt.Printf("failed to read chunk data %d:%v\n", i+1, err)
			return nil, err
		}
		if verbose {
			fmt.Printf("chunk %d: %v ...\n", i+1, newRofl.Chunks[i].Data[:25])
		}
	}

	// read keyframe data
	for i := 0; i < len(newRofl.Keyframes); i++ {
		_, err = newRofl.Keyframes[i].ReadKeyframe(file, newRofl.headerEnd)
		if err != nil {
			fmt.Printf("failed to read keyframe data %d:%v\n", i+1, err)
			return nil, err
		}
		if verbose {
			fmt.Printf("keyframe %d: %v ...\n", i+1, newRofl.Keyframes[i].Data[:25])
		}
	}

	return &newRofl, nil
}

// !! NOT FUNCTIONAL !!
// func ConvertGlr(path string) (*Rofl, error) {
// 	file, err := os.OpenFile(path, os.O_RDWR, 0666)
// 	if err != nil {
// 		return nil, err
// 	}

// 	glr, err := recording.NewRecording(file)
// 	if err != nil {
// 		fmt.Println("failed to read glr:", err)
// 		file.Close()
// 		return nil, err
// 	}

// 	nRofl := Rofl{}

// 	gameInfo := glr.RetrieveGameInfo()
// 	nRofl.PayloadHeader.EncryptionKey = gameInfo.EncryptionKey
// 	nRofl.PayloadHeader.EncryptionKeyLength = uint16(len(gameInfo.EncryptionKey))
// 	// glr are missing payload headers besides first/last
// 	// all chunks/keyframes are still accounted for, can I construct them from that data?
// 	// are all the payload headers required to play a rofl?
// 	return nil, nil
// }

func (r Rofl) WriteToJson(path string) error {
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

// Creates a new rofl file based on what was loaded
// used to debug mapping
func (r Rofl) WriteToRofl(path string) error {
	// calculate the length fields required
	r.CalculateLengthFields()

	outFile, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	outFile.Write(r.Magic[:])

	outFile.Write(r.Signature[:])

	binary.Write(outFile, binary.LittleEndian, r.Lengths.Header)

	binary.Write(outFile, binary.LittleEndian, r.Lengths.File)

	binary.Write(outFile, binary.LittleEndian, r.Lengths.MetadataOffset)

	binary.Write(outFile, binary.LittleEndian, r.Lengths.Metadata)

	binary.Write(outFile, binary.LittleEndian, r.Lengths.PayloadHeaderOffset)

	binary.Write(outFile, binary.LittleEndian, r.Lengths.PayloadHeader)

	binary.Write(outFile, binary.LittleEndian, r.Lengths.PayloadOffset)

	exp, _ := json.Marshal(r.Metadata.Stats)
	r.Metadata.StatsJSON = string(exp)

	exp, _ = json.Marshal(r.Metadata)
	binary.Write(outFile, binary.LittleEndian, exp)

	binary.Write(outFile, binary.LittleEndian, r.PayloadHeader.GameId)

	binary.Write(outFile, binary.LittleEndian, r.PayloadHeader.GameLength)

	binary.Write(outFile, binary.LittleEndian, r.PayloadHeader.KeyframeCount)

	binary.Write(outFile, binary.LittleEndian, r.PayloadHeader.ChunkCount)

	binary.Write(outFile, binary.LittleEndian, r.PayloadHeader.EndStartupChunkId)

	binary.Write(outFile, binary.LittleEndian, r.PayloadHeader.StartGameChunkId)

	binary.Write(outFile, binary.LittleEndian, r.PayloadHeader.KeyframeInterval)

	binary.Write(outFile, binary.LittleEndian, r.PayloadHeader.EncryptionKeyLength)

	binary.Write(outFile, binary.LittleEndian, []byte(r.PayloadHeader.EncryptionKey))

	for i := 0; i < int(r.PayloadHeader.ChunkCount); i++ {
		binary.Write(outFile, binary.LittleEndian, r.Chunks[i].Id)

		binary.Write(outFile, binary.LittleEndian, r.Chunks[i].ChunkType)

		binary.Write(outFile, binary.LittleEndian, r.Chunks[i].Length)

		binary.Write(outFile, binary.LittleEndian, r.Chunks[i].NextId)

		binary.Write(outFile, binary.LittleEndian, r.Chunks[i].Offset)
	}

	for i := 0; i < int(r.PayloadHeader.KeyframeCount); i++ {
		binary.Write(outFile, binary.LittleEndian, r.Keyframes[i].Id)

		binary.Write(outFile, binary.LittleEndian, r.Keyframes[i].KeyframeType)

		binary.Write(outFile, binary.LittleEndian, r.Keyframes[i].Length)

		binary.Write(outFile, binary.LittleEndian, r.Keyframes[i].NextId)

		binary.Write(outFile, binary.LittleEndian, r.Keyframes[i].Offset)
	}

	for i := 0; i < int(r.PayloadHeader.ChunkCount); i++ {
		binary.Write(outFile, binary.LittleEndian, r.Chunks[i].Data)
	}

	for i := 0; i < int(r.PayloadHeader.KeyframeCount); i++ {
		binary.Write(outFile, binary.LittleEndian, r.Keyframes[i].Data)
	}

	return nil
}

// Get lengths based on data in memory
func (r *Rofl) CalculateLengthFields() {
	// Header is always 288 bytes
	r.Lengths.Header = 288
	r.Lengths.MetadataOffset = 288

	// calculate metadata size
	j, _ := json.Marshal(r.Metadata)
	r.Lengths.Metadata = uint32(len(j))

	// calculate payload header size
	r.Lengths.PayloadHeaderOffset = r.Lengths.MetadataOffset + r.Lengths.Metadata
	r.Lengths.PayloadHeader = uint32(34 + r.PayloadHeader.EncryptionKeyLength)

	// calculate payload
	r.Lengths.PayloadOffset = r.Lengths.PayloadHeaderOffset + r.Lengths.PayloadHeader

	// calculate total
	chunkHeaderLength := r.PayloadHeader.ChunkCount * 17
	keyframeHeaderLength := r.PayloadHeader.KeyframeCount * 17
	var chunkLength int
	var keyframeLength int

	for i := 0; i < int(r.PayloadHeader.ChunkCount); i++ {
		chunkLength += len(r.Chunks[i].Data)
	}
	for i := 0; i < int(r.PayloadHeader.KeyframeCount); i++ {
		chunkLength += len(r.Keyframes[i].Data)
	}

	r.Lengths.File = r.Lengths.PayloadOffset + chunkHeaderLength + keyframeHeaderLength + uint32(chunkLength) + uint32(keyframeLength)
}

// Check a ROFL file has the correct magic number
func (r *Rofl) CheckMagicNumber(file *os.File) error {
	_, err := file.Read(r.Magic[:])
	if err != nil {
		return err
	}

	if !bytes.Equal(magic, r.Magic[:]) {
		return errors.New("magic number does not match")
	}

	return nil
}

// Read the length fields of a file, critical to file reading
func (r *Rofl) ReadLengthFields(file *os.File, verbose bool) error {
	var err error

	r.Lengths.Header, err = helper.ReadUint16(file)
	if err != nil {
		fmt.Println("failed to read length.header:", err)
		return err
	}
	if verbose {
		fmt.Printf("length of header: %d\n", r.Lengths.Header)
	}

	r.Lengths.File, err = helper.ReadUint32(file)
	if err != nil {
		fmt.Println("failed to read length.file:", err)
		return err
	}
	if verbose {
		fmt.Printf("length of file: %d\n", r.Lengths.File)
	}

	r.Lengths.MetadataOffset, err = helper.ReadUint32(file)
	if err != nil {
		fmt.Println("failed to read length.metadataOffset:", err)
		return err
	}
	if verbose {
		fmt.Printf("metadata offset: %d\n", r.Lengths.MetadataOffset)
	}

	r.Lengths.Metadata, err = helper.ReadUint32(file)
	if err != nil {
		fmt.Println("failed to read length.metadata:", err)
		return err
	}
	if verbose {
		fmt.Printf("length of metadata: %d\n", r.Lengths.Metadata)
	}

	r.Lengths.PayloadHeaderOffset, err = helper.ReadUint32(file)
	if err != nil {
		fmt.Println("failed to read length.payloadHeaderOffset:", err)
		return err
	}
	if verbose {
		fmt.Printf("payload header offset: %d\n", r.Lengths.PayloadHeaderOffset)
	}

	r.Lengths.PayloadHeader, err = helper.ReadUint32(file)
	if err != nil {
		fmt.Println("failed to read length.payloadHeader:", err)
		return err
	}
	if verbose {
		fmt.Printf("length of payloadHeader: %d\n", r.Lengths.PayloadHeader)
	}

	r.Lengths.PayloadOffset, err = helper.ReadUint32(file)
	if err != nil {
		fmt.Println("failed to read length.payloadOffset:", err)
		return err
	}
	if verbose {
		fmt.Printf("payload offset: %d\n", r.Lengths.PayloadOffset)
	}

	return nil
}

// Read the payload headers, contain data relevent to the game
func (r *Rofl) ReadPayloadHeaders(file *os.File, verbose bool) error {
	var err error

	r.PayloadHeader.GameId, err = helper.ReadUint64(file)
	if err != nil {
		fmt.Println("failed to read payloadHeader.gameId:", err)
		return err
	}
	if verbose {
		fmt.Printf("gameId: %d\n", r.PayloadHeader.GameId)
	}

	r.PayloadHeader.GameLength, err = helper.ReadUint32(file)
	if err != nil {
		fmt.Println("failed to read payloadHeader.gameLength:", err)
		return err
	}
	if verbose {
		fmt.Printf("gameLength: %d\n", r.PayloadHeader.GameLength)
	}

	r.PayloadHeader.KeyframeCount, err = helper.ReadUint32(file)
	if err != nil {
		fmt.Println("failed to read payloadHeader.keyframeCount:", err)
		return err
	}
	if verbose {
		fmt.Printf("keyframeCount: %d\n", r.PayloadHeader.KeyframeCount)
	}

	r.PayloadHeader.ChunkCount, err = helper.ReadUint32(file)
	if err != nil {
		fmt.Println("failed to read payloadHeader.chunkCount:", err)
		return err
	}
	if verbose {
		fmt.Printf("chunkCount: %d\n", r.PayloadHeader.ChunkCount)
	}

	r.PayloadHeader.EndStartupChunkId, err = helper.ReadUint32(file)
	if err != nil {
		fmt.Println("failed to read payloadHeader.endStartupChunkId:", err)
		return err
	}
	if verbose {
		fmt.Printf("endStartupChunkId: %d\n", r.PayloadHeader.EndStartupChunkId)
	}

	r.PayloadHeader.StartGameChunkId, err = helper.ReadUint32(file)
	if err != nil {
		fmt.Println("failed to read payloadHeader.startGameChunkId:", err)
		return err
	}
	if verbose {
		fmt.Printf("startGameChunkId: %d\n", r.PayloadHeader.StartGameChunkId)
	}

	r.PayloadHeader.KeyframeInterval, err = helper.ReadUint32(file)
	if err != nil {
		fmt.Println("failed to read payloadHeader.keyframeInterval:", err)
		return err
	}
	if verbose {
		fmt.Printf("keyframeInterval: %d\n", r.PayloadHeader.KeyframeInterval)
	}

	r.PayloadHeader.EncryptionKeyLength, err = helper.ReadUint16(file)
	if err != nil {
		fmt.Println("failed to read payloadHeader.encryptionKeyLength:", err)
		return err
	}
	if verbose {
		fmt.Printf("encryptionKeyLength: %d\n", r.PayloadHeader.EncryptionKeyLength)
	}

	// resize
	buffer := make([]byte, r.PayloadHeader.EncryptionKeyLength)
	_, err = file.Read(buffer)
	if err != nil {
		fmt.Println("failed to read newRofl.payloadHeader.encryptionKey:", err)
		return err
	}
	r.PayloadHeader.EncryptionKey = string(buffer)
	if verbose {
		fmt.Printf("encryptionKey: %s\n", r.PayloadHeader.EncryptionKey)
	}

	return nil
}

// Read chunk headers, contains basic information on chunks
func (r *Rofl) ReadChunksHeaders(file *os.File, verbose bool) error {
	var err error
	if verbose {
		fmt.Println("chunk headers:")
	}

	r.Chunks = make([]Chunk, r.PayloadHeader.ChunkCount)
	for i := 0; i < int(r.PayloadHeader.ChunkCount); i++ {
		r.Chunks[i].Id, err = helper.ReadUint32(file)
		if err != nil {
			fmt.Printf("failed to read chunk header %d id:%v\n", i+1, err)
			return err
		}
		if verbose {
			fmt.Printf("\tnumber %d \tid: %d\t", i+1, r.Chunks[i].Id)
		}

		r.Chunks[i].ChunkType, err = helper.ReadByte(file)
		if err != nil {
			fmt.Printf("failed to read chunk header %d type:%v\n", i+1, err)
			return err
		}
		if verbose {
			fmt.Printf("type: %d\t", r.Chunks[i].ChunkType)
		}

		r.Chunks[i].Length, err = helper.ReadUint32(file)
		if err != nil {
			fmt.Printf("failed to read chunk header %d length:%v\n", i+1, err)
			return err
		}
		if verbose {
			fmt.Printf("length: %d\t", r.Chunks[i].Length)
		}

		r.Chunks[i].NextId, err = helper.ReadUint32(file)
		if err != nil {
			fmt.Printf("failed to read chunk header %d nextId:%v\n", i+1, err)
			return err
		}
		if verbose {
			fmt.Printf("nextId: %d\t", r.Chunks[i].NextId)
		}

		r.Chunks[i].Offset, err = helper.ReadUint32(file)
		if err != nil {
			fmt.Printf("failed to read chunk header %d offset:%v\n", i+1, err)
			return err
		}
		if verbose {
			fmt.Printf("offset: %d\n", r.Chunks[i].Offset+44720)
		}
	}

	return nil
}

// Read keyframe headers, contains basic information on keyframes
func (r *Rofl) ReadKeyframeHeaders(file *os.File, verbose bool) error {
	var err error

	if verbose {
		fmt.Println("keyframe headers:")
	}

	r.Keyframes = make([]Keyframe, r.PayloadHeader.KeyframeCount)
	for i := 0; i < int(r.PayloadHeader.KeyframeCount); i++ {
		r.Keyframes[i].Id, err = helper.ReadUint32(file)
		if err != nil {
			fmt.Printf("failed to read keyframe header %d id:%v\n", i+1, err)
			return err
		}
		if verbose {
			fmt.Printf("\tnumber %d \tid: %d\t", i+1, r.Keyframes[i].Id)
		}

		r.Keyframes[i].KeyframeType, err = helper.ReadByte(file)
		if err != nil {
			fmt.Printf("failed to read keyframe header %d type:%v\n", i+1, err)
			return err
		}
		if verbose {
			fmt.Printf("type: %d\t", r.Keyframes[i].KeyframeType)
		}

		r.Keyframes[i].Length, err = helper.ReadUint32(file)
		if err != nil {
			fmt.Printf("failed to read keyframe header %d length:%v\n", i+1, err)
			return err
		}
		if verbose {
			fmt.Printf("length: %d\t", r.Keyframes[i].Length)
		}

		r.Keyframes[i].NextId, err = helper.ReadUint32(file)
		if err != nil {
			fmt.Printf("failed to read keyframe header %d nextId:%v\n", i+1, err)
			return err
		}
		if verbose {
			fmt.Printf("nextId: %d\t", r.Keyframes[i].NextId)
		}

		r.Keyframes[i].Offset, err = helper.ReadUint32(file)
		if err != nil {
			fmt.Printf("failed to read keyframe header %d offset:%v\n", i+1, err)
			return err
		}
		if verbose {
			fmt.Printf("offset: %d\n", r.Keyframes[i].Offset+44720)
		}
	}

	return nil
}

// Read chunks, this the "replay" data. Pretty much impossible to make sense of
func (c *Chunk) ReadChunk(file *os.File, headerOffset int64) (int, error) {
	c.Data = make([]byte, c.Length)

	totalRead, err := file.ReadAt(c.Data, int64(c.Offset)+headerOffset)
	if err != nil {
		return 0, err
	}

	return totalRead, nil
}

// Read keyframes
func (k *Keyframe) ReadKeyframe(file *os.File, headerOffset int64) (int, error) {
	k.Data = make([]byte, k.Length)

	totalRead, err := file.ReadAt(k.Data, int64(k.Offset)+headerOffset)
	if err != nil {
		return 0, err
	}

	return totalRead, nil
}
