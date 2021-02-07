package storage

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/xujiajun/nutsdb"
)

const HistoryKeyFormat = "20060102150405.00000"

type RequestResponse struct {
	Request  RequestInput
	Response RequestResult
}

// RequestInput holds the request information
type RequestInput struct {
	Body    string
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

type HistoryList []HistoryEntry

type HistoryEntry struct {
	Key string
	RR  RequestResponse
}

type HistoryStorage struct {
	db           *nutsdb.DB
	activeRecord *RequestResponse
}

const bucketNameHistory = "history"

func SetupHistory(db *nutsdb.DB) HistoryStorage {
	return HistoryStorage{
		db,
		&RequestResponse{},
	}
}

func (h *HistoryStorage) SetActiveRecord(rr *RequestResponse) {
	h.activeRecord = rr
}

func (h *HistoryStorage) GetActiveRecord() *RequestResponse {
	return h.activeRecord
}

func (h *HistoryStorage) GetAllRequests() HistoryList {
	var hl HistoryList
	if err := h.db.View(
		func(tx *nutsdb.Tx) error {
			entries, err := tx.GetAll(bucketNameHistory)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				var rqrs RequestResponse
				err = json.Unmarshal(entry.Value, &rqrs)
				if err != nil {
					return err
				}
				hl = append(hl, HistoryEntry{
					string(entry.Key),
					rqrs,
				})
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

func (h *HistoryStorage) RequestCompleted(key []byte, reqRes RequestResponse) {
	val, err := json.Marshal(reqRes)
	if err != nil {
		log.Fatal("Error marshaling reqres data: ", err)
	}
	if err := h.db.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Put(bucketNameHistory, key, val, 0); err != nil {
				return err
			}
			return nil
		}); err != nil {
		log.Fatal(err)
	}
}

func (h *HistoryStorage) RemoveEntry(key string) {
	if err := h.db.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Delete(bucketNameHistory, []byte(key)); err != nil {
				return err
			}
			return nil
		}); err != nil {
		log.Fatal(err)
	}
}

func (h *HistoryStorage) RemoveAll() {
	list := h.GetAllRequests()
	if err := h.db.Update(
		func(tx *nutsdb.Tx) error {
			for _, entry := range list {
				if err := tx.Delete(bucketNameHistory, []byte(entry.Key)); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
		log.Fatal(err)
	}
}

func (h *HistoryStorage) GetEntry(key string) HistoryEntry {
	var he HistoryEntry
	if err := h.db.View(
		func(tx *nutsdb.Tx) error {
			entry, err := tx.Get(bucketNameHistory, []byte(key))
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
