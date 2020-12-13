package database

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/keyan/simpledb/rpc"
)

const (
	logFilename        = "logfile"
	checkpointFilename = "checkpoint"
	versionFilename    = "version"
	newVersionFilename = "newVersion"
	dataPath           = "data/"
)

// Keep a pool to use during write-ahead-logging encoding to save on
// bytes.Buffer allocations.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type journal interface {
	initialize(s memStorage) error
	checkpoint(s memStorage)
	addOp(msg *rpc.Msg)
}

// fileJournal allows access and manipulation of the on-disk data storage.
type fileJournal struct {
	version     int
	logFile     *os.File
	jsonEncoder *json.Encoder
}

// Ensure that fileJournal implements journal.
var _ journal = (*fileJournal)(nil)

// init adds the data directory in case it is not present.
func init() {
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		os.Mkdir(dataPath, os.ModeDir|os.ModePerm)
	}
}

// newJournal returns an empty fileJournal, users must call initialize() before use.
func newJournal() *fileJournal {
	return &fileJournal{}
}

// initialize prepares the journal for use by using on-disk files to load
// any prior state to the provided memStorage, along with replaying any
// commands in the logFile.
func (fj *fileJournal) initialize(memS memStorage) error {
	var vfn string
	if _, err := os.Stat(filepath.Join(dataPath, newVersionFilename)); err == nil {
		vfn = newVersionFilename
	} else {
		vfn = versionFilename
	}
	if dat, err := ioutil.ReadFile(filepath.Join(dataPath, vfn)); err == nil {
		fj.version, err = strconv.Atoi(string(dat))
	}
	fmt.Printf("Starting from version %v\n", fj.version)

	cpf := checkpointFilename + strconv.Itoa(fj.version)
	cpp := filepath.Join(dataPath, cpf)
	if _, err := os.Stat(cpp); err == nil {
		fmt.Printf("Reading checkpoint file: %v\n", cpp)
		dat, err := ioutil.ReadFile(cpp)
		if err != nil {
			panic("Checkpoint file exists but could not be read")
		}

		dec := gob.NewDecoder(bytes.NewBuffer(dat))
		if err = dec.Decode(&memS); err != nil {
			fmt.Printf("Could not decode checkpoint file, err: %v\n", err)
		}
	}

	lf := logFilename + strconv.Itoa(fj.version)
	lfp := filepath.Join(dataPath, lf)
	if _, err := os.Stat(lfp); err == nil {
		fmt.Println("Replaying commands from logFile...")
		cmdCnt := 0

		f, _ := os.Open(lfp)
		dec := json.NewDecoder(f)
		for {
			var msg rpc.Msg
			if err := dec.Decode(&msg); err == io.EOF {
				break
			} else if err != nil {
				panic("Could not decode message in logFile")
			}

			switch msg.Op {
			case rpc.Set:
				memS[msg.Key] = msg.Value
			case rpc.Delete:
				delete(memS, msg.Key)
			}
			cmdCnt++
		}
		fmt.Printf("Replayed %v commands from logFile\n", cmdCnt)
	}
	// If the file doesn't exist, or it does we finished replaying, overwrite
	// with a new logFile.
	fj.logFile, _ = os.Create(lfp)
	fj.jsonEncoder = json.NewEncoder(fj.logFile)

	return nil
}

// checkpoint stores a new version of the memStorage state on disk, overwriting
// the prior copy of the state and incrementing the journal version.
func (fj *fileJournal) checkpoint(memS memStorage) {
	newVersion := fj.version + 1
	fmt.Printf("Saving DB checkpoint, version: %v\n", newVersion)

	// Using gob here was perhaps a mistake, json encoder might be better.
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(memS)

	newLogPath := filepath.Join(dataPath, logFilename+strconv.Itoa(newVersion))
	newCheckPath := filepath.Join(dataPath, checkpointFilename+strconv.Itoa(newVersion))
	newVersionPath := filepath.Join(dataPath, newVersionFilename)

	oldLogPath := filepath.Join(dataPath, logFilename+strconv.Itoa(fj.version))
	oldCheckPath := filepath.Join(dataPath, checkpointFilename+strconv.Itoa(fj.version))
	oldVersionPath := filepath.Join(dataPath, versionFilename)

	// Create the new logFile, checkpoint, and newVersion files. The is the commit point.
	fj.logFile, _ = os.Create(newLogPath)
	fj.jsonEncoder = json.NewEncoder(fj.logFile)
	// Use os.Create() so we can fsync to disk, ioutil.WriteFile doesn't allow this.
	checkF, _ := os.Create(newCheckPath)
	defer checkF.Close()
	checkF.Write(buf.Bytes())
	checkF.Sync()

	versionF, _ := os.Create(newVersionPath)
	defer versionF.Close()
	versionF.Write([]byte(strconv.Itoa(newVersion)))
	versionF.Sync()

	// Cleanup the old files and rename newVersion to version.
	os.Remove(oldLogPath)
	os.Remove(oldCheckPath)
	os.Remove(oldVersionPath)
	os.Rename(newVersionPath, oldVersionPath)

	fj.version = newVersion
}

// addOp adds an operation to the logFile.
func (fj *fileJournal) addOp(msg *rpc.Msg) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	fj.jsonEncoder.Encode(msg)
}
