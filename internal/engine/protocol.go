package engine

import (
	"context"
	"encoding/binary"
	"io"
	"log"
	"sync/atomic"
)

func TunToStream(ctx context.Context, tun io.Reader, stream io.Writer, counter *uint64) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			buf := bufPool.Get().([]byte)
			n, err := tun.Read(buf)
			if err != nil {
				bufPool.Put(buf)
				log.Printf("engine: tun read ended err=%v", err)
				return
			}
			err = binary.Write(stream, binary.BigEndian, uint16(n))
			if err != nil {
				bufPool.Put(buf)
				log.Printf("engine: stream write header ended err=%v", err)
				return
			}
			_, err = stream.Write(buf[:n])
			bufPool.Put(buf)
			if err != nil {
				log.Printf("engine: stream write payload ended err=%v", err)
				return
			}
			atomic.AddUint64(counter, uint64(n))
		}
	}
}

func StreamToTun(ctx context.Context, tun io.Writer, stream io.Reader, counter *uint64) {
	header := make([]byte, 2)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := io.ReadFull(stream, header)
			if err != nil {
				log.Printf("engine: stream read header ended err=%v", err)
				return
			}
			packetLen := binary.BigEndian.Uint16(header)
			if packetLen > 1600 {
				log.Printf("engine: invalid packet length=%d; closing direction", packetLen)
				return
			}
			buf := bufPool.Get().([]byte)
			_, err = io.ReadFull(stream, buf[:packetLen])
			if err != nil {
				bufPool.Put(buf)
				log.Printf("engine: stream read payload ended err=%v", err)
				return
			}
			n, err := tun.Write(buf[:packetLen])
			bufPool.Put(buf)
			if err != nil {
				log.Printf("engine: tun write ended err=%v", err)
				return
			}
			atomic.AddUint64(counter, uint64(n))
		}
	}
}
