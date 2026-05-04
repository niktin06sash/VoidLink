package engine

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"
)

func TestTunToStream_ValidData(t *testing.T) {
	tunR, tunW := io.Pipe()
	streamR, streamW := io.Pipe()
	var counter uint64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		TunToStream(ctx, tunR, streamW, &counter)
	}()

	testData := []byte("hello world")
	go func() {
		tunW.Write(testData)
		tunW.Close()
	}()

	header := make([]byte, 2)
	if _, err := io.ReadFull(streamR, header); err != nil {
		t.Fatal(err)
	}
	length := binary.BigEndian.Uint16(header)

	payload := make([]byte, length)
	if _, err := io.ReadFull(streamR, payload); err != nil {
		t.Fatal(err)
	}

	cancel()
	wg.Wait()

	if atomic.LoadUint64(&counter) != uint64(len(testData)) {
		t.Errorf("Expected %d bytes in counter, got %d", len(testData), counter)
	}
}

func TestStreamToTun_ValidData(t *testing.T) {
	streamR, streamW := io.Pipe()
	tunR, tunW := io.Pipe()
	var counter uint64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go StreamToTun(ctx, tunW, streamR, &counter)

	testData := []byte("p2p packet")
	go func() {
		_ = binary.Write(streamW, binary.BigEndian, uint16(len(testData)))
		_, _ = streamW.Write(testData)
		cancel()
	}()

	result := make([]byte, len(testData))
	if _, err := io.ReadFull(tunR, result); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(result, testData) {
		t.Errorf("Data mismatch: got %s, want %s", result, testData)
	}
}

func TestStreamToTun_InvalidLength(t *testing.T) {
	streamR, streamW := io.Pipe()
	var counter uint64
	ctx := t.Context()
	go func() {
		_ = binary.Write(streamW, binary.BigEndian, uint16(2000))
		_, _ = streamW.Write(make([]byte, 2000))
	}()

	StreamToTun(ctx, io.Discard, streamR, &counter)

	if atomic.LoadUint64(&counter) != 0 {
		t.Error("Should not process packets larger than 1600")
	}
}
func TestStreamToTun_PartialData(t *testing.T) {
	streamR, streamW := io.Pipe()
	tun := io.Discard
	var counter uint64
	ctx := t.Context()

	go func() {
		binary.Write(streamW, binary.BigEndian, uint16(100))
		streamW.Write(make([]byte, 10))
		streamW.Close()
	}()
	StreamToTun(ctx, tun, streamR, &counter)
}
func TestEngine_StressRace(t *testing.T) {
	ctx := t.Context()
	const workers = 5
	const packetsPerWorker = 50
	var wg sync.WaitGroup
	var totalBytes uint64
	for i := range workers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			tunR, tunW := io.Pipe()
			streamR, streamW := io.Pipe()
			engineDone := make(chan struct{})
			go func() {
				TunToStream(ctx, tunR, streamW, &totalBytes)
				close(engineDone)
			}()
			go func() {
				header := make([]byte, 2)
				for {
					if _, err := io.ReadFull(streamR, header); err != nil {
						return
					}
					pLen := binary.BigEndian.Uint16(header)
					payload := make([]byte, pLen)
					if _, err := io.ReadFull(streamR, payload); err != nil {
						return
					}
				}
			}()
			for j := range packetsPerWorker {
				data := fmt.Appendf(nil, "w%d-p%d", workerID, j)
				_, err := tunW.Write(data)
				if err != nil {
					break
				}
			}
			tunW.Close()
			<-engineDone
			streamW.Close()
		}(i)
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		t.Logf("Success! Total bytes processed: %d", atomic.LoadUint64(&totalBytes))
	case <-ctx.Done():
		t.Fatal("Test timed out! Deadlock suspected.")
	}
}
