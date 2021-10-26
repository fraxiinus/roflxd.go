package glr

import (
	"time"

	"github.com/1lann/lol-replay/recording"
)

type Glr struct {
	HasGameMetadata bool
	HasUserMetadata bool
	IsComplete      bool
	GameInfo        recording.GameInfo
	Metadata        Metadata
	FirstChunkInfo  ChunkInfo
	LastChunkInfo   ChunkInfo
	Chunks          []Chunk
	Keyframes       []Keyframe
}

type GameInfo struct {
	Platform      string
	GameVersion   string
	RecordTime    time.Time
	EncryptionKey string
}

type Metadata struct {
	GameKey struct {
		GameID     int64  `json:"gameId"`
		PlatformID string `json:"platformId"`
	} `json:"gameKey"`
	GameServerAddress         string `json:"gameServerAddress"`
	Port                      int    `json:"port"`
	EncryptionKey             string `json:"encryptionKey"`
	ChunkTimeInterval         int    `json:"chunkTimeInterval"`
	StartTime                 string `json:"startTime"`
	GameEnded                 bool   `json:"gameEnded"`
	LastChunkID               int    `json:"lastChunkId"`
	LastKeyFrameID            int    `json:"lastKeyFrameId"`
	EndStartupChunkID         int    `json:"endStartupChunkId"`
	DelayTime                 int    `json:"delayTime"`
	PendingAvailableChunkInfo []struct {
		ChunkID      int    `json:"chunkId"`
		Duration     int    `json:"duration"`
		ReceivedTime string `json:"receivedTime"`
	} `json:"pendingAvailableChunkInfo"`
	PendingAvailableKeyFrameInfo []struct {
		KeyFrameID   int    `json:"keyFrameId"`
		ReceivedTime string `json:"receivedTime"`
		NextChunkID  int    `json:"nextChunkId"`
	} `json:"pendingAvailableKeyFrameInfo"`
	KeyFrameTimeInterval      int    `json:"keyFrameTimeInterval"`
	DecodedEncryptionKey      string `json:"decodedEncryptionKey"`
	StartGameChunkID          int    `json:"startGameChunkId"`
	GameLength                int    `json:"gameLength"`
	ClientAddedLag            int    `json:"clientAddedLag"`
	ClientBackFetchingEnabled bool   `json:"clientBackFetchingEnabled"`
	ClientBackFetchingFreq    int    `json:"clientBackFetchingFreq"`
	InterestScore             int    `json:"interestScore"`
	FeaturedGame              bool   `json:"featuredGame"`
	CreateTime                string `json:"createTime"`
	EndGameChunkID            int    `json:"endGameChunkId"`
	EndGameKeyFrameID         int    `json:"endGameKeyFrameId"`
}

type ChunkInfo struct {
	ChunkID            int `json:"chunkId"`
	AvailableSince     int `json:"availableSince"`
	NextAvailableChunk int `json:"nextAvailableChunk"`
	KeyFrameID         int `json:"keyFrameId"`
	NextChunkID        int `json:"nextChunkId"`
	EndStartupChunkID  int `json:"endStartupChunkId"`
	StartGameChunkID   int `json:"startGameChunkId"`
	EndGameChunkID     int `json:"endGameChunkId"`
	Duration           int `json:"duration"`
}

type Chunk struct {
	Length int
	Data   []byte
}

type Keyframe struct {
	Length int
	Data   []byte
}
