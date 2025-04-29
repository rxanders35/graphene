package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

var errObjectNotFound = errors.New("object not found in needle file")

const TCPport = "80811"

var worker *Worker

type Worker struct {
	mu         sync.Mutex
	needleFile *os.File
	idxFile    *os.File

	idx map[[16]byte]struct {
		offset int64
		size   int32
	}
}

func startWorker(port string) {
	w, err := newWorker(port)
	if err != nil {
		log.Fatal("Failed to start worker server (line 41)")
	}
	worker = w
	defer w.Close()
	log.Print("Starting server")
	go worker.initHTTP(port)
	go worker.initTCP(port)
}

func newWorker(port string) (*Worker, error) {
	w := &Worker{}
	w.idx = make(map[[16]byte]struct {
		offset int64
		size   int32
	})

	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		log.Printf("Created dir: %v", err)
		return nil, err
	}

	needleFileName := filepath.Join(dataDir, volume+port+dataExt)
	idxFileName := filepath.Join(dataDir, volume+port+idxExt)

	log.Print("Using files " + needleFileName + " and " + idxFileName)

	log.Print("Opening " + "data/" + needleFileName)

	needle, err := os.OpenFile(needleFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0664)
	if err != nil {
		log.Printf("Error specifically caused by: %v", err)
		return nil, err
	}
	w.needleFile = needle

	log.Print("Opening " + idxFileName)
	idx, err := os.OpenFile(needleFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0664)
	if err != nil {
		log.Printf("Error specifically caused by: %v", err)
		return nil, err
	}
	w.idxFile = idx

	log.Print("retrieving index file info: " + idxFileName)
	indexInfo, err := w.idxFile.Stat()
	if err != nil {
		log.Print("Failed to retrieve info")
		w.Close()
		return nil, err
	}

	if indexInfo.Size()%idxSizeTotal != 0 {
		log.Print("corrupt idx file")
		w.Close()
		return nil, err
	}

	numEntries := indexInfo.Size() / idxSizeTotal
	log.Print("Now attempting to read the needle file")
	for i := int64(0); i < numEntries; i++ {
		_, err := idx.Seek(i*idxSizeTotal, 0)
		if err != nil {
			w.Close()
			return nil, fmt.Errorf("failed to seek to entry %d: %w", i, err)
		}
		id := [16]byte{}
		n, err := idx.Read(id[:])

		offsetBytes := make([]byte, 8)
		n, err = idx.Read(offsetBytes)
		if err != nil || n != 8 {
			w.Close()
			log.Print("offset not found")
			return nil, err
		}

		offset := int64(binary.BigEndian.Uint64(offsetBytes))

		sizeBytes := make([]byte, 4)
		n, err = idx.Read(sizeBytes)
		if err != nil || n != 4 {
			w.Close()
			log.Print("failed to read size")
			return nil, err
		}

		size := int32(binary.BigEndian.Uint32(sizeBytes))

		w.idx[id] = struct {
			offset int64
			size   int32
		}{offset, size}

	}
	return w, nil
}

func (w *Worker) Store(id [16]byte, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	needleFileStat, err := w.needleFile.Stat()
	if err != nil {
		return err
	}

	offset := needleFileStat.Size()
	checksum := crc32.ChecksumIEEE(data)

	needleSize := needleMagicSize + needleIDSize + needleDataSize + len(data) + needleChecksum
	needle := make([]byte, needleSize)

	binary.BigEndian.PutUint16(needle[0:2], needleMagicVal)

	copy(needle[2:18], id[:])
	binary.BigEndian.PutUint32(needle[18:22], uint32(len(data)))

	copy(needle[22:22+len(data)], data)
	binary.BigEndian.PutUint32(needle[22+len(data):], checksum)

	n, err := w.needleFile.Write(needle)
	if err != nil {
		return fmt.Errorf("failed to write needle: %w", err)
	}
	if n != needleSize {
		return errors.New("incomplete write to needle file")
	}

	idxEntry := make([]byte, 28)
	copy(idxEntry[0:16], id[:])
	binary.BigEndian.PutUint64(idxEntry[16:24], uint64(offset))
	binary.BigEndian.PutUint32(idxEntry[24:28], uint32(len(data)))
	_, err = w.idxFile.Write(idxEntry)
	if err != nil {
		return err
	}

	w.idx[id] = struct {
		offset int64
		size   int32
	}{offset, int32(len(data))}

	return nil
}

func (w *Worker) Get(id [16]byte) ([]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	entry, exists := w.idx[id]
	if !exists {
		return nil, errObjectNotFound
	}
	offset := entry.offset
	size := entry.size

	_, err := w.needleFile.Seek(offset, 0)
	if err != nil {
		return nil, fmt.Errorf("Failed to seek: %w", err)
	}

	magicBytes := make([]byte, 2)
	n, err := w.needleFile.Read(magicBytes[:])
	if err != nil || n != 2 {
		return nil, fmt.Errorf("failed to read magic number: %w", err)
	}
	if binary.BigEndian.Uint16(magicBytes[:]) != needleMagicVal {
		return nil, fmt.Errorf("invalid needle magic number")
	}

	idBytes := make([]byte, 16)
	n, err = w.needleFile.Read(idBytes[:])
	if err != nil || n != 16 {
		return nil, fmt.Errorf("failed to read Needle ID: %w", err)
	}

	sizeBytes := make([]byte, 4)
	n, err = w.needleFile.Read(sizeBytes[:])
	if err != nil || n != 4 {
		return nil, fmt.Errorf("failed to read data size: %w", err)
	}
	s := binary.BigEndian.Uint32(sizeBytes[:])
	if int32(s) != size {
		return nil, errors.New("needle size mismatch")
	}

	data := make([]byte, entry.size)
	if n, err = w.needleFile.Read(data); err != nil || n != int(entry.size) {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	checksumBytes := make([]byte, 4)
	n, err = w.needleFile.Read(checksumBytes[:])
	if err != nil || n != 4 {
		return nil, fmt.Errorf("failed to read checksum: %w", err)
	}
	storedChecksum := binary.BigEndian.Uint32(checksumBytes[:])
	checksum := crc32.ChecksumIEEE(data)
	if checksum != storedChecksum {
		return nil, errors.New("checksum mismatch")
	}

	return data, nil
}

func (w *Worker) Close() error {
	var errs []error

	if w.needleFile != nil {
		if err := w.needleFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if w.idxFile != nil {
		if err := w.needleFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (w *Worker) initHTTP(port string) {
	log.Print("calling from within initHTTP")
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", w.handleUpload)
	mux.HandleFunc("/object/", w.handleGet)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal("Failed to start HTTP server: %v", err)
	}
}

func (w *Worker) handleUpload(rw http.ResponseWriter, req *http.Request) {
	log.Printf("Recieved %s request to %s", req.Method, req.URL.Path)
	if req.Method != "POST" {
		http.Error(rw, "Only POST allowed", http.StatusMethodNotAllowed)
	}

	req.Body = http.MaxBytesReader(rw, req.Body, 10<<20)
	defer req.Body.Close()
	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Failed to read body: %v", err)
		http.Error(rw, "Invalid req body", http.StatusBadRequest)
	}

	idBytes := uuid.New()
	idBytes.MarshalBinary()

	err = w.Store(idBytes, data)
	if err != nil {
		log.Printf("Failed to write needle: %v", err)
		http.Error(rw, "Storage Error", http.StatusInternalServerError)
	}

	resp := struct{ ID string }{ID: idBytes.String()}
	j, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v", err)
		http.Error(rw, "Response error", http.StatusInternalServerError)
	}

	rw.Header().Set("Content-Type", "applicaton/json")
	rw.WriteHeader(http.StatusCreated)

	_, err = rw.Write(j)
	if err != nil {
		log.Printf("JSON Write failed: %v", err)
	}

	log.Printf("Stored object: %s with size: %d", idBytes.String, len(data))

}

func (w *Worker) handleGet(rw http.ResponseWriter, req *http.Request) {
	log.Printf("Request method: %s to: %s", req.Method, req.URL.Path)
	if req.Method != "GET" {
		http.Error(rw, "Only GET allowed", http.StatusMethodNotAllowed)
	}

	id := strings.TrimPrefix(req.URL.Path, "/object/")
	if len(id) == 0 || id == "/" {
		http.Error(rw, "No ID present", http.StatusBadRequest)
	}

	uuid, err := uuid.Parse(id)
	if err != nil {
		http.Error(rw, "Invalid ID format", http.StatusBadRequest)
	}
	uuid.MarshalBinary()
	data, err := w.Get(uuid)
	if errors.Is(err, errObjectNotFound) {
		http.Error(rw, "Object not found", http.StatusInternalServerError)
	}
	if err != nil {
		log.Printf("Get failed:  %v", err)
		return
	}

	rw.Header().Set("Content-Length", strconv.Itoa(len(data)))

	_, err = rw.Write(data)
	if err != nil {
		log.Printf("Write failed: %v", err)
	}
	log.Printf("Retrieved object: %s successfully", uuid.String())
}

func (w *Worker) initTCP(port string) {
	w.listenAndAccept(port)
}

func (w *Worker) listenAndAccept(port string) {
	listener, err := net.Listen("tcp", ":"+TCPport)
	if err != nil {
		log.Fatalf("Failed to init listener on port %s due to: %v", TCPport, err)
	}
	defer listener.Close()
	log.Printf("TCP starting on port: %s", TCPport)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept due to: %v", err)
		}
		go w.handleTCPConn(conn)
	}
}

func (w *Worker) handleTCPConn(c net.Conn) {
	defer c.Close()
	log.Printf("New TCP connection from: %v", c.RemoteAddr())
r:
	for {
		msgType, payload, err := decodeMessage(c)
		if err != nil {
			log.Printf("Failed to decode msg: %v", err)
			break
		}
		switch msgType {
		case registerMsg:
			log.Printf("Recieved message: REGISTER from: %v", c.RemoteAddr())
			r := encodeSuccess("OK")
			if _, err := c.Write(r); err != nil {
				log.Printf("Failed to write response: %v", err)
				break r
			}

		case storeMsg:
			log.Printf("Received message: STORE payload size: %v", len(payload))
			r := encodeSuccess("OK")
			if len(payload) < 16 {
				log.Printf("Invalid payload with size: %v", len(payload))
				e := encodeSuccess("Invalid Payload")
				if _, err := c.Write(e); err != nil {
					log.Printf("Failed to write response: %v", err)
					break r
				}
			}

			uuid := [16]byte{}
			copy(uuid[:16], payload[:])
			data := payload[16:]
			err := w.Store(uuid, data)
			if err != nil {
				log.Printf("Failed to write data to volume server: %v", err)
				e := encodeSuccess("Failed to write data to volume server.")
				if _, err := c.Write(e); err != nil {
					log.Printf("Failed to write response: %v", err)
					break r
				}
			}
			if _, err := c.Write(r); err != nil {
				log.Printf("Failed to write response: %v", err)
				break r
			}
		default:
			log.Printf("Unknown msg type: %v", msgType)
		}
	}
}

func encodeMessage(msgType byte, payload []byte) []byte {
	msgLength := uint32(msgTypeSize + len(payload))
	buf := make([]byte, msgSize+msgLength)
	binary.BigEndian.PutUint32(buf[:msgSize], msgLength)
	buf[4] = msgType
	copy(buf[5:], payload)
	return buf
}

func decodeMessage(r io.Reader) (msgType byte, payload []byte, err error) {
	lenBuf := make([]byte, 4)
	if _, err = io.ReadFull(r, lenBuf); err != nil {
		return 0, nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf)

	buf := make([]byte, length)
	if _, err = io.ReadFull(r, buf); err != nil {
		return 0, nil, err
	}
	t := buf[0]
	p := buf[1:]
	return t, p, nil
}

func encodeRegister(addr string) []byte {
	payload := []byte(addr)
	p := encodeMessage(registerMsg, payload)
	return p
}

func encodeStore(uuid [16]byte, data []byte) []byte {
	payload := make([]byte, 16+len(data))
	copy(payload[:16], uuid[:])
	copy(payload[16:], data)
	p := encodeMessage(storeMsg, payload)
	return p
}

func encodeSuccess(resp string) []byte {
	payload := []byte(resp)
	p := encodeMessage(successMsg, payload)
	return p
}
