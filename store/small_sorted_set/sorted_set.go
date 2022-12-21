package small_sorted_set

import (
	"golang.org/x/exp/slices"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/store/store_db/store_db_interface"
)

type SmallSortedSet struct {
	Data     []*SmallSortedSetNode
	Dict     map[string]*SmallSortedSetNode
	Tx       store_db_interface.StoreDBTransactionInterface
	maxCount int
	prefix   string
	changed  bool
}

func (this *SmallSortedSet) Read() (err error) {

	this.Dict = map[string]*SmallSortedSetNode{}
	this.Data = []*SmallSortedSetNode{}
	this.changed = false

	data := this.Tx.Get(this.prefix + ":count")
	if data == nil {
		return
	}

	var count uint64
	if count, err = advanced_buffers.NewBufferReader(data).ReadUvarint(); err != nil {
		return
	}

	data = this.Tx.Get(this.prefix + ":data")
	b := advanced_buffers.NewBufferReader(data)

	this.Data = make([]*SmallSortedSetNode, count)
	for i := range this.Data {
		this.Data[i] = &SmallSortedSetNode{}
		if err = this.Data[i].Deserialize(b); err != nil {
			return
		}
		this.Dict[this.Data[i].Key] = this.Data[i]
	}

	return
}

//func GetByRank(prefix string, start, count int, tx transan) (list []*api_types.APIMethodFindListItem, err error) {
//
//	prefix = "smallSortedSet:" + prefix
//
//	data := tx.Get(prefix + ":count")
//	if data == nil {
//		return
//	}
//
//	var count uint64
//	if count, err = advanced_buffers.NewBufferReader(data).ReadUvarint(); err != nil {
//		return
//	}
//
//	for i := start; i < count && len(list) < config.CHAT_MESSAGES_LIST_COUNT; i++ {
//		result := ss.Data[i]
//		list = append(list, &api_types.APIMethodFindListItem{
//			result.Key,
//			result.Score,
//		})
//	}
//}

func (this *SmallSortedSet) Save() {
	if !this.changed {
		return
	}

	w := advanced_buffers.NewBufferWriter()
	w.WriteUvarint(uint64(len(this.Data)))
	this.Tx.Put(this.prefix+":count", w.Bytes())

	w = advanced_buffers.NewBufferWriter()
	for _, d := range this.Data {
		d.Serialize(w)
	}
	this.Tx.Put(this.prefix+":data", w.Bytes())
}

func (this *SmallSortedSet) Add(key string, score float64) {

	if found := this.Dict[string(key)]; found != nil {
		found.Score = score
		slices.SortFunc(this.Data, func(a, b *SmallSortedSetNode) bool {
			return a.Score > b.Score
		})
		this.changed = true
		return
	}

	x := &SmallSortedSetNode{
		key,
		score,
	}

	this.Data = append(this.Data, x)
	this.Dict[key] = x

	slices.SortFunc(this.Data, func(a, b *SmallSortedSetNode) bool {
		return a.Score > b.Score
	})

	if len(this.Data) > this.maxCount {
		last := this.Data[len(this.Data)-1]
		delete(this.Dict, last.Key)
		this.Data = this.Data[:len(this.Data)-1]
	}

	this.changed = true
}

func (this *SmallSortedSet) Delete(key string) bool {

	var found *SmallSortedSetNode
	if found = this.Dict[key]; found == nil {
		return false
	}

	for i := range this.Data {
		if this.Data[i] == found {
			for r := i; r < len(this.Data)-1; r++ {
				this.Data[r] = this.Data[r+1]
			}
			this.Data = this.Data[:len(this.Data)-1]
			delete(this.Dict, key)
			break
		}
	}

	slices.SortFunc(this.Data, func(a, b *SmallSortedSetNode) bool {
		return a.Score > b.Score
	})
	this.changed = true

	return true
}

func NewSmallSortedSet(maxCount int, prefix string, tx store_db_interface.StoreDBTransactionInterface) *SmallSortedSet {

	ss := &SmallSortedSet{
		make([]*SmallSortedSetNode, 0),
		make(map[string]*SmallSortedSetNode),
		tx,
		maxCount,
		"smallSortedSet:" + prefix,
		false,
	}

	return ss
}
