package storage

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/xujiajun/nutsdb"
)

type Setting struct {
	Key   string
	Value interface{}
}

type Settings map[string]interface{}

type SettingsStorage struct {
	db *nutsdb.DB
}

const bucketNameSettings = "settings"

const SettingCheckUpdates = "checkUpdates"
const SettingTheme = "theme"

func SetupSettings(db *nutsdb.DB) SettingsStorage {
	return SettingsStorage{
		db,
	}
}

func (h *SettingsStorage) GetAll() Settings {
	setList := make(Settings)
	if err := h.db.View(
		func(tx *nutsdb.Tx) error {
			entries, err := tx.GetAll(bucketNameSettings)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				var st Setting
				err = json.Unmarshal(entry.Value, &st.Value)
				if err != nil {
					return err
				}
				setList[string(entry.Key)] = st.Value
			}

			return nil
		}); err != nil {
		if err == nutsdb.ErrBucketEmpty {
			return setList
		} else {
			log.Fatal(err)
		}
	}
	return setList
}

func (h *SettingsStorage) UpdateSetting(key string, value interface{}) {
	valByte, err := json.Marshal(value)
	if err != nil {
		log.Fatal("Error marshaling setting data: ", err)
	}
	if err := h.db.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Put(bucketNameSettings, []byte(key), valByte, 0); err != nil {
				return err
			}
			return nil
		}); err != nil {
		log.Fatal(err)
	}
}

func (h *SettingsStorage) RemoveSetting(key string) {
	if err := h.db.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Delete(bucketNameSettings, []byte(key)); err != nil {
				return err
			}
			return nil
		}); err != nil {
		log.Fatal(err)
	}
}

func (h *SettingsStorage) GetSetting(key string) Setting {
	var st Setting
	if err := h.db.View(
		func(tx *nutsdb.Tx) error {
			entry, err := tx.Get(bucketNameSettings, []byte(key))
			if err != nil {
				return err
			}
			err = json.Unmarshal(entry.Value, &st.Value)
			if err != nil {
				return err
			}
			st.Key = string(entry.Key)
			return nil
		}); err != nil {
		log.Fatal(err)
	}
	return st
}
