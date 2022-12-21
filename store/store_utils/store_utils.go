package store_utils

import (
	"liberty-town/node/store/store_db/store_db_interface"
	"math/rand"
	"strconv"
)

func GetBetterScore(table, key string, tx store_db_interface.StoreDBTransactionInterface) (uint64, error) {
	if data := tx.Get(table + ":by_key:better_score:" + key); data != nil {
		score, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return 0, err
		}
		return score, nil
	}
	return 0, nil
}

func IncreaseCount(table, key string, betterScore uint64, tx store_db_interface.StoreDBTransactionInterface) (err error) {

	tx.Put(table+":by_key:better_score:"+key, []byte(strconv.FormatUint(betterScore, 10)))

	if data := tx.Get(table + "_indexes:" + key); len(data) > 0 {
		return nil
	}

	var count uint64
	if data := tx.Get(table + ":count"); data != nil {
		if count, err = strconv.ParseUint(string(data), 10, 64); err != nil {
			return
		}
	}
	tx.Put(table+":by_index:"+strconv.FormatUint(count, 10), []byte(key))
	tx.Put(table+"_indexes:"+key, []byte(strconv.FormatUint(count, 10)))
	count++
	tx.Put(table+":count", []byte(strconv.FormatUint(count, 10)))
	return
}

func DecreaseCount(table string, key string, tx store_db_interface.StoreDBTransactionInterface) (err error) {

	if data := tx.Get(table + "_indexes:" + key); len(data) == 0 {
		return nil
	}

	var index, count uint64
	if count, err = strconv.ParseUint(string(tx.Get(table+":count")), 10, 64); err != nil {
		return
	}
	if index, err = strconv.ParseUint(string(tx.Get(table+"_indexes:"+key)), 10, 64); err != nil {
		return
	}

	tx.Put(table+":count", []byte(strconv.FormatUint(count-1, 10)))

	if count > 1 {
		last := tx.Get(table + ":by_index:" + strconv.FormatUint(count-1, 10))
		tx.Put(table+":by_index:"+strconv.FormatUint(index, 10), last)
		tx.Put(table+"_indexes:"+string(last), []byte(strconv.FormatUint(index, 10)))
	}

	tx.Delete(table + ":by_index:" + strconv.FormatUint(count-1, 10))
	tx.Delete(table + ":by_key:better_score:" + key)
	tx.Delete(table + "_indexes:" + key)
	return
}

// 选择随机元素
func GetRandomItems(table string, tx store_db_interface.StoreDBTransactionInterface, length uint64) (keys []string, betterScores []uint64, err error) {

	var count uint64
	if data := tx.Get(table + ":count"); data != nil {
		if count, err = strconv.ParseUint(string(data), 10, 64); err != nil {
			return
		}
	}

	if count == 0 {
		return nil, nil, nil
	}

	start := rand.Uint64() % count
	if start >= length {
		start -= length
	} else {
		start = 0
	}

	for i := start; i < count && uint64(len(keys)) < length; i++ {
		key := tx.Get(table + ":by_index:" + strconv.FormatUint(i, 10))
		keys = append(keys, string(key))
		score, err := GetBetterScore(table, string(key), tx)
		if err != nil {
			return nil, nil, err
		}
		betterScores = append(betterScores, score)
	}

	return
}
