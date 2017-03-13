package api

import (
	"bytes"
	"io"
)

// NewChunkReader creates an io.Reader from a ChunkReceiver.
// You can optionally pass in your own buffer to use or one will be created.
// The reader works to stitch back together Chunks (a protobuf encapsulated stream of bytes)
// as a single stream.
func NewChunkReader(recv ChunkReceiver, buf []byte) io.Reader {
	if buf != nil {
		buf = buf[0:]
	}
	return &chunkReader{r: recv, buf: bytes.NewBuffer(buf)}
}

// ChunkReceiver is when creating a chunk reader to be able to accept different
// streaming endpoints from the API server.
type ChunkReceiver interface {
	Recv() (*Chunk, error)
}

type chunkReader struct {
	buf *bytes.Buffer
	r   ChunkReceiver
}

func (r *chunkReader) Read(b []byte) (nr int, err error) {
	nr, err = r.buf.Read(b)

	for {
		if nr < len(b) {
			chunk, err := r.r.Recv()
			if err != nil {
				return nr, err
			}

			if _, err := r.buf.Write(chunk.Chunk); err != nil {
				return nr, err
			}

			n, err := r.buf.Read(b[nr:])
			nr += n
			if err != nil {
				return nr, err
			}
		}
	}
}