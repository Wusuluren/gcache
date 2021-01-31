# gcache
Gcahe is a pre-generated golang cache library with solid types.

## Examples
First generate cache code with your types or default.
```go
go run cmd/gen/main.go
```
Code generated will like following:
```go
package gcache

import (
	"sort"
	"sync"
	"time"
)

type valueItemSort struct {
	count int
	key   int
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
	Value    int
}

type cacheMap struct {
	data   map[int]valueItem
	access map[int]int
	lock   sync.Mutex
}

func (c *cacheMap) Set(key int, val int, duration time.Duration) {
	c.lock.Lock()
	c.data[key] = valueItem{
		Value:    val,
		ExpireAt: time.Now().Add(duration),
	}
	c.access[key]=1
	c.lock.Unlock()
}

func (c *cacheMap) Get(key int) (val int, ok bool) {
	if item, ok := c.data[key]; ok {
		c.access[key]++
		return item.Value, ok
	}
	return
}

func (c *cacheMap) Del(key int) {
	c.lock.Lock()
	delete(c.data, key)
	c.lock.Unlock()
}

func (c *cacheMap) MDel(key ...int) {
	c.lock.Lock()
	for _, k := range key {
		delete(c.data, k)
	}
	c.lock.Unlock()
}

type CacheIntInt struct {
	slotData           []cacheMap
	hashFn             func(v int) int64
	cleanupInterval    time.Duration
	maxSlotSize        int
	reduceSlotSizeRate float64
}

func NewCacheIntInt() CacheIntInt {
	slotData := make([]cacheMap, defaultSlotNum)
	for i := range slotData {
		slotData[i] = cacheMap{
			data: make(map[int]valueItem),
			lock: sync.Mutex{},
		}
	}
	c := CacheIntInt{
		slotData:           slotData,
		hashFn:             hashInt,
		cleanupInterval:    10 * time.Minute,
		maxSlotSize:        1024*1024,
		reduceSlotSizeRate: 0.75,
	}
	go c.cleanup()
	return c
}

func (c *CacheIntInt) cleanup() {
	t := time.NewTicker(c.cleanupInterval)
	defer t.Stop()
	for now := range t.C {
		for i := range c.slotData {
			expiredKeys := make([]int, 0, 1024)
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

func (c *CacheIntInt) Set(key int, val int, duration time.Duration) {
	h := c.hashFn(key)
	c.slotData[h].Set(key, val, duration)
}

func (c *CacheIntInt) Get(key int) (val int, ok bool) {
	h := c.hashFn(key)
	val, ok = c.slotData[h].Get(key)
	return
}

func (c *CacheIntInt) Del(key int) {
	h := c.hashFn(key)
	c.slotData[h].Del(key)
}

func (c *CacheIntInt) MDel(key ...int) {
	slotKeyMap := map[int64][]int{}
	for _, k := range key {
		h := c.hashFn(k)
		if _, ok := slotKeyMap[h]; ok {
			slotKeyMap[h] = append(slotKeyMap[h], k)
		} else {
			slotKeyMap[h] = []int{k}
		}
	}
	for h, keys := range slotKeyMap {
		c.slotData[h].MDel(keys...)
	}
}
```
Write tests like following:
```go
func TestNewCache(t *testing.T) {
	c := NewCacheIntInt()
	c.Set(1, 1, time.Second)
	t.Log(c.Get(1))
	c.Del(1)
	t.Log(c.Get(1))
}
```