package storage

import (
	"encoding/json"
	"log"
	"time"

	"github.com/xujiajun/nutsdb"
)

const HistoryKeyFormat = "20060102150405.00000"

type RequestResponse struct {
	Request  RequestInput
	Response RequestResult
}

// RequestInput holds the request information
type RequestInput struct {
	Method  string
	Path    string
	Headers map[string][]string
}

// RequestResult holds response information
type RequestResult struct {
	StatusCode   int
	Headers      map[string][]string
	ResponseBody []byte
	Dur          time.Duration
}

type HistoryList map[string]RequestResponse

type HistoryEntry struct {
	Key string
	RR  RequestResponse
}

type History struct {
	db *nutsdb.DB
}

const bucketName = "history"

func Setup(db *nutsdb.DB) History {
	return History{
		db,
	}
}

func (h *History) GetAllRequests() HistoryList {
	hl := make(HistoryList)
	if err := h.db.View(
		func(tx *nutsdb.Tx) error {
			entries, err := tx.GetAll(bucketName)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				var rqrs RequestResponse
				err = json.Unmarshal(entry.Value, &rqrs)
				if err != nil {
					return err
				}
				hl[string(entry.Key)] = rqrs
			}

			return nil
		}); err != nil {
		if err == nutsdb.ErrBucketEmpty {
			return hl
		} else {
			log.Fatal(err)
		}
	}
	return hl
}

func (h *History) RequestCompleted(key []byte, reqRes RequestResponse) {
	val, err := json.Marshal(reqRes)
	if err != nil {
		log.Fatal("Error marshaling reqres data: ", err)
	}
	if err := h.db.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Put(bucketName, key, val, 0); err != nil {
				return err
			}
			return nil
		}); err != nil {
		log.Fatal(err)
	}
}

func (h *History) RemoveEntry(key string) {
	if err := h.db.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Delete(bucketName, []byte(key)); err != nil {
				return err
			}
			return nil
		}); err != nil {
		log.Fatal(err)
	}
}

func (h *History) GetEntry(key string) HistoryEntry {
	var he HistoryEntry
	if err := h.db.View(
		func(tx *nutsdb.Tx) error {
			entry, err := tx.Get(bucketName, []byte(key))
			if err != nil {
				return err
			}
			var rqrs RequestResponse
			err = json.Unmarshal(entry.Value, &rqrs)
			if err != nil {
				return err
			}
			he.Key = string(entry.Key)
			he.RR = rqrs
			return nil
		}); err != nil {
		log.Fatal(err)
	}
	return he
}
