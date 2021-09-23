package metrics

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

const RECORDS_DB = "records"
const RECORDS_TABLE = "records"
const JS_TIME_LAYOUT = "Mon Jan 02 15:04:05 MST 2006"

type record struct {
	Startdate       string `json:"startDate"`
	Enddate         string `json:"endDate"`
	IntegrationName string `json:"taskName"`
	Stage           string `json:"status"`
}

func toBytes(r record) []byte {
	bts, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return bts
}

func CreateRecordsDB(dbPath string) error {
	db, err := bolt.Open(filepath.Join(dbPath, RECORDS_DB), 0755, &bolt.Options{})
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(RECORDS_TABLE))
		return err
	})
}

func AddRecord(dbPath []byte, start, end time.Time, integrationName, stage string) {
	st := start.Format(JS_TIME_LAYOUT)
	en := end.Format(JS_TIME_LAYOUT)

	db, err := bolt.Open(filepath.Join(string(dbPath), RECORDS_DB), 0755, &bolt.Options{})
	if err != nil {
		panic(fmt.Errorf("error opening db %s . %s", dbPath, err))
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(RECORDS_TABLE))
		id, _ := b.NextSequence()
		clusterID := itob(int(id))

		c2 := record{
			Startdate:       fmt.Sprintf("_%s-", st),
			Enddate:         fmt.Sprintf(";%s+", en),
			IntegrationName: integrationName,
			Stage:           stage,
		}

		return b.Put(clusterID, toBytes(c2))
	})

}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func GetJSArray(dbPath []byte) string {

	db, err := bolt.Open(filepath.Join(string(dbPath), RECORDS_DB), 0755, &bolt.Options{})
	if err != nil {
		panic(fmt.Errorf("error opening db %s in get cluster %w", dbPath, err))
	}
	defer db.Close()

	records := "["

	count := 0
	err = db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(RECORDS_TABLE))
		return b.ForEach(func(_, bts []byte) error {

			fmt.Println("check0")
			rec := record{}
			err := json.Unmarshal(bts, &rec)
			if err != nil {
				return fmt.Errorf("error on unmarshal on iteration0 %w", err)
			}

			fmt.Println("check1")
			firstStart := strings.Index(string(bts), "_") - 2
			endStart := bytes.IndexByte(bts, '-')
			bts = append(bts[:firstStart+1], append([]byte(fmt.Sprintf("new Date(\"%s\")", rec.Startdate[1:len(rec.Startdate)-1])), bts[endStart+2:]...)...)

			fmt.Println("check2")
			firstStart = strings.Index(string(bts), ";") - 2
			endStart = bytes.IndexByte(bts, '+')
			bts = append(bts[:firstStart+1], append([]byte(fmt.Sprintf("new Date(\"%s\")", rec.Enddate[1:len(rec.Enddate)-1])), bts[endStart+2:]...)...)

			fmt.Println("check3", records)
			records += string(bts) + ",\n"
			count++
			return nil
		})
	})
	if err != nil {
		panic(err)
	}

	if count != 0 {
		records = records[:len(records)-2]
	}

	records += "]"

	return records
}

func PrintOutTasksInOrder(dbPath []byte) {

	db, err := bolt.Open(filepath.Join(string(dbPath), RECORDS_DB), 0755, &bolt.Options{})

	if err != nil {
		panic(fmt.Errorf("error opening db %s in get cluster %w", dbPath, err))
	}

	defer db.Close()

	taskNames := make([]string, 0)
	err = db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(RECORDS_TABLE))
		return b.ForEach(func(_, bts []byte) error {
			rec := record{}
			err := json.Unmarshal(bts, &rec)

			if err != nil {
				return fmt.Errorf("error on unmarshal on iteration0 %w", err)
			}

			taskNames = append(taskNames, rec.IntegrationName)

			return nil
		})
	})

	if err != nil {
		panic(err)
	}

	bts, err := json.Marshal(taskNames)

	if err != nil {
		panic(err)
	}

	fmt.Println("TASK-NAMES", string(bts))

}
