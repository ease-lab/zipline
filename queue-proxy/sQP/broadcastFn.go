package sQP

import (
	"io"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/ease-lab/vhive-xdt/proto/crossXDT"
	"github.com/ease-lab/vhive-xdt/proto/upXDT"
)

// BroadcastUpload is called by BroadcastUpload to push data to sQP
func (s upXDTServer) BroadcastUpload(srv upXDT.StreamData_BroadcastUploadServer) error {
	chunkCount := 0
	byteCount := 0
	var onlyOnce sync.Once
	var totalChunks int64
	var key string
	var payloadBytes []byte
	for {
		chunk, err := srv.Recv()
		if err == io.EOF {
			log.Infof("sQP: Received %d chunks at DstFn with first/last bytes as:", chunkCount)
			log.Info(payloadBytes[0:9], payloadBytes[byteCount-9:byteCount])
			bufferPool.StoreSlice(key, totalChunks, payloadBytes)
			return srv.SendAndClose(&upXDT.Empty{})
		}
		if err != nil {
			log.Errorf("sQP: receive error: %v", err)
			return err
		}
		log.Debugf("sQP: Received chunk no. %d", chunkCount)
		onlyOnce.Do(func() {
			key = chunk.Key
			totalChunks = chunk.TotalChunks
			log.Infof("sQP: creating a new buffer")
			payloadBytes = make([]byte, int(totalChunks)*len(chunk.Chunk))
			log.Infof("sQP: chunkTotal = %d", totalChunks)
		})
		log.Debugf("sQP: appending chunk number %d", chunkCount)
		copy(payloadBytes[byteCount:], chunk.Chunk)
		byteCount += len(chunk.Chunk)
		chunkCount += 1
	}
}

// ServeBroadcastData is the gRPC server to serve the available data to the dQP
func (s crossXDTServer) ServeBroadcastData(in *crossXDT.BroadcastRequest, srv crossXDT.StreamData_ServeBroadcastDataServer) error {

	log.Infof("sQP: dQP is fetching key: %s", in.Key)

	var slice []byte
	var chunkTotal int64

	// Check whether the first packet has been received at sQP or not
	for {
		if slice, chunkTotal = bufferPool.GetSlice(in.Key); slice != nil {
			log.Infof("sQP: found chunkTotal %d for key %s", chunkTotal, in.Key)
			break
		}
	}
	payloadSize := len(slice)
	log.Infof("Broadcasting %d bytes to sQP", payloadSize)

	chunkSizeInBytes := int(in.ChunkSizeInBytes)

	for currentByte := 0; currentByte < payloadSize; currentByte += chunkSizeInBytes {

		if currentByte+chunkSizeInBytes > payloadSize {
			resp := crossXDT.Response{Chunk: slice[currentByte:payloadSize], TotalChunks: chunkTotal}
			if err := srv.Send(&resp); err != nil {
				log.Errorf("send error %v", err)
				return err
			}
			log.Debugf("finishing request number : %d", currentByte)
		} else {
			resp := crossXDT.Response{Chunk: slice[currentByte : currentByte+chunkSizeInBytes], TotalChunks: chunkTotal}
			if err := srv.Send(&resp); err != nil {
				log.Errorf("send error %v", err)
				return err

			}
			log.Debugf("finishing request number : %d", currentByte)
		}

	}
	return nil
}
