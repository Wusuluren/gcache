package gcache

import (
	"sort"
	"sync"
	"time"
)

type valueItemSort struct {
	count int
	key   KeyType
}

type valueItemsSort []valueItemSort

func (l valueItemsSort) Len() int {
	return len(l)
}
func (l valueItemsSort) Less(i, j int) bool {
	return l[i].count < l[i].count
}
func (l valueItemsSort) Swap(i, j int) {
	l[i], l[i] = l[j], l[i]
}

type valueItem struct {
	ExpireAt time.Time
	Value    ValueType
}

type cacheMap struct {
	data   map[KeyType]valueItem
	access map[KeyType]int
	lock   sync.Mutex
}

func (c *cacheMap) Set(key KeyType, val ValueType, duration time.Duration) {
	c.lock.Lock()
	c.data[key] = valueItem{
		Value:    val,
		ExpireAt: time.Now().Add(duration),
	}
	c.access[key]=1
	c.lock.Unlock()
}

func (c *cacheMap) Get(key KeyType) (val ValueType, ok bool) {
	if item, ok := c.data[key]; ok {
		c.access[key]++
		return item.Value, ok
	}
	return
}

func (c *cacheMap) Del(key KeyType) {
	c.lock.Lock()
	delete(c.data, key)
	c.lock.Unlock()
}

func (c *cacheMap) MDel(key ...KeyType) {
	c.lock.Lock()
	for _, k := range key {
		delete(c.data, k)
	}
	c.lock.Unlock()
}

type Cache struct {
	slotData           []cacheMap
	cleanupInterval    time.Duration
	maxSlotSize        int
	reduceSlotSizeRate float64
}

func NewCache() Cache {
	slotData := make([]cacheMap, defaultSlotNum)
	for i := range slotData {
		slotData[i] = cacheMap{
			data:   make(map[KeyType]valueItem),
			access: make(map[int]int),
			lock:   sync.Mutex{},
		}
	}
	c := Cache{
		slotData:           slotData,
		cleanupInterval:    _cleanupInterval,
		maxSlotSize:        _maxSlotSize,
		reduceSlotSizeRate: _reduceSlotSizeRate,
	}
	go c.cleanup()
	return c
}

func (c *Cache) cleanup() {
	t := time.NewTicker(c.cleanupInterval)
	defer t.Stop()
	for now := range t.C {
		for i := range c.slotData {
			expiredKeys := make([]KeyType, 0, 1024)
			dropItems := make([]valueItemSort, 0, 1024)
			for key, item := range c.slotData[i].data {
				if item.ExpireAt.Before(now) {
					expiredKeys = append(expiredKeys, key)
				} else {
					dropItems = append(dropItems, valueItemSort{
						count: c.slotData[i].access[key],
						key:   key,
					})
				}
			}
			c.MDel(expiredKeys...)

			if len(c.slotData[i].data) > c.maxSlotSize {
				expiredKeys = expiredKeys[0:0]
				sort.Reverse(valueItemsSort(dropItems))
				leftTotal := int(float64(len(dropItems)) * c.reduceSlotSizeRate)
				for i := range dropItems[leftTotal:] {
					expiredKeys = append(expiredKeys, dropItems[i].key)
				}
				c.MDel(expiredKeys...)
			}
		}
	}
}

func (c *Cache) Set(key KeyType, val ValueType, duration time.Duration) {
	h := _hashFn(key)
	c.slotData[h].Set(key, val, duration)
}

func (c *Cache) Get(key KeyType) (val ValueType, ok bool) {
	h := _hashFn(key)
	val, ok = c.slotData[h].Get(key)
	return
}

func (c *Cache) Del(key KeyType) {
	h := _hashFn(key)
	c.slotData[h].Del(key)
}

func (c *Cache) MDel(key ...KeyType) {
	slotKeyMap := map[int][]KeyType{}
	for _, k := range key {
		h := _hashFn(k)
		if _, ok := slotKeyMap[h]; ok {
			slotKeyMap[h] = append(slotKeyMap[h], k)
		} else {
			slotKeyMap[h] = []KeyType{k}
		}
	}
	for h, keys := range slotKeyMap {
		c.slotData[h].MDel(keys...)
	}
}
