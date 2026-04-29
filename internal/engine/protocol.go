package engine

import (
	"context"
	"encoding/binary"
	"io"
)

func TunToStream(ctx context.Context, tun io.Reader, stream io.Writer) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			buf := bufPool.Get().([]byte)
			n, err := tun.Read(buf)
			if err != nil {
				bufPool.Put(buf)
				return
			}
			err = binary.Write(stream, binary.BigEndian, uint16(n))
			if err != nil {
				bufPool.Put(buf)
				return
			}
			_, err = stream.Write(buf[:n])
			bufPool.Put(buf)
			if err != nil {
				return
			}
		}
	}
}

func StreamToTun(ctx context.Context, tun io.Writer, stream io.Reader) {
	header := make([]byte, 2)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := io.ReadFull(stream, header)
			if err != nil {
				return
			}
			packetLen := binary.BigEndian.Uint16(header)
			if packetLen > 1600 {
				return
			}
			buf := bufPool.Get().([]byte)
			_, err = io.ReadFull(stream, buf[:packetLen])
			if err != nil {
				bufPool.Put(buf)
				return
			}
			_, err = tun.Write(buf[:packetLen])
			bufPool.Put(buf)
			if err != nil {
				return
			}
		}
	}
}
