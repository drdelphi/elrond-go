package heartbeat

import (
	"time"
)

// Heartbeat represents the heartbeat message that is sent between peers
type Heartbeat struct {
	Payload         []byte
	Pubkey          []byte
	Signature       []byte
	ShardID         uint32
	VersionNumber   string
	NodeDisplayName string
}

// PubKeyHeartbeat returns the heartbeat status for a public key
type PubKeyHeartbeat struct {
	HexPublicKey    string    `json:"hexPublicKey"`
	TimeStamp       time.Time `json:"timeStamp"`
	MaxInactiveTime Duration  `json:"maxInactiveTime"`
	IsActive        bool      `json:"isActive"`
	//TODO should have 2 fields for this: receivedShardID and computedShardID
	ShardID         uint32   `json:"shardID"`
	TotalUpTime     Duration `json:"totalUpTime"`
	TotalDownTime   Duration `json:"totalDownTime"`
	VersionNumber   string   `json:"versionNumber"`
	IsValidator     bool     `json:"isValidator"`
	NodeDisplayName string   `json:"nodeDisplayName"`
}
