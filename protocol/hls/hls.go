package hls

import (
	"bytes"
	"container/list"
	"fmt"
	"sync"
)

const (
	_maxTSCacheNum = 3
)

type TSItem struct {
	Name     string
	SeqNum   uint32
	Duration float64
	Data     []byte
}

type TSCache struct {
	num  int
	lock sync.RWMutex
	ll   *list.List
	lm   map[string]TSItem
}

func NewTSCache() *TSCache {
	return &TSCache{
		ll:  list.New(),
		num: _maxTSCacheNum,
		lm:  make(map[string]TSItem),
	}
}

func (tcCacheItem *TSCache) GetPlayList() ([]byte, error) {
	var seq uint32
	var getSeq bool
	var maxDuration float64
	m3u8body := bytes.NewBuffer(nil)
	for e := tcCacheItem.ll.Front(); e != nil; e = e.Next() {
		key := e.Value.(string)
		v, ok := tcCacheItem.lm[key]
		if ok {
			if v.Duration > maxDuration {
				maxDuration = v.Duration
			}
			if !getSeq {
				getSeq = true
				seq = v.SeqNum
			}
			_, _ = fmt.Fprintf(m3u8body, "#EXTINF:%.3f,\n%s\n", v.Duration, v.Name)
		}
	}
	w := bytes.NewBuffer(nil)
	_, _ = fmt.Fprintf(w,
		"#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-ALLOW-CACHE:NO\n#EXT-X-TARGETDURATION:%d\n#EXT-X-MEDIA-SEQUENCE:%d\n\n",
		int(maxDuration), seq)
	w.Write(m3u8body.Bytes())
	return w.Bytes(), nil
}

func (tcCacheItem *TSCache) SetItem(key string, seq uint32, duration float64, d []byte) {
	item := TSItem{
		Name:     key,
		SeqNum:   seq,
		Duration: duration,
		Data:     d,
	}
	if tcCacheItem.ll.Len() == tcCacheItem.num {
		e := tcCacheItem.ll.Front()
		tcCacheItem.ll.Remove(e)
		k := e.Value.(string)
		delete(tcCacheItem.lm, k)
	}
	tcCacheItem.lm[key] = item
	tcCacheItem.ll.PushBack(key)
}

func (tcCacheItem *TSCache) GetItem(key string) (TSItem, error) {
	item, ok := tcCacheItem.lm[key]
	if !ok {
		return item, fmt.Errorf("No key for cache")
	}
	return item, nil
}
