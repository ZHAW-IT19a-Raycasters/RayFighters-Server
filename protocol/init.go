package protocol

import "github.zhaw.ch/raycasters/server/protocol/client"

func DetectClient(opcode uint8) client.ClientOpcode {
	return client.Detect(opcode)
}
