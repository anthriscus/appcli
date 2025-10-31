package store

import (
	"context"
)

// ?
// type StoreData interface {
// 	Read(key int64) TodoListRecord
// 	Write(record TodoListRecord)
// 	Keys() []int64
// }

type rdData struct {
	key        int64
	returnChan *chan TodoListRecord
}
type rdKeysData struct {
	returnChan *chan []int64
}

type TodoListRecord struct {
	// index int64
	item TodoListItem
	ok   bool
}

type StoreChannels struct {
	writeChan    chan TodoListRecord
	readChan     chan rdData
	readKeysChan chan rdKeysData
}

// exploring
// channels way of enabling read, write aand iterator over todolist map during multi go routines
// There is a case for Mutexes but here we are learnng about channels
func NewStoreChannels(ctx context.Context) *StoreChannels {
	chans := StoreChannels{
		writeChan:    make(chan TodoListRecord),
		readChan:     make(chan rdData),
		readKeysChan: make(chan rdKeysData),
	}

	// fetch the number keys from the map
	var getKeys = func(data TodoListItems) []int64 {
		keys := make([]int64, 0, len(data))
		for i := range data {
			keys = append(keys, i)
		}
		return keys
	}

	// actor
	go func() {
		for {
			select {
			// end
			case <-ctx.Done():
				return
			// read record
			case rdData := <-chans.readChan:
				item, ok := sessionDatabase[rdData.key]
				*rdData.returnChan <- TodoListRecord{
					item: item, ok: ok,
				}
				close(*rdData.returnChan)
			// write record
			case rec := <-chans.writeChan:
				sessionDatabase[rec.item.Line] = rec.item
			// get keys. needed a safe iterator over keys during writes on other routines
			case rdKData := <-chans.readKeysChan:
				*rdKData.returnChan <- getKeys(sessionDatabase)
				close(*rdKData.returnChan)
			}
		}
	}()
	return &chans
}

func (c *StoreChannels) Read(key int64) TodoListRecord {
	resultsChan := make(chan TodoListRecord)
	c.readChan <- rdData{key: key, returnChan: &resultsChan}
	return <-resultsChan
}

func (c *StoreChannels) Write(record TodoListRecord) {
	c.writeChan <- record
}

func (c *StoreChannels) Keys() []int64 {
	resultsChan := make(chan []int64)
	c.readKeysChan <- rdKeysData{returnChan: &resultsChan}
	return <-resultsChan
}
